package mongo

import (
	"math"
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const AvatarBase = "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/"

type Group struct {
	ID64          int64     `bson:"_id"`
	ID            int       `bson:"id"`
	CreatedAt     time.Time `bson:"created_at"`
	UpdatedAt     time.Time `bson:"updated_at"`
	Name          string    `bson:"name"`
	URL           string    `bson:"url"`
	Headline      string    `bson:"headline"`
	Summary       string    `bson:"summary"`
	Icon          string    `bson:"icon"`
	Members       int       `bson:"members"`
	MembersInChat int       `bson:"members_in_chat"`
	MembersInGame int       `bson:"members_in_game"`
	MembersOnline int       `bson:"members_online"`
}

func (group Group) BSON() (ret interface{}) {

	if group.CreatedAt.IsZero() {
		group.CreatedAt = time.Now()
	}

	return M{
		"_id":             group.ID64,
		"id":              group.ID,
		"created_at":      group.CreatedAt,
		"updated_at":      time.Now(),
		"name":            group.Name,
		"url":             group.URL,
		"headline":        group.Headline,
		"summary":         group.Summary,
		"icon":            group.Icon,
		"members":         group.Members,
		"members_in_chat": group.MembersInChat,
		"members_in_game": group.MembersInGame,
		"members_online":  group.MembersOnline,
	}
}

func (group Group) OutputForJSON() (output []interface{}) {

	return []interface{}{
		group.ID64,
		group.Name,
		group.GetPath(),
		group.GetIcon(),
		group.Headline,
		group.Members,
		group.URL,
	}
}

func (group Group) GetPath() string {
	return "/groups/" + strconv.FormatInt(group.ID64, 10) + "/" + slug.Make(group.Name)
}

func (group Group) GetName() string {
	return group.Name
}

func (group Group) GetIcon() string {
	return AvatarBase + group.Icon
}

func (group Group) Save() error {

	_, err := ReplaceDocument(CollectionGroups, M{"_id": group.ID64}, group)
	return err
}

func GetGroup(id int64) (group Group, err error) {

	// if !IsValidPlayerID(id) {
	// 	return group, ErrInvalidPlayerID
	// }

	err = FindDocument(CollectionGroups, "_id", id, nil, &group)

	if id > math.MaxInt32 {
		group.ID64 = id
	} else {
		group.ID = int(id)
	}

	return group, err
}

func GetGroupsByID(ids []int64, projection M) (groups []Group, err error) {

	if len(ids) < 1 {
		return groups, nil
	}

	var idsBSON A
	for _, v := range ids {
		idsBSON = append(idsBSON, v)
	}

	return getGroups(0, 0, D{{"name", 1}}, M{"_id": M{"$in": idsBSON}}, projection)
}

func GetGroups(offset int64, sort D, filter M, projection M) (groups []Group, err error) {

	return getGroups(offset, 100, sort, filter, projection)
}

func getGroups(offset int64, limit int64, sort D, filter interface{}, projection M) (groups []Group, err error) {

	if filter == nil {
		filter = M{}
	}

	client, ctx, err := getMongo()
	if err != nil {
		return groups, err
	}

	ops := options.Find()
	if offset > 0 {
		ops.SetSkip(offset)
	}
	if limit > 0 {
		ops.SetLimit(limit)
	}
	if sort != nil {
		ops.SetSort(sort)
	}

	if projection != nil {
		ops.SetProjection(projection)
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionGroups.String())
	cur, err := c.Find(ctx, filter, ops)
	if err != nil {
		return groups, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var group Group
		err := cur.Decode(&group)
		if err != nil {
			log.Err(err, group.ID)
		}
		groups = append(groups, group)
	}

	return groups, cur.Err()
}
