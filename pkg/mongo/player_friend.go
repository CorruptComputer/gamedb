package mongo

import (
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PlayerFriend struct {
	PlayerID     int64     `bson:"player_id"`
	FriendID     int64     `bson:"friend_id"`
	FriendSince  time.Time `bson:"since"`
	Avatar       string    `bson:"avatar"`
	Name         string    `bson:"name"`
	Games        int       `bson:"games"`
	Level        int       `bson:"level"`
	LoggedOff    time.Time `bson:"logged_off"`
	Relationship string    `bson:"relationship"`
}

func (friend PlayerFriend) BSON() bson.D {
	return bson.D{
		{"_id", friend.getKey()},
		{"player_id", friend.PlayerID},
		{"friend_id", friend.FriendID},
		{"since", friend.FriendSince},
		{"avatar", friend.Avatar},
		{"name", friend.Name},
		{"games", friend.Games},
		{"level", friend.Level},
		{"logged_off", friend.LoggedOff},
	}
}

func (friend PlayerFriend) getKey() (ret interface{}) {

	return strconv.FormatInt(friend.PlayerID, 10) + "-" + strconv.FormatInt(friend.FriendID, 10)
}

func (friend PlayerFriend) Scanned() bool {
	return !friend.LoggedOff.IsZero()
}

func (friend PlayerFriend) GetPath() string {
	return helpers.GetPlayerPath(friend.FriendID, friend.Name)
}

func (friend PlayerFriend) GetLoggedOff() string {
	if friend.Scanned() {
		return friend.LoggedOff.Format(helpers.DateYearTime)
	}
	return "-"
}

func (friend PlayerFriend) GetFriendSince() string {
	return friend.FriendSince.Format(helpers.DateYearTime)
}

func (friend PlayerFriend) GetName() string {
	return helpers.GetPlayerName(friend.FriendID, friend.Name)
}

func (friend PlayerFriend) GetLevel() string {
	if friend.Scanned() {
		return humanize.Comma(int64(friend.Level))
	}
	return "-"
}

func (friend PlayerFriend) CommunityLink() string {
	return "https://steamcommunity.com/profiles/" + strconv.FormatInt(friend.FriendID, 10)
}

func CountFriends(playerID int64) (count int64, err error) {

	return CountDocuments(CollectionPlayerFriends, bson.D{{"player_id", playerID}}, 0)
}

func DeleteFriends(playerID int64, friends []int64) (err error) {

	if len(friends) < 1 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	keys := bson.A{}
	for _, friendID := range friends {

		friend := PlayerFriend{}
		friend.PlayerID = playerID
		friend.FriendID = friendID

		keys = append(keys, friend.getKey())
	}

	collection := client.Database(MongoDatabase).Collection(CollectionPlayerFriends.String())
	_, err = collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": keys}})
	return err
}

func UpdateFriends(friends []*PlayerFriend) (err error) {

	if len(friends) < 1 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	var writes []mongo.WriteModel
	for _, friend := range friends {

		write := mongo.NewReplaceOneModel()
		write.SetFilter(bson.M{"_id": friend.getKey()})
		write.SetReplacement(friend.BSON())
		write.SetUpsert(true)

		writes = append(writes, write)
	}

	collection := client.Database(MongoDatabase).Collection(CollectionPlayerFriends.String())
	_, err = collection.BulkWrite(ctx, writes, options.BulkWrite())
	return err
}

func GetFriends(playerID int64, offset int64, limit int64, sort bson.D) (friends []PlayerFriend, err error) {

	var filter = bson.D{{"player_id", playerID}}

	cur, ctx, err := Find(CollectionPlayerFriends, offset, limit, sort, filter, nil, nil)
	if err != nil {
		return friends, err
	}

	defer func(cur *mongo.Cursor) {
		err = cur.Close(ctx)
		log.Err(err)
	}(cur)

	for cur.Next(ctx) {

		var friend PlayerFriend
		err := cur.Decode(&friend)
		if err != nil {
			log.Err(err, friend.getKey())
		} else {
			friends = append(friends, friend)
		}
	}

	return friends, cur.Err()
}
