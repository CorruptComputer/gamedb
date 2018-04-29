package queue

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/steam"
	"github.com/streadway/amqp"
)

func processPlayer(msg amqp.Delivery) (ack bool, requeue bool) {

	// Get message
	message := new(PlayerMessage)

	err := json.Unmarshal(msg.Body, message)
	if err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(msg.Body))
		}

		return false, false
	}

	// Update player
	player, err := datastore.GetPlayer(message.PlayerID)
	if err != nil {
		if err != datastore.ErrNoSuchEntity {
			logger.Error(err)
			return false, true
		}
	}

	errs := player.Update("")
	if len(errs) > 0 {
		for _, v := range errs {
			logger.Error(v)
		}

		// API is probably down
		for _, v := range errs {
			if v.Error() == steam.ErrInvalidJson {
				return false, true
			}
		}
	}

	return true, false
}

type PlayerMessage struct {
	Time     time.Time
	PlayerID int
}
