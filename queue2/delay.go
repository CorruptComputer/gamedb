package queue

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/gamedb/website/helpers"
	"github.com/streadway/amqp"
)

type RabbitMessageDelay struct {
	rabbitConsumer
	OriginalQueue   QueueName
	OriginalMessage string
}

func (d RabbitMessageDelay) getConsumeQueue() QueueName {
	return QueueDelaysData
}

func (d RabbitMessageDelay) getProduceQueue() QueueName {
	return ""
}

func (d RabbitMessageDelay) getRetryData() RabbitMessageDelay {
	return RabbitMessageDelay{}
}

func (d RabbitMessageDelay) process(msg amqp.Delivery) (requeue bool, err error) {

	if len(msg.Body) == 0 {
		return false, errEmptyMessage
	}

	delayMessage := RabbitMessageDelay{}

	err = helpers.Unmarshal(msg.Body, &delayMessage)
	if err != nil {
		return false, err
	}

	if len(delayMessage.OriginalMessage) == 0 {
		return false, errEmptyMessage
	}

	if delayMessage.EndTime.UnixNano() > time.Now().UnixNano() {

		// Re-delay
		logInfo("Re-delay: attemp: " + strconv.Itoa(delayMessage.Attempt))

		delayMessage.IncrementAttempts()

		b, err := json.Marshal(delayMessage)
		if err != nil {
			return false, err
		}

		err = Produce(delayMessage.getConsumeQueue(), b)
		logError(err)

	} else {

		// Add to original queue
		logInfo("Re-trying after attempt: " + strconv.Itoa(delayMessage.Attempt))

		err = Produce(delayMessage.getConsumeQueue(), []byte(delayMessage.OriginalMessage))
	}

	if err != nil {
		return true, err
	}

	return false, nil
}
