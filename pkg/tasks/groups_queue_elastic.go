package tasks

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type GroupsQueueElastic struct {
	BaseTask
}

func (c GroupsQueueElastic) ID() string {
	return "groups-queue-elastic"
}

func (c GroupsQueueElastic) Name() string {
	return "Queue all groups to Elastic"
}

func (c GroupsQueueElastic) Cron() string {
	return ""
}

func (c GroupsQueueElastic) work() (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		var projection = bson.M{}

		groups, err := mongo.GetGroups(limit, offset, bson.D{{"_id", 1}}, nil, projection)
		if err != nil {
			return err
		}

		for _, group := range groups {

			err = queue.ProduceGroupSearch(group)
			log.Err(err)
		}

		if int64(len(groups)) != limit {
			break
		}

		offset += limit
	}

	return nil
}
