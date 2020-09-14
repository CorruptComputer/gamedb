package queue

import (
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	influx "github.com/influxdata/influxdb1-client"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type StatsMessage struct {
	Type      mongo.StatsType `json:"type"`
	StatID    int             `json:"id"`
	AppsCount int64           `json:"apps_count"`
}

func (m StatsMessage) Queue() rabbit.QueueName {
	return QueueStats
}

func statsHandler(message *rabbit.Message) {

	payload := StatsMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToFailQueue(message)
		return
	}

	if payload.AppsCount == 0 {
		log.Err("Missing app count", zap.ByteString("message", message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	var totalApps int
	var totalAppsWithScore int
	var totalScore float32
	var totalPrice = map[steamapi.ProductCC]int{}
	var totalPlayers int

	projection := bson.M{"reviews_score": 1, "prices": 1, "player_peak_week": 1}
	filter := bson.D{{payload.Type.MongoCol(), payload.StatID}}

	err = mongo.BatchApps(filter, projection, func(apps []mongo.App) {

		for _, app := range apps {

			// Counts
			totalApps++

			if app.ReviewsScore > 0 {
				totalAppsWithScore++
			}

			// Score
			totalScore += float32(app.ReviewsScore)

			// Prices
			for k, v := range app.Prices {
				totalPrice[k] += v.Final
			}

			// Players
			totalPlayers += app.PlayerPeakWeek
		}
	})

	var meanScore float32
	var meanPlayers float64
	var meanPrice = map[steamapi.ProductCC]float32{}

	if totalAppsWithScore > 0 {
		meanScore = totalScore / float32(totalAppsWithScore)
	}

	if totalApps > 0 {

		meanPlayers = float64(totalPlayers) / float64(totalApps)

		for k, v := range totalPrice {
			meanPrice[k] = float32(v) / float32(totalApps)
		}
	}

	// Update Mongo
	filter = bson.D{
		{"type", payload.Type},
		{"id", payload.StatID},
	}
	update := bson.D{
		{Key: "apps", Value: totalApps},
		{Key: "mean_price", Value: meanPrice},
		{Key: "mean_score", Value: meanScore},
		{Key: "mean_players", Value: meanPlayers},
	}

	_, err = mongo.UpdateOne(mongo.CollectionStats, filter, update)
	if err != nil {
		log.Err(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	// Update Influx
	fields := map[string]interface{}{
		"apps_count":   totalApps,
		"apps_percent": (float64(totalApps) / float64(payload.AppsCount)) * 100,
		"mean_score":   meanScore,
		"mean_players": meanPlayers,
	}

	for k, v := range meanPrice {
		fields["mean_price_"+string(k)] = v
	}

	stat := mongo.Stat{}
	stat.Type = payload.Type
	stat.ID = payload.StatID

	point := influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementStats),
		Tags: map[string]string{
			"key": stat.GetKey(),
			// "type": string(payload.Type),
			// "id":   strconv.Itoa(payload.StatID),
		},
		Fields:    fields,
		Time:      time.Now(),
		Precision: "h",
	}

	_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, point)
	if err != nil {
		log.Err(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	//
	message.Ack()
}
