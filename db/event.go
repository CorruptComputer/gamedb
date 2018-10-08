package db

import (
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/memcache"
)

const (
	EventLogin   = "login"
	EventRefresh = "refresh"
)

type Event struct {
	CreatedAt time.Time `datastore:"created_at"`
	Type      string    `datastore:"type"`
	PlayerID  int64     `datastore:"player_id"`
	UserAgent string    `datastore:"user_agent,noindex"`
	IP        string    `datastore:"ip,noindex"`
}

func (event Event) GetKey() (key *datastore.Key) {
	return datastore.IncompleteKey(KindEvent, nil)
}

func (event Event) GetCreatedNice() (t string) {
	return event.CreatedAt.Format(helpers.DateTime)
}

func (event Event) GetUserAgentShort() (t string) {

	if len(event.UserAgent) > 50 {
		event.UserAgent = event.UserAgent[0:50] + "&hellip;"
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

func (event Event) GetCreatedUnix() int64 {
	return event.CreatedAt.Unix()
}

func (event Event) GetType() string {

	switch event.Type {
	case EventLogin:
		return "User Login"
	case EventRefresh:
		return "Profile Update"
	default:
		return strings.Title(event.Type)
	}
}

// Data array for datatables
func (event Event) OutputForJSON(r *http.Request) (output []interface{}) {

	return []interface{}{
		event.CreatedAt.Unix(),
		event.GetCreatedNice(),
		event.GetType(),
		event.GetIP(""),
		event.UserAgent,
		event.GetUserAgentShort(),
		event.GetIP(r.Header.Get("X-Forwarded-For")),
	}
}

func CountPlayerEvents(playerID int64) (count int, err error) {

	return memcache.GetSetInt(memcache.PlayerEventsCount(playerID), &count, func() (count int, err error) {

		client, ctx, err := GetDSClient()
		if err != nil {
			return count, err
		}

		q := datastore.NewQuery(KindEvent).Filter("player_id = ", playerID).Limit(10000)
		count, err = client.Count(ctx, q)
		return count, err
	})
}

func CreateEvent(r *http.Request, playerID int64, eventType string) (err error) {

	login := new(Event)
	login.CreatedAt = time.Now()
	login.PlayerID = playerID
	login.Type = eventType
	login.UserAgent = r.Header.Get("User-Agent")
	login.IP = r.Header.Get("X-Forwarded-For")

	_, err = SaveKind(login.GetKey(), login)
	return err
}
