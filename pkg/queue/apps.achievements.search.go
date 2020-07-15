package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type AppsAchievementsSearchMessage struct {
	AppAchievement mongo.AppAchievement `json:"app_achievement"`
	AppName        string               `json:"app_name"`
	AppOwners      int64                `json:"app_owners"`
}

func appsAchievementsSearchHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := AppsAchievementsSearchMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		achievement := elasticsearch.Achievement{}
		achievement.ID = payload.AppAchievement.Key
		achievement.AppID = payload.AppAchievement.AppID
		achievement.Name = payload.AppAchievement.Name
		achievement.Icon = payload.AppAchievement.Icon
		achievement.Description = payload.AppAchievement.Description
		achievement.Hidden = payload.AppAchievement.Hidden
		achievement.Completed = payload.AppAchievement.Completed
		achievement.AppName = payload.AppName
		achievement.AppOwners = payload.AppOwners

		if achievement.ID == "" || achievement.AppID == 0 {
			sendToFailQueue(message)
			continue
		}

		if payload.AppAchievement.Deleted {
			err = elasticsearch.DeleteDocument(elasticsearch.IndexAchievements, achievement.GetKey())
		} else {
			err = elasticsearch.IndexAchievement(achievement)
		}
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
}
