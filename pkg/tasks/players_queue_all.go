package tasks

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
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

func (c PlayersQueueAll) Cron() string {
	return ""
}

func (c PlayersQueueAll) work() (err error) {

	players, err := mongo.GetPlayers(0, 0, bson.D{{"_id", 1}}, nil, bson.M{"_id": 1})
	if err != nil {
		return err
	}

	var playerIDs []int64
	for _, player := range players {
		playerIDs = append(playerIDs, player.ID)
	}

	err = queue.ProduceToSteam(queue.SteamPayload{ProfileIDs: playerIDs, Force: true})
	if err != nil {
		return err
	}

	//
	log.Info(strconv.Itoa(len(players)) + " players added to rabbit")

	return nil
}
