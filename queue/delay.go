package queue

import (
	"strconv"
	"time"

	"github.com/gamedb/website/helpers"
	"github.com/streadway/amqp"
)

type delayQueue struct {
	baseQueue
}

func (q delayQueue) processMessage(msg amqp.Delivery) {

	time.Sleep(time.Second / 10)

	var err error
	var payload = baseMessage{}

	err = helpers.Unmarshal(msg.Body, &payload)
	if err != nil {
		logError(err)
		return
	}

	// Limits
	if q.maxTime > 0 && payload.FirstSeen.Add(q.maxTime).Unix() < time.Now().Unix() {

		logInfo("Message removed from delay queue (Over " + q.maxTime.String() + "): " + string(msg.Body))
		payload.ack(msg)
		return
	}

	if q.maxAttempts > 0 && payload.Attempt > q.maxAttempts {

		logInfo("Message removed from delay queue (" + strconv.Itoa(payload.Attempt) + "/" + strconv.Itoa(q.maxAttempts) + " attempts): " + string(msg.Body))
		payload.ack(msg)
		return
	}

	//
	var queue queueName

	if payload.getNextAttempt().Unix() <= time.Now().Unix() {

		logInfo("Sending back")
		queue = payload.OriginalQueue

	} else {

		// logInfo("Sending back in " + payload.NextAttempt.Sub(time.Now()).String())
		queue = queueGoDelays
	}

	err = produce(payload, queue)
	if err != nil {
		logError(err)
		return
	}

	if err == nil {
		err = msg.Ack(false)
		logError(err)
	}
}
