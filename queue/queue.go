package queue

import (
	"errors"
	"os"
	"time"

	"github.com/steam-authority/steam-authority/logger"
	"github.com/streadway/amqp"
)

const (
	namespace = "STEAM_"

	ChangeQueue  = "Changes"
	AppQueue     = "Apps"
	PackageQueue = "Packages"
	PlayerQueue  = "Players"

	HeaderRetry = "retry"
)

var (
	queues          map[string]queue
	enableConsumers bool = true
)

func init() {

	qs := []queue{
		{Name: ChangeQueue, Callback: processChange},
		{Name: AppQueue, Callback: processApp},
		{Name: PackageQueue, Callback: processPackage},
		{Name: PlayerQueue, Callback: processPlayer},
	}

	queues = make(map[string]queue)
	for _, v := range qs {
		queues[v.Name] = v
	}
}

func RunConsumers() {

	for _, v := range queues {
		go v.consume()
	}
}

// todo, use interface so we can set the payload time in here?
func Produce(queue string, data []byte) (err error) {

	if val, ok := queues[queue]; ok {
		return val.produce(data)
	}

	return errors.New("no such queue")
}

type queue struct {
	Name     string
	Callback func(msg amqp.Delivery) (ack bool, requeue bool)
}

func (s queue) getConnection() (conn *amqp.Connection, ch *amqp.Channel, q amqp.Queue, closeChannel chan *amqp.Error, err error) {

	closeChannel = make(chan *amqp.Error)

	conn, err = amqp.Dial(os.Getenv("STEAM_AMQP"))
	conn.NotifyClose(closeChannel)
	if err != nil {
		logger.Error(err)
	}

	ch, err = conn.Channel()
	if err != nil {
		logger.Error(err)
	}

	q, err = ch.QueueDeclare(namespace+s.Name, true, false, false, false, nil)
	if err != nil {
		logger.Error(err)
	}

	return conn, ch, q, closeChannel, err
}

func (s queue) produce(data []byte) (err error) {

	conn, ch, q, _, err := s.getConnection()
	defer conn.Close()
	defer ch.Close()
	if err != nil {
		return err
	}

	err = ch.Publish("", q.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         data,
		Headers: amqp.Table{
			HeaderRetry: int16(0),
		},
	})
	if err != nil {
		logger.Error(err)
	}

	return nil

}

func (s queue) consume() {

	var breakFor = false

	for {
		conn, ch, q, closeChan, err := s.getConnection()
		if err != nil {
			logger.Error(err)
			return
		}

		msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
		if err != nil {
			logger.Error(err)
			return
		}

		for {
			select {
			case err = <-closeChan:
				breakFor = true
				break

			case msg := <-msgs:

				ack, requeue := s.Callback(msg)
				if ack {
					msg.Ack(false)
				} else {

					if requeue {
						go func() {
							retrys := incRetrys(&msg)
							time.Sleep(time.Second * 10 * time.Duration(retrys)) // Exponential backoff
							msg.Nack(false, requeue)
						}()
					} else {
						msg.Nack(false, requeue)
					}
				}

				time.Sleep(time.Second * 2)
			}

			if breakFor {
				break
			}
		}

		conn.Close()
		ch.Close()
	}
}

func incRetrys(msg *amqp.Delivery) int16 {

	var retrys int16

	if msg.Headers == nil {

		msg.Headers = amqp.Table{}
		retrys = 1

	} else {

		retrysInterface, ok := msg.Headers[HeaderRetry]
		if ok {

			retrysInt, ok := retrysInterface.(int16)
			if ok {
				retrys = retrysInt + 1
			} else {
				retrys = 1
			}

		} else {
			retrys = 1
		}
	}

	msg.Headers[HeaderRetry] = retrys

	return retrys
}
