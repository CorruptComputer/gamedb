package mongo

import (
	"errors"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrInvalidGroupID = errors.New("invalid group id")

type Group struct {
	ID            string    `bson:"_id"` // Too big for int64 in Javascript (Mongo BD)
	CreatedAt     time.Time `bson:"created_at"`
	UpdatedAt     time.Time `bson:"updated_at"`
	Name          string    `bson:"name"`
	Abbr          string    `bson:"abbreviation"`
	URL           string    `bson:"url"`
	AppID         int       `bson:"app_id"`
	Headline      string    `bson:"headline"`
	Summary       string    `bson:"summary"`
	Icon          string    `bson:"icon"`
	Trending      int64     `bson:"trending"`
	Members       int       `bson:"members"`
	MembersInChat int       `bson:"members_in_chat"`
	MembersInGame int       `bson:"members_in_game"`
	MembersOnline int       `bson:"members_online"`
	Error         string    `bson:"error"`
	Type          string    `bson:"type"`
}

func (group Group) BSON() bson.D {

	if group.CreatedAt.IsZero() || group.CreatedAt.Unix() == 0 {
		group.CreatedAt = time.Now()
	}

	group.UpdatedAt = time.Now()

	return bson.D{
		{"_id", group.ID},
		{"created_at", group.CreatedAt},
		{"updated_at", group.UpdatedAt},
		{"name", group.Name},
		{"abbreviation", group.Abbr},
		{"url", group.URL},
		{"app_id", group.AppID},
		{"headline", group.Headline},
		{"summary", group.Summary},
		{"icon", group.Icon},
		{"trending", group.Trending},
		{"members", group.Members},
		{"members_in_chat", group.MembersInChat},
		{"members_in_game", group.MembersInGame},
		{"members_online", group.MembersOnline},
		{"error", group.Error},
		{"type", group.Type},
	}
}

func CreateGroupIndexes() {

	var indexModels = []mongo.IndexModel{
		{
			Keys:    bson.D{{"name", "text"}, {"abbreviation", "text"}, {"url", "text"}},
			Options: options.Index().SetName("text").SetWeights(bson.D{{"name", 3}, {"abbreviation", 2}, {"url", 1}}),
		},
		{
			Keys: bson.D{{"type", 1}, {"members", -1}},
		},
		{
			Keys: bson.D{{"type", 1}, {"trending", 1}},
		},
		{
			Keys: bson.D{{"type", 1}, {"trending", -1}},
		},
	}

	//
	client, ctx, err := getMongo()
	if err != nil {
		log.Err(err)
		return
	}

	_, err = client.Database(MongoDatabase).Collection(CollectionGroups.String()).Indexes().CreateMany(ctx, indexModels)
	log.Err(err)
}

func (group Group) GetPath() string {
	return helpers.GetGroupPath(group.ID, group.Name)
}

func (group Group) GetType() string {
	return helpers.GetGroupType(group.Type)
}

func (group Group) IsOfficial() bool {
	return helpers.IsGroupOfficial(group.Type)
}

func (group Group) GetURL() string {
	return helpers.GetGroupLink(group.Type, group.URL)
}

func (group Group) GetName() string {
	return helpers.GetGroupName(group.Name, group.ID)
}

func (group Group) GetIcon() string {
	return helpers.GetGroupIcon(group.Icon)
}

func (group Group) ShouldUpdate() bool {
	return group.UpdatedAt.Before(time.Now().Add(time.Hour * -6))
}

func GetGroup(id string) (group Group, err error) {

	id, err = helpers.IsValidGroupID(id)
	if err != nil {
		return group, ErrInvalidGroupID
	}

	var item = memcache.MemcacheGroup(id)

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &group, func() (interface{}, error) {

		err = FindOne(CollectionGroups, bson.D{{"_id", id}}, nil, nil, &group)
		if err != nil {
			return group, err
		}
		if group.ID == "" {
			return group, ErrNoDocuments
		}

		return group, err
	})

	return group, err
}

func GetGroupsByID(ids []string, projection bson.M) (groups []Group, err error) {

	if len(ids) == 0 {
		return groups, nil
	}

	chunks := helpers.ChunkStrings(ids, 100)

	for _, chunk := range chunks {

		var idsBSON bson.A

		for _, groupID := range chunk {

			groupID, err = helpers.IsValidGroupID(groupID)
			if err != nil {
				log.Err(err)
				continue
			}
			idsBSON = append(idsBSON, groupID)
		}

		resp, err := getGroups(0, 0, nil, bson.D{{"_id", bson.M{"$in": idsBSON}}}, projection)
		if err != nil {
			return groups, err
		}

		groups = append(groups, resp...)
	}

	return groups, err
}

func GetGroups(limit int64, offset int64, sort bson.D, filter bson.D, projection bson.M) (groups []Group, err error) {

	return getGroups(offset, limit, sort, filter, projection)
}

func SearchGroups(s string) (group Group, err error) {

	filter := bson.D{}

	s2, err := helpers.IsValidGroupID(s)
	if err == nil {

		filter = bson.D{{"_id", s2}}

	} else {

		filter = append(filter, bson.E{Key: "$text", Value: bson.M{"$search": s}})

		// quoted := regexp.QuoteMeta(s)
		// filter = bson.D{{Key: "$or", Value: bson.A{
		// 	bson.M{"name": bson.M{"$regex": "^" + quoted + "$", "$options": "i"}},
		// 	bson.M{"abbreviation": bson.M{"$regex": "^" + quoted + "$", "$options": "i"}},
		// 	bson.M{"url": bson.M{"$regex": "^" + quoted + "$", "$options": "i"}},
		// }}}
	}

	err = FindOne(CollectionGroups, filter, bson.D{{"members", -1}}, nil, &group)
	if group.ID == "" {
		return group, ErrNoDocuments
	}

	return group, err
}

func getGroups(offset int64, limit int64, sort bson.D, filter bson.D, projection bson.M) (groups []Group, err error) {

	cur, ctx, err := Find(CollectionGroups, offset, limit, sort, filter, projection, nil)
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
		} else {
			groups = append(groups, group)
		}
	}

	return groups, cur.Err()
}
