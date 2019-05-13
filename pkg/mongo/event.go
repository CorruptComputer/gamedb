package mongo

import (
	"net/http"
	"strings"
	"time"

	"github.com/Jleagle/memcache-go/memcache"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	EventLogin   = "login"
	EventLogout  = "logout"
	EventRefresh = "refresh"
	EventPatreon = "patreon"
)

type Event struct {
	CreatedAt time.Time `bson:"created_at"`
	Type      string    `bson:"type"`
	SteamID   int64     `bson:"player_id"`
	UserAgent string    `bson:"user_agent"`
	IP        string    `bson:"ip"`
}

func (event Event) BSON() (ret interface{}) {

	return M{
		"created_at": event.CreatedAt,
		"type":       event.Type,
		"player_id":  event.SteamID,
		"user_agent": event.UserAgent,
		"ip":         event.IP,
	}
}

// Data array for datatables
func (event Event) OutputForJSON(ip string) (output []interface{}) {

	return []interface{}{
		event.CreatedAt.Unix(),
		event.GetCreatedNice(),
		event.GetType(),
		event.GetIP(""),
		event.UserAgent,
		event.GetUserAgentShort(),
		event.GetIP(ip),
		event.GetIcon(),
	}
}

func (event Event) GetCreatedNice() (t string) {
	return event.CreatedAt.Format(helpers.DateTime)
}

func (event Event) GetUserAgentShort() (t string) {

	if len(event.UserAgent) > 50 {
		return event.UserAgent[0:50] + "&hellip;"
	}

	return event.UserAgent
}

// Defaults to IP on struct
func (event Event) GetIP(ip string) string {

	if ip == "" {
		ip = event.IP
	}

	var ips = strings.Split(ip, ", ")
	if len(ips) > 0 && ips[0] != "" {
		return ips[0]
	}
	return "-"
}

func (event Event) GetType() string {

	switch event.Type {
	case EventLogin:
		return "User Login"
	case EventLogout:
		return "User Logout"
	case EventRefresh:
		return "Profile Update"
	default:
		return strings.Title(event.Type)
	}
}

func (event Event) GetIcon() string {

	switch event.Type {
	case EventLogin:
		return "fa-sign-in-alt"
	case EventLogout:
		return "fa-sign-out-alt"
	case EventRefresh:
		return "fa-sync-alt"
	default:
		return "fa-star"
	}
}

func GetEvents(playerID int64, offset int64) (events []Event, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return events, err
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionEvents.String())

	cur, err := c.Find(ctx, M{"player_id": playerID}, options.Find().SetLimit(100).SetSkip(offset).SetSort(M{"created_at": -1}))
	if err != nil {
		return events, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var event Event
		err := cur.Decode(&event)
		log.Err(err)
		events = append(events, event)
	}

	return events, cur.Err()
}

func CreateEvent(r *http.Request, steamID int64, eventType string) (err error) {

	event := &Event{}
	event.CreatedAt = time.Now()
	event.SteamID = steamID
	event.Type = eventType
	event.UserAgent = r.Header.Get("User-Agent")
	event.IP = r.RemoteAddr

	_, err = InsertDocument(CollectionEvents, event)
	if err != nil {
		return err
	}

	if config.HasMemcache() {
		err = helpers.GetMemcache().Delete(helpers.MemcachePlayerEventsCount(steamID).Key)
		err = helpers.IgnoreErrors(err, memcache.ErrCacheMiss)
	}

	return err
}

func CountEvents(playerID int64) (count int64, err error) {

	var item = helpers.MemcachePlayerEventsCount(playerID)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		return CountDocuments(CollectionEvents, M{"player_id": playerID})
	})

	return count, err
}
