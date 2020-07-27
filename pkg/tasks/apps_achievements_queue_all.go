package tasks

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsAchievementsQueueAll struct {
	BaseTask
}

func (c AppsAchievementsQueueAll) ID() string {
	return "queue-all-app-achievements"
}

func (c AppsAchievementsQueueAll) Name() string {
	return "Queue all app achievements"
}

func (c AppsAchievementsQueueAll) Group() string {
	return TaskGroupApps
}

func (c AppsAchievementsQueueAll) Cron() string {
	return ""
}

func (c AppsAchievementsQueueAll) work() (err error) {

	var projection = bson.M{"_id": 1, "name": 1, "owners": 1}

	return mongo.BatchApps(nil, projection, func(apps []mongo.App) {

		for _, app := range apps {

			err = queue.ProduceAppAchievement(app.ID, app.Name, app.Owners)
			if err != nil {
				log.Err(err)
				return
			}
		}
	})
}
