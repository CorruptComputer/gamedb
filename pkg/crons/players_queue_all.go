package crons

import (
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type PlayersQueueAll struct {
	BaseTask
}

func (c PlayersQueueAll) ID() string {
	return "queue-all-players"
}

func (c PlayersQueueAll) Name() string {
	return "Queue all players"
}

func (c PlayersQueueAll) Group() TaskGroup {
	return TaskGroupPlayers
}

func (c PlayersQueueAll) Cron() TaskTime {
	return ""
}

func (c PlayersQueueAll) work() (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		players, err := mongo.GetPlayers(offset, limit, bson.D{{"_id", 1}}, nil, bson.M{"_id": 1})
		if err != nil {
			return err
		}

		for _, player := range players {

			err = consumers.ProducePlayer(consumers.PlayerMessage{ID: player.ID}, "crons-queue-all")
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
