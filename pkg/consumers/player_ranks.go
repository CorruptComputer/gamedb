package consumers

import (
	"time"

	"github.com/gamedb/gamedb/pkg/consumers/framework"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	mongodb "go.mongodb.org/mongo-driver/mongo"
)

type PlayerRanksMessage struct {
	ObjectKey  string  `json:"object_key"`
	SortColumn string  `json:"sort_column"`
	Continent  *string `json:"continent"`
	Country    *string `json:"country"`
	State      *string `json:"state"`
}

func (msg PlayerRanksMessage) Produce() error {
	return queues[framework.Producer][queuePlayerRanks].ProduceInterface(msg)
}

func playerRanksHandler(messages []framework.Message) {

	for _, message := range messages {

		payload := PlayerRanksMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			return
		}

		if payload.ObjectKey == "" || payload.SortColumn == "" {
			sendToFailQueue(message)
			return
		}

		// Create filter
		var filter = bson.D{}
		if payload.Continent != nil {
			filter = append(filter, bson.E{Key: "continent_code", Value: *payload.Continent})
		}
		if payload.Country != nil {
			filter = append(filter, bson.E{Key: "country_code", Value: *payload.Country})
		}
		if payload.State != nil {
			filter = append(filter, bson.E{Key: "status_code", Value: *payload.State})
		}
		filter = append(filter, bson.E{Key: payload.SortColumn, Value: bson.M{"$exists": true, "$gt": 0}}) // Put last to help indexes

		log.Info(message.Queue.Name, filter)

		// Get players
		players, err := mongo.GetPlayers(0, 0, bson.D{{payload.SortColumn, -1}}, filter, bson.M{"_id": 1})
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			return
		}

		// Build bulk update
		var writes []mongodb.WriteModel
		for position, player := range players {

			write := mongodb.NewUpdateOneModel()
			write.SetFilter(bson.M{"_id": player.ID})
			write.SetUpdate(bson.M{"$set": bson.M{"ranks." + payload.ObjectKey: position + 1}})
			write.SetUpsert(true)

			writes = append(writes, write)
		}

		// Update player ranks
		chunks := mongo.ChunkWriteModels(writes, 100000)
		for _, chunk := range chunks {

			err = mongo.BulkUpdatePlayers(chunk)
			if val, ok := err.(mongodb.BulkWriteException); ok {
				for _, err2 := range val.WriteErrors {
					log.Err(err2, err2.Request)
				}
			} else {
				log.Err(err)
			}

			time.Sleep(time.Second)
		}

		message.Ack()
	}
}
