package tasks

import (
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsPlayerCheckTop struct {
	BaseTask
}

func (c AppsPlayerCheckTop) ID() string {
	return "app-players-top"
}

func (c AppsPlayerCheckTop) Name() string {
	return "Check apps for players (Top)"
}

func (c AppsPlayerCheckTop) Group() string {
	return TaskGroupApps
}

func (c AppsPlayerCheckTop) Cron() string {
	return CronTimeAppPlayersTop
}

const topAppPlayers = 10 // And up are top apps

func (c AppsPlayerCheckTop) work() (err error) {

	var filter = bson.D{{"player_peak_week", bson.M{"$gte": topAppPlayers}}}
	var projection = bson.M{"_id": 1}

	return mongo.BatchApps(filter, projection, func(apps []mongo.App) {

		var ids []int
		for _, v := range apps {
			if v.ID > 0 { // This is just here to stop storing things on app 0, which we use to store steam stats on
				ids = append(ids, v.ID)
			}
		}

		var chunks = helpers.ChunkInts(ids, 20)

		for _, chunk := range chunks {
			err = queue.ProduceAppPlayersTop(chunk)
			if err != nil {
				log.Err(err)
				return
			}
		}
	})
}
