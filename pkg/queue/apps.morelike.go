package queue

import (
	"regexp"
	"strconv"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/steam"
	"github.com/gocolly/colly/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type AppMorelikeMessage struct {
	AppID int `json:"id"`
}

func (m AppMorelikeMessage) Queue() rabbit.QueueName {
	return QueueAppsMorelike
}

func appMorelikeHandler(message *rabbit.Message) {

	payload := AppMorelikeMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		zap.L().Error(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToFailQueue(message)
		return
	}

	c := colly.NewCollector(
		colly.URLFilters(regexp.MustCompile(`store\.steampowered\.com/recommended/morelike/app/[0-9]+$`)),
		steam.WithAgeCheckCookie,
		steam.WithTimeout(0),
	)

	var relatedAppIDs []int

	c.OnHTML(".similar_grid_capsule", func(e *colly.HTMLElement) {
		i, err := strconv.Atoi(e.Attr("data-ds-appid"))
		if err == nil {
			relatedAppIDs = append(relatedAppIDs, i)
		}
	})

	err = c.Visit("https://store.steampowered.com/recommended/morelike/app/" + strconv.Itoa(payload.AppID))
	if err != nil {
		steam.LogSteamError(err)
		sendToRetryQueue(message)
		return
	}

	if len(relatedAppIDs) == 0 {
		message.Ack()
		return
	}

	// Update app
	_, err = mongo.UpdateOne(mongo.CollectionApps, bson.D{{"_id", payload.AppID}}, bson.D{{"related_app_ids", relatedAppIDs}})
	if err != nil {
		zap.S().Error(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	// Clear cache
	err = memcache.Delete(memcache.MemcacheApp(payload.AppID).Key)
	if err != nil {
		zap.S().Error(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	// Update in Elastic
	err = ProduceAppSearch(nil, payload.AppID)
	if err != nil {
		zap.S().Error(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	message.Ack()
}
