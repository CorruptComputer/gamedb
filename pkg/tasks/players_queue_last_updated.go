package tasks

import (
	"strconv"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/tasks/helpers/rabbitweb"
	"go.mongodb.org/mongo-driver/bson"
)

type PlayersQueueLastUpdated struct {
	BaseTask
}

func (c PlayersQueueLastUpdated) ID() string {
	return "update-last-updated-players"
}

func (c PlayersQueueLastUpdated) Name() string {
	return "Update last updated players"
}

func (c PlayersQueueLastUpdated) Group() TaskGroup {
	return TaskGroupPlayers
}

func (c PlayersQueueLastUpdated) Cron() TaskTime {
	return CronTimeUpdateLastUpdatedPlayers
}

const toQueue = 10
const cronTime = time.Minute

func (c PlayersQueueLastUpdated) work() (err error) {

	// Skip if queues have activity
	limits := map[rabbit.QueueName]int{
		queue.QueueApps:     50,
		queue.QueuePackages: 50,
		queue.QueuePlayers:  5,
	}

	queues, err := rabbitweb.GetRabbitWebClient().GetQueues()
	if err != nil {
		return err
	}

	var consumers int
	for _, q := range queues {
		if val, ok := limits[rabbit.QueueName(q.Name)]; ok && q.Messages > val {
			log.InfoS("skipping " + c.ID() + " as " + q.Name + " has " + strconv.Itoa(q.Messages) + " messages")
			return nil
		}
		if q.Name == string(queue.QueuePlayers) {
			consumers = q.Consumers
		}
	}

	if consumers == 0 {
		if config.IsLocal() {
			consumers = 1
		} else {
			log.WarnS("no consumers")
			return nil
		}
	}

	// Queue last updated players
	players, err := mongo.GetPlayers(0, int64(toQueue*consumers), bson.D{{"updated_at", 1}}, helpers.LastUpdatedQuery, bson.M{"_id": 1})
	if err != nil {
		return err
	}

	for _, player := range players {

		m := queue.PlayerMessage{
			ID:               player.ID,
			SkipGroupUpdate:  true,
			SkipAchievements: true,
		}

		err = queue.ProducePlayer(m, "crons-last-updated")
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		if err != nil {
			return err
		}

		time.Sleep(cronTime / time.Duration(toQueue*consumers))
	}

	return err
}
