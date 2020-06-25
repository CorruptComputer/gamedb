package tasks

import (
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type AutoPlayerRefreshes struct {
	BaseTask
}

func (c AutoPlayerRefreshes) ID() string {
	return "update-donator-profiles"
}

func (c AutoPlayerRefreshes) Name() string {
	return "Update donator profiles"
}

func (c AutoPlayerRefreshes) Cron() string {
	return CronTimeAutoPlayerRefreshes
}

func (c AutoPlayerRefreshes) work() (err error) {

	// Get users
	db, err := mysql.GetMySQLClient()
	if err != nil {
		return err
	}

	var users []mysql.User
	db = db.Select([]string{"steam_id", "steam_id"})
	db = db.Where("level >= ?", mysql.UserLevel3)
	db = db.Where("steam_id > ?", 0)
	db = db.Find(&users)
	if db.Error != nil {
		return db.Error
	}

	// Update players
	var playerIDs []int64
	for _, user := range users {

		playerID := user.GetSteamID()

		if playerID > 0 {

			playerIDs = append(playerIDs, playerID)

			err = queue.ProducePlayer(queue.PlayerMessage{ID: playerID})
			err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
			if err != nil {
				return err
			}
		}
	}

	// Update groups
	players, err := mongo.GetPlayersByID(playerIDs, bson.M{"primary_clan_id_string": 1})
	if err != nil {
		return err
	}

	for _, v := range players {
		if v.PrimaryGroupID != "" {
			err = queue.ProduceGroup(queue.GroupMessage{ID: v.PrimaryGroupID})
			err = helpers.IgnoreErrors(err, queue.ErrIsBot, memcache.ErrInQueue)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
