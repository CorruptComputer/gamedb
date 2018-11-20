package queue

import (
	"errors"
	"net/http"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/streadway/amqp"
)

type RabbitMessageProfile struct {
	ProfileInfo RabbitMessageProfilePICS `json:"ProfileInfo"`
}

func (d RabbitMessageProfile) getQueueName() string {
	return QueueProfilesData
}

func (d RabbitMessageProfile) getRetryData() RabbitMessageDelay {
	return RabbitMessageDelay{}
}

func (d RabbitMessageProfile) process(msg amqp.Delivery) (ack bool, requeue bool, err error) {

	// Get message
	rabbitMessage := new(RabbitMessageProfile)

	err = helpers.Unmarshal(msg.Body, rabbitMessage)
	if err != nil {
		return false, false, err
	}

	var message = rabbitMessage.ProfileInfo

	if !message.SteamID.IsValid {
		return false, false, errors.New("not valid account id")
	}
	if !message.SteamID.IsIndividualAccount {
		return false, false, errors.New("not individual account id")
	}

	// Update player
	player, err := db.GetPlayer(message.SteamID.AccountID)
	err = helpers.IgnoreErrors(datastore.ErrNoSuchEntity)
	if err != nil {
		return false, true, err
	}

	err = player.ShouldUpdate(new(http.Request), db.PlayerUpdateAdmin)
	err = helpers.IgnoreErrors(db.ErrUpdatingPlayerTooSoon, db.ErrUpdatingPlayerInQueue)
	if err != nil {
		return false, false, err
	}

	player.PlayerID = message.SteamID.AccountID
	player.RealName = message.RealName
	player.StateCode = message.StateName
	player.CountryCode = message.CountryName

	err = player.Update()
	if err != nil {
		return false, true, err
	}

	return true, false, err
}

type RabbitMessageProfilePICS struct {
	Result  int `json:"Result"`
	SteamID struct {
		IsBlankAnonAccount            bool  `json:"IsBlankAnonAccount"`
		IsGameServerAccount           bool  `json:"IsGameServerAccount"`
		IsPersistentGameServerAccount bool  `json:"IsPersistentGameServerAccount"`
		IsAnonGameServerAccount       bool  `json:"IsAnonGameServerAccount"`
		IsContentServerAccount        bool  `json:"IsContentServerAccount"`
		IsClanAccount                 bool  `json:"IsClanAccount"`
		IsChatAccount                 bool  `json:"IsChatAccount"`
		IsLobby                       bool  `json:"IsLobby"`
		IsIndividualAccount           bool  `json:"IsIndividualAccount"`
		IsAnonAccount                 bool  `json:"IsAnonAccount"`
		IsAnonUserAccount             bool  `json:"IsAnonUserAccount"`
		IsConsoleUserAccount          bool  `json:"IsConsoleUserAccount"`
		IsValid                       bool  `json:"IsValid"`
		AccountID                     int64 `json:"AccountID"`
		AccountInstance               int   `json:"AccountInstance"`
		AccountType                   int   `json:"AccountType"`
		AccountUniverse               int   `json:"AccountUniverse"`
	} `json:"SteamID"`
	TimeCreated time.Time   `json:"TimeCreated"`
	RealName    string      `json:"RealName"`
	CityName    string      `json:"CityName"`
	StateName   string      `json:"StateName"`
	CountryName string      `json:"CountryName"`
	Headline    string      `json:"Headline"`
	Summary     string      `json:"Summary"`
	JobID       SteamKitJob `json:"JobID"`
}
