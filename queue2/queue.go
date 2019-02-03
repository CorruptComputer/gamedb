package queue

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/log"
	"github.com/streadway/amqp"
)

type QueueName string

const (
	// C#
	QueueApps     QueueName = "Steam_Apps"
	QueuePackages QueueName = "Steam_Packages"
	QueueProfiles QueueName = "Steam_Profiles"

	// Go
	QueueAppsData     QueueName = "Steam_Apps_Data"
	QueueBundlesData  QueueName = "Steam_Bundles_Data"
	QueueChangesData  QueueName = "Steam_Changes_Data"
	QueueDelaysData   QueueName = "Steam_Delays_Data"
	QueuePackagesData QueueName = "Steam_Packages_Data"
	QueueProfilesData QueueName = "Steam_Profiles_Data"
)

var (
	consumeLock sync.Mutex
	produceLock sync.Mutex

	errInvalidQueue = errors.New("invalid queue")
	errEmptyMessage = errors.New("empty message")

	consumerConnection *amqp.Connection
	producerConnection *amqp.Connection

	consumerCloseChannel = make(chan *amqp.Error)
	producerCloseChannel = make(chan *amqp.Error)

	queues = map[QueueName]queueInterface{
		QueueAppsData:   AppQueue{},
		QueueDelaysData: DelayQueue{},
	}
)

type BaseMessage struct {
	Message interface{}

	// Retry info
	FirstSeen   time.Time
	Attempt     int
	NextAttempt time.Time

	// Limits
	MaxAttempts int
	MaxTime     time.Duration
}

func (q *BaseMessage) init() {

	if q.FirstSeen.IsZero() {
		q.FirstSeen = time.Now()
	}

}

func (q BaseMessage) requeueMessage(msg amqp.Delivery) error {

	q.Attempt++

	m := DelayMessage{}
	m.OriginalMessage = msg.Body
	m.OriginalQueue = q.Name

	// Update end time
	// var min float64 = 1
	// var max float64 = 600

	// var seconds = math.Pow(1.3, float64(q.Attempt))
	// var minmaxed = math.Min(min+seconds, max)
	// var rounded = math.Round(minmaxed)

	// q.EndTime = q.StartTime.Add(time.Second * time.Duration(rounded))

	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	err = produce(QueueDelaysData, b)
	log.Err(err)

	return nil
}

type queueInterface interface {
	setQueueName(QueueName)
	process(msg amqp.Delivery, queue QueueName) (requeue bool)
	consume()
}

type BaseQueue struct {
	queueInterface
	Name QueueName
}

func (q BaseQueue) setQueueName(name QueueName) {
	q.Name = name
}

func (q BaseQueue) consume() {

	var err error

	for {

		// Connect
		err = func() error {

			consumeLock.Lock()
			defer consumeLock.Unlock()

			if consumerConnection == nil {

				consumerConnection, err = makeAConnection()
				if err != nil {
					log.Critical("Connecting to Rabbit: " + err.Error())
					return err
				}
				consumerConnection.NotifyClose(consumerCloseChannel)
			}

			return nil
		}()

		if err != nil {
			log.Err(err)
			return
		}

		//
		ch, qu, err := getQueue(consumerConnection, q.Name)
		if err != nil {
			log.Err(err)
			return
		}

		msgs, err := ch.Consume(qu.Name, "", false, false, false, false, nil)
		if err != nil {
			log.Err(err)
			return
		}

		// In a anon function so can return at anytime
		func(msgs <-chan amqp.Delivery, q BaseQueue) {

			for {
				select {
				case err = <-consumerCloseChannel:
					log.Warning(err)
					return
				case msg := <-msgs:

					requeue := q.process(msg, q.Name)

					if requeue {
						logInfo("Requeuing")
						err = q.requeueMessage(msg)
						logError(err)
					}

					err = msg.Ack(false)
					logError(err)
				}
			}

		}(msgs, q)

		// We only get here if the amqp connection gets closed

		err = ch.Close()
		log.Err(err)
	}
}

func RunConsumers() {
	for k, v := range queues {
		v.setQueueName(k)
		go v.consume()
	}
}

func produce(queue QueueName, data []byte) (err error) {

	// log.Info("Producing to: " + q.Message.getProduceQueue().String())s

	// Connect
	err = func() error {

		produceLock.Lock()
		defer produceLock.Unlock()

		if producerConnection == nil {

			producerConnection, err = makeAConnection()
			if err != nil {
				log.Critical("Connecting to Rabbit: " + err.Error())
				return err
			}
			producerConnection.NotifyClose(producerCloseChannel)
		}

		return nil
	}()

	if err != nil {
		return err
	}

	//
	ch, qu, err := getQueue(producerConnection, queue)
	if err != nil {
		return err
	}

	// Close channel
	if ch != nil {
		defer func(ch *amqp.Channel) {
			err := ch.Close()
			log.Err(err)
		}(ch)
	}

	return ch.Publish("", qu.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         data,
	})
}

func makeAConnection() (conn *amqp.Connection, err error) {

	operation := func() (err error) {

		log.Info("Connecting to Rabbit")

		conn, err = amqp.Dial(config.Config.RabbitDSN())
		log.Err(err) // Logging here as no max elasped time
		return err
	}

	policy := backoff.NewExponentialBackOff()
	policy.MaxElapsedTime = 0

	err = backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { logInfo(err) })

	return conn, err
}

func getQueue(conn *amqp.Connection, queue QueueName) (ch *amqp.Channel, qu amqp.Queue, err error) {

	ch, err = conn.Channel()
	if err != nil {
		return
	}

	err = ch.Qos(10, 0, false)
	if err != nil {
		return
	}

	qu, err = ch.QueueDeclare(string(queue), true, false, false, false, nil)

	return ch, qu, err
}

//
type SteamKitJob struct {
	SequentialCount int    `json:"SequentialCount"`
	StartTime       string `json:"StartTime"`
	ProcessID       int    `json:"ProcessID"`
	BoxID           int    `json:"BoxID"`
	Value           int64  `json:"Value"`
}

func logInfo(interfaces ...interface{}) {
	log.Info(append(interfaces, log.LogNameConsumers)...)
}

func logError(interfaces ...interface{}) {
	log.Err(append(interfaces, log.LogNameConsumers)...)
}

func logWarning(interfaces ...interface{}) {
	log.Warning(append(interfaces, log.LogNameConsumers)...)
}
