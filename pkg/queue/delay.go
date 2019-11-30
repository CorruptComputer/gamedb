package queue

import (
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type delayQueue struct {
	BaseQueue baseQueue
}

func (q delayQueue) processMessages(msgs []amqp.Delivery) {

	time.Sleep(time.Second / 10)

	msg := msgs[0]

	var err error
	var message baseMessage

	err = helpers.Unmarshal(msg.Body, &message)
	if err != nil {
		log.Err(err)
		ackFail(msg, &message)
		return
	}

	// Limits
	if q.BaseQueue.getMaxTime() > 0 && message.FirstSeen.Add(q.BaseQueue.getMaxTime()).Before(time.Now()) {

		log.Info("Message removed from delay queue (Over " + q.BaseQueue.getMaxTime().String() + " / " + message.FirstSeen.Add(q.BaseQueue.getMaxTime()).String() + "): " + string(msg.Body))
		ackFail(msg, &message)
		return
	}

	if q.BaseQueue.maxAttempts > 0 && message.Attempt > q.BaseQueue.maxAttempts {

		log.Info("Message removed from delay queue (" + strconv.Itoa(message.Attempt) + "/" + strconv.Itoa(q.BaseQueue.maxAttempts) + " attempts): " + string(msg.Body))
		ackFail(msg, &message)
		return
	}

	if message.OriginalQueue == queueDelays {

		log.Info("Message removed from delay queue (Stuck in delay queue): " + string(msg.Body))
		ackFail(msg, &message)
		return
	}

	//
	var queue queueName

	if message.getNextAttempt().Before(time.Now()) {

		log.Info("Sending back to " + string(message.OriginalQueue))
		queue = message.OriginalQueue

	} else {

		// log.Info("Sending " + msg.MessageId + " back in " + message.getNextAttempt().Sub(time.Now()).String())
		queue = queueDelays
	}

	switch message.OriginalQueue {
	case queueApps:

		message2 := appMessage{}
		message2.Force = message.Force

		err = helpers.Unmarshal(msg.Body, &message2)
		if err != nil {
			log.Err(err)
			return
		}

		err = produce(&message2, queue)
		if err != nil {
			log.Err(err)
			return
		}

	case queueAppPlayer:

		message2 := appPlayerMessage{}
		message2.Force = message.Force

		err = helpers.Unmarshal(msg.Body, &message2)
		if err != nil {
			log.Err(err)
			return
		}

		err = produce(&message2, queue)
		if err != nil {
			log.Err(err)
			return
		}

	case queueBundles:

		message2 := bundleMessage{}
		message2.Force = message.Force

		err = helpers.Unmarshal(msg.Body, &message2)
		if err != nil {
			log.Err(err)
			return
		}

		err = produce(&message2, queue)
		if err != nil {
			log.Err(err)
			return
		}

	case queueChanges:

		message2 := changeMessage{}
		message2.Force = message.Force

		err = helpers.Unmarshal(msg.Body, &message2)
		if err != nil {
			log.Err(err)
			return
		}

		err = produce(&message2, queue)
		if err != nil {
			log.Err(err)
			return
		}

	case queueGroups, queueGroups2:

		message2 := groupMessage{}
		message2.Force = message.Force

		err = helpers.Unmarshal(msg.Body, &message2)
		if err != nil {
			log.Err(err)
			return
		}

		err = produce(&message2, queue)
		if err != nil {
			log.Err(err)
			return
		}

	case queueGroupsNew:

		message2 := groupMessage{}
		message2.Force = message.Force

		err = helpers.Unmarshal(msg.Body, &message2)
		if err != nil {
			log.Err(err)
			return
		}

		err = produce(&message2, queue)
		if err != nil {
			log.Err(err)
			return
		}

	case queuePackages:

		message2 := packageMessage{}
		message2.Force = message.Force

		err = helpers.Unmarshal(msg.Body, &message2)
		if err != nil {
			log.Err(err)
			return
		}

		err = produce(&message2, queue)
		if err != nil {
			log.Err(err)
			return
		}

	case queuePlayers, queuePlayers2:

		message2 := playerMessage{}
		message2.Force = message.Force

		err = helpers.Unmarshal(msg.Body, &message2)
		if err != nil {
			log.Err(err)
			return
		}

		err = produce(&message2, queue)
		if err != nil {
			log.Err(err)
			return
		}

	case QueueSteam:

		message2 := steamMessage{}
		message2.Force = message.Force

		err = helpers.Unmarshal(msg.Body, &message2)
		if err != nil {
			log.Err(err)
			return
		}

		err = produce(&message2, queue)
		if err != nil {
			log.Err(err)
			return
		}

	case queueTest:

		message2 := testMessage{}
		message2.Force = message.Force

		err = helpers.Unmarshal(msg.Body, &message2)
		if err != nil {
			log.Err(err)
			return
		}

		err = produce(&message2, queue)
		if err != nil {
			log.Err(err)
			return
		}

	default:
		log.Critical("Wrong message type", msg.Body)
		ackFail(msg, &message)
		return
	}

	err = msg.Ack(false)
	if err != nil {
		log.Err(err)
		return
	}
}
