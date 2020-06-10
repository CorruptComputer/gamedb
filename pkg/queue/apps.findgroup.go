package queue

import (
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type FindGroupMessage struct {
	AppID int `json:"app_id"`
}

//noinspection RegExpRedundantEscape
var regexpGroupID = regexp.MustCompile(`\(\s?\'(\d{18})\'\s?\)`)

func appsFindGroupHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := FindGroupMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		resp, err := helpers.GetWithTimeout("https://steamcommunity.com/app/"+strconv.Itoa(payload.AppID), 0)
		if err != nil {
			steam.LogSteamError(err, message.Message.Body)
			sendToRetryQueue(message)
			continue
		}
		//noinspection GoDeferInLoop
		defer func() {
			err := resp.Body.Close()
			log.Err(err)
		}()

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToRetryQueue(message)
			continue
		}

		var groupID string
		ret := regexpGroupID.FindAllStringSubmatch(string(b), -1)
		for _, v := range ret {
			if len(v) == 2 && strings.HasPrefix(v[1], "103") {
				groupID = v[1]
			}
		}

		if groupID == "" {
			message.Ack(false)
			continue
		}

		// Update app
		filter := bson.D{
			{"_id", payload.AppID},
			{"group_id", ""},
		}

		_, err = mongo.UpdateOne(mongo.CollectionApps, filter, bson.D{{"group_id", groupID}})
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToRetryQueue(message)
			continue
		}

		// Clear cache
		err = memcache.Delete(memcache.MemcacheApp(payload.AppID).Key)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToRetryQueue(message)
			continue
		}

		//
		message.Ack(false)
	}
}
