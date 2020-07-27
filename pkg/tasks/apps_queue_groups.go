package tasks

import (
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsQueueGroups struct {
	BaseTask
}

func (c AppsQueueGroups) ID() string {
	return "queue-app-groups"
}

func (c AppsQueueGroups) Name() string {
	return "Queue app groups"
}

func (c AppsQueueGroups) Group() string {
	return TaskGroupGroups
}

func (c AppsQueueGroups) Cron() string {
	return CronTimeQueueAppGroups
}

func (c AppsQueueGroups) work() (err error) {

	var filter = bson.D{{"group_id", bson.M{"$ne": ""}}}
	var projection = bson.M{"group_id": 1}

	return mongo.BatchApps(filter, projection, func(apps []mongo.App) {

		for _, app := range apps {

			err = queue.ProduceGroup(queue.GroupMessage{ID: app.GroupID})
			err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
			if err != nil {
				log.Err(err)
				return
			}
		}
	})
}
