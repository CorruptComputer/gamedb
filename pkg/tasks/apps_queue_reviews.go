package tasks

import (
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

func (c AppsQueueReviews) Cron() string {
	return CronTimeAppsReviews
}

func (c AppsQueueReviews) work() (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		var filter = bson.D{{"reviews_count", bson.M{"$gt": 0}}}

		apps, err := mongo.GetApps(offset, limit, bson.D{{"_id", 1}}, filter, bson.M{"_id": 1}, nil)
		if err != nil {
			return err
		}

		for _, app := range apps {

			err = queue.ProduceAppsReviews(app.ID)
			if err != nil {
				return err
			}
		}

		if int64(len(apps)) != limit {
			break
		}

		offset += limit
	}

	return nil
}
