package queue

import (
	"time"

	"github.com/steam-authority/steam-authority/db"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logging"
	"github.com/streadway/amqp"
)

type RabbitMessageProfile struct {
	Time     time.Time
	PlayerID int64
}

func (d RabbitMessageProfile) getQueueName() string {
	return QueueProfilesData
}

func processPlayer(msg amqp.Delivery) (ack bool, requeue bool) {

	// Get message
	message := new(RabbitMessageProfile)

	err := helpers.Unmarshal(msg.Body, message)
	if err != nil {
		return false, false
	}

	// Update player
	player, err := db.GetPlayer(int64(message.PlayerID))
	if err != nil {
		if err != db.ErrNoSuchEntity {
			logging.Error(err)
			return false, true
		}
	}

	errs := player.Update("")
	if len(errs) > 0 {
		for _, v := range errs {
			logging.Error(v)
		}

		// API is probably down, todo
		//for _, v := range errs {
		//	if v.Error() == steam.ErrInvalidJson {
		//		return false, true
		//	}
		//}
	}

	return true, false
}
