package queue

import (
	"encoding/json"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/cenkalti/backoff"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/log"
	"github.com/streadway/amqp"
)

var consumeLock sync.Mutex
var produceLock sync.Mutex

type RabbitQueue string

func (rq RabbitQueue) String() string {
	return string(rq)
}

const (
	QueueApps         RabbitQueue = "Steam_Apps"          // Only takes IDs
	QueueAppsData     RabbitQueue = "Steam_Apps_Data"     //
	QueueBundlesData  RabbitQueue = "Steam_Bundles_Data"  //
	QueueChangesData  RabbitQueue = "Steam_Changes_Data"  //
	QueueDelaysData   RabbitQueue = "Steam_Delays_Data"   //
	QueuePackages     RabbitQueue = "Steam_Packages"      // Only takes IDs
	QueuePackagesData RabbitQueue = "Steam_Packages_Data" //
	QueueProfiles     RabbitQueue = "Steam_Profiles"      // Only takes IDs
	QueueProfilesData RabbitQueue = "Steam_Profiles_Data" //
)

var (
	consumers = map[RabbitQueue]rabbitConsumer{}

	errInvalidQueue = errors.New("invalid queue")
	errEmptyMessage = errors.New("empty message")

	consumerConnection   *amqp.Connection
	consumerCloseChannel chan *amqp.Error

	producerConnection   *amqp.Connection
	producerCloseChannel chan *amqp.Error
)

type queueInterface interface {
	getProduceQueue() RabbitQueue
	getConsumeQueue() RabbitQueue
	getRetryData() RabbitMessageDelay
	process(msg amqp.Delivery) (requeue bool)
}

func init() {

	consumerCloseChannel = make(chan *amqp.Error)
	producerCloseChannel = make(chan *amqp.Error)

	qs := []rabbitConsumer{
		{Message: RabbitMessageApp{}},
		{Message: RabbitMessageChanges{}},
		{Message: RabbitMessageDelay{}},
		{Message: RabbitMessagePackage{}},
		{Message: RabbitMessagePlayer{}},
		{Message: RabbitMessageBundle{}},
	}

	for _, v := range qs {
		consumers[v.Message.getConsumeQueue()] = v
	}
}

func RunConsumers() {
	for _, v := range consumers {
		go v.consume()
	}
}

func Produce(queue RabbitQueue, data []byte) (err error) {

	for _, v := range consumers {
		if queue == v.Message.getProduceQueue() {
			return v.produce(data)
		}
	}

	return errInvalidQueue
}

type rabbitConsumer struct {
	Message   queueInterface
	Attempt   int
	StartTime time.Time // Time first placed in delay queue
	EndTime   time.Time // Time to retry from delay queue
}

func (s rabbitConsumer) makeAConnection() (conn *amqp.Connection, err error) {

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

func (s rabbitConsumer) getQueue(conn *amqp.Connection, queue RabbitQueue) (ch *amqp.Channel, qu amqp.Queue, err error) {

	ch, err = conn.Channel()
	if err != nil {
		return
	}

	err = ch.Qos(10, 0, false)
	if err != nil {
		return
	}

	qu, err = ch.QueueDeclare(queue.String(), true, false, false, false, nil)

	return ch, qu, err
}

func (s rabbitConsumer) produce(data []byte) (err error) {

	// log.Info("Producing to: " + s.Message.getProduceQueue().String())

	// Connect
	err = func() error {

		produceLock.Lock()
		defer produceLock.Unlock()

		if producerConnection == nil {

			producerConnection, err = s.makeAConnection()
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
	ch, qu, err := s.getQueue(producerConnection, s.Message.getProduceQueue())
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

func (s rabbitConsumer) consume() {

	var err error

	for {

		// Connect
		err = func() error {

			consumeLock.Lock()
			defer consumeLock.Unlock()

			if consumerConnection == nil {

				consumerConnection, err = s.makeAConnection()
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
		ch, qu, err := s.getQueue(consumerConnection, s.Message.getConsumeQueue())
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
		func(msgs <-chan amqp.Delivery, s rabbitConsumer) {

			for {
				select {
				case err = <-consumerCloseChannel:
					log.Warning(err)
					return
				case msg := <-msgs:

					requeue := s.Message.process(msg)

					if requeue {
						logInfo("Requeuing")
						err = s.requeueMessage(msg)
						if err == nil {
							err = msg.Ack(false)
						}
					} else {
						err = msg.Ack(false)
					}

					logError(err)
				}
			}

		}(msgs, s)

		// We only get here if the amqp connection gets closed

		err = ch.Close()
		log.Err(err)
	}
}

func (s rabbitConsumer) requeueMessage(msg amqp.Delivery) error {

	delayeMessage := rabbitConsumer{
		Attempt:   s.Attempt,
		StartTime: s.StartTime,
		EndTime:   s.EndTime,
		Message: RabbitMessageDelay{
			OriginalMessage: string(msg.Body),
			OriginalQueue:   s.Message.getConsumeQueue(),
		},
	}

	delayeMessage.IncrementAttempts()

	b, err := json.Marshal(delayeMessage)
	if err != nil {
		return err
	}

	err = Produce(QueueDelaysData, b)
	log.Err(err)

	return nil
}

func (s *rabbitConsumer) IncrementAttempts() {

	// Increment attemp
	s.Attempt++

	// Update end time
	var min float64 = 1
	var max float64 = 600

	var seconds = math.Pow(1.3, float64(s.Attempt))
	var minmaxed = math.Min(min+seconds, max)
	var rounded = math.Round(minmaxed)

	s.EndTime = s.StartTime.Add(time.Second * time.Duration(rounded))
}

func handleError(err error, requeue bool) bool {

	logError(err)

	// Might be getting rate limited
	if err == steam.ErrNullResponse {
		logInfo("Null response, sleeping for 10 seconds")
		time.Sleep(time.Second * 10)
	}

	// No point in retrying if Steam has issues
	if err == steam.ErrNullResponse {
		logInfo("HTML response, sleeping for 10 seconds")
		time.Sleep(time.Second * 10)
	}

	return requeue
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
