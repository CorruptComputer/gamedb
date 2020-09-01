package tasks

import (
	"math"

	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsQueueYoutube struct {
	BaseTask
}

func (c AppsQueueYoutube) ID() string {
	return "apps-update-youtube"
}

func (c AppsQueueYoutube) Name() string {
	return "Queue top apps for youtube stats"
}

func (c AppsQueueYoutube) Group() TaskGroup {
	return TaskGroupApps
}

func (c AppsQueueYoutube) Cron() TaskTime {
	return CronTimeAppsYoutube
}

func (c AppsQueueYoutube) work() (err error) {

	// Each app takes 101 api "points"
	limit := math.Floor(1_000_000 / 101)

	limit = 100 // Temp

	apps, err := mongo.GetApps(0, int64(limit), bson.D{{"player_peak_week", -1}}, nil, bson.M{"_id": 1, "name": 1})
	if err != nil {
		return err
	}

	for _, app := range apps {

		err = queue.ProduceAppsYoutube(app.ID, app.Name)
		if err != nil {
			return err
		}
	}

	return nil
}
