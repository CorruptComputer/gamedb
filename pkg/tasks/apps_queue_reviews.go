package tasks

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsQueueReviews struct {
	BaseTask
}

func (c AppsQueueReviews) ID() string {
	return "apps-queue-reviews"
}

func (c AppsQueueReviews) Name() string {
	return "Update all app reviews"
}

func (c AppsQueueReviews) Group() string {
	return TaskGroupApps
}

func (c AppsQueueReviews) Cron() string {
	return CronTimeAppsReviews
}

func (c AppsQueueReviews) work() (err error) {

	var filter = bson.D{{"reviews_count", bson.M{"$gt": 0}}}
	var projection = bson.M{"_id": 1}

	return mongo.BatchApps(filter, projection, func(apps []mongo.App) {

		for _, app := range apps {

			err = queue.ProduceAppsReviews(app.ID)
			if err != nil {
				log.Err(err)
				return
			}
		}
	})
}
