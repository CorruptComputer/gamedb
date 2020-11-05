package tasks

import (
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type PlayersQueueElastic struct {
	BaseTask
}

func (c PlayersQueueElastic) ID() string {
	return "players-queue-elastic"
}

func (c PlayersQueueElastic) Name() string {
	return "Queue all players to Elastic"
}

func (c PlayersQueueElastic) Group() TaskGroup {
	return TaskGroupElastic
}

func (c PlayersQueueElastic) Cron() TaskTime {
	return ""
}

func (c PlayersQueueElastic) work() (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		players, err := mongo.GetPlayers(offset, limit, bson.D{{"_id", 1}}, nil, nil)
		if err != nil {
			return err
		}

		for _, player := range players {

			err = queue.ProducePlayerSearch(&player, 0)
			if err != nil {
				return err
			}
		}

		if int64(len(players)) != limit {
			break
		}

		offset += limit
	}

	return nil
}
