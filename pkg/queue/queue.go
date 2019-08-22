package queue

import (
	"encoding/json"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/go-durationfmt"
	"github.com/Philipp15b/go-steam"
	"github.com/cenkalti/backoff"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type queueName string

//noinspection GoUnusedConst
const (
	queueGoApps      queueName = "GameDB_Go_Apps"
	queueGoAppPlayer queueName = "GameDB_Go_App_Players"
	queueGoBundles   queueName = "GameDB_Go_Bundles"
	queueGoChanges   queueName = "GameDB_Go_Changes"
	queueGoDelays    queueName = "GameDB_Go_Delays"
	queueGoFailed    queueName = "GameDB_Go_Failed"
	queueGoGroups    queueName = "GameDB_Go_Groups"
	queueGoGroupsNew queueName = "GameDB_Go_Groups_New"
	queueGoPackages  queueName = "GameDB_Go_Packages"
	queueGoPlayers   queueName = "GameDB_Go_Profiles"
	QueueSteam       queueName = "GameDB_Go_Steam"

	//
	maxBytesToStore int = 1024 * 10
)

var (
	consumeLock sync.Mutex
	produceLock sync.Mutex

	consumerConnection *amqp.Connection
	producerConnection *amqp.Connection

	consumerConnectionChannel = make(chan *amqp.Error)
	producerConnectionChannel = make(chan *amqp.Error)

	QueueRegister = map[queueName]baseQueue{
		queueGoApps: {
			Name:  queueGoApps,
			queue: &appQueue{},
		},
		queueGoBundles: {
			Name:  queueGoBundles,
			queue: &bundleQueue{},
		},
		queueGoChanges: {
			Name:  queueGoChanges,
			queue: &changeQueue{},
		},
		queueGoDelays: {
			Name:  queueGoDelays,
			queue: &delayQueue{},
		},
		queueGoGroups: {
			Name:  queueGoGroups,
			queue: &groupQueueScrape{},
		},
		queueGoGroupsNew: {
			Name:    queueGoGroupsNew,
			queue:   &groupQueueAPI{},
			maxTime: time.Hour * 24 * 7,
		},
		queueGoPackages: {
			Name:  queueGoPackages,
			queue: &packageQueue{},
		},
		queueGoPlayers: {
			Name:  queueGoPlayers,
			queue: &playerQueue{},
		},
		queueGoAppPlayer: {
			Name:  queueGoAppPlayer,
			queue: &appPlayerQueue{},
		},
		QueueSteam: {
			Name:       QueueSteam,
			queue:      &steamQueue{},
			DoNotScale: true,
		},
	}
)

type baseMessage struct {
	Message       interface{} `json:"message"`
	FirstSeen     time.Time   `json:"first_seen"`
	LastSeen      time.Time   `json:"last_seen"`
	Attempt       int         `json:"attempt"`
	OriginalQueue queueName   `json:"original_queue"`
	actionTaken   bool        `json:"-"`
	sync.Mutex    `json:"-"`
}

func (payload baseMessage) getNextAttempt() time.Time {

	var min = time.Second * 2
	var max = time.Hour

	var seconds float64
	seconds = math.Pow(1.5, float64(payload.Attempt))
	seconds = math.Max(seconds, min.Seconds())
	seconds = math.Min(seconds, max.Seconds())

	return payload.LastSeen.Add(time.Second * time.Duration(int64(seconds)))
}

// Remove from queue
func (payload *baseMessage) ack(msg amqp.Delivery) {

	payload.Lock()
	defer payload.Unlock()

	if payload.actionTaken {
		return
	}
	payload.actionTaken = true

	err := msg.Ack(false)
	logError(err)
}

func (payload *baseMessage) ackMulti(msg amqp.Delivery) {

	payload.Lock()
	defer payload.Unlock()

	if payload.actionTaken {
		return
	}
	payload.actionTaken = true

	err := msg.Ack(true)
	logError(err)
}

// Send to failed queue
func (payload *baseMessage) fail(msg amqp.Delivery) {

	payload.Lock()
	defer payload.Unlock()

	if payload.actionTaken {
		return
	}
	payload.actionTaken = true

	err := produce(*payload, queueGoFailed)
	if err != nil {
		logError(err)
		return
	}

	err = msg.Ack(false)
	if err != nil {
		logError(err)
		return
	}
}

// Send to delay queue
func (payload *baseMessage) ackRetry(msg amqp.Delivery) {

	payload.Lock()
	defer payload.Unlock()

	if payload.actionTaken {
		return
	}
	payload.actionTaken = true

	totalStr, err := durationfmt.Format(payload.getNextAttempt().Sub(payload.FirstSeen), "%mm %ss")
	if err != nil {
		logError(err)
	}

	leftStr, err := durationfmt.Format(payload.getNextAttempt().Sub(time.Now()), "%mm %ss")
	if err != nil {
		logError(err)
	}

	logInfo("Adding to delay queue for " + leftStr + ", " + totalStr + " total, attempt " + strconv.Itoa(payload.Attempt))

	err = produce(*payload, queueGoDelays)
	if err != nil {
		logError(err)
		return
	}

	err = msg.Ack(false)
	if err != nil {
		logError(err)
		return
	}
}

type queueInterface interface {
	processMessages(msgs []amqp.Delivery)
}

type baseQueue struct {
	queue       queueInterface
	Name        queueName
	DoNotScale  bool
	SteamClient *steam.Client // Just used for Steam queue
	qos         int
	batchSize   int
	maxAttempts int
	maxTime     time.Duration
}

func (q baseQueue) getQOS() int {

	if q.qos != 0 {
		return q.qos
	}

	return 10
}

func (q baseQueue) getMaxTime() time.Duration {

	if q.maxTime != 0 {
		return q.maxTime
	}

	return time.Hour * 24 * 7
}

func (q baseQueue) ConsumeMessages() {

	var err error

	for {

		func() {

			// Connect
			err = func() error {

				consumeLock.Lock()
				defer consumeLock.Unlock()

				if consumerConnection == nil {

					log.Info("Getting new consumer connection")

					consumerConnection, err = getConnection()
					if err != nil {
						return err
					}
					consumerConnection.NotifyClose(consumerConnectionChannel)
				}

				return nil
			}()

			if err != nil {
				logCritical("Connecting to Rabbit: " + err.Error())
				return
			}

			//
			ch, qu, err := getQueue(consumerConnection, q.Name, q.getQOS())
			if err != nil {
				logError(err)
				return
			}

			defer func(ch *amqp.Channel) {
				err = ch.Close()
				logError(err)
			}(ch)

			tag := config.Config.Environment.Get() + "-" + config.GetSteamKeyTag()

			msgs, err := ch.Consume(qu.Name, tag, false, false, false, false, nil)
			if err != nil {
				logError(err)
				return
			}

			// In a anon function so can return at anytime
			func(msgs <-chan amqp.Delivery, q baseQueue) {

				var msgSlice []amqp.Delivery

				for {
					select {
					case err = <-consumerConnectionChannel:
						logWarning("Consumer connection closed", err)
						return
					case msg := <-msgs:
						msgSlice = append(msgSlice, msg)
					}

					if len(msgSlice) >= q.batchSize {
						q.queue.processMessages(msgSlice)
						msgSlice = []amqp.Delivery{}
					}
				}

			}(msgs, q)

			logWarning("Rabbit consumer connection has disconnected")

		}()
	}
}

func produce(payload baseMessage, queue queueName) (err error) {

	if payload.OriginalQueue == "" {
		payload.OriginalQueue = queue
	}

	if payload.FirstSeen.IsZero() {
		payload.FirstSeen = time.Now()
	}

	if queue != queueGoDelays && queue != queueGoFailed {
		payload.LastSeen = time.Now()
		payload.Attempt++
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Connect
	err = func() error {

		produceLock.Lock()
		defer produceLock.Unlock()

		if producerConnection == nil {

			log.Info("Getting new producer connection")

			producerConnection, err = getConnection()
			if err != nil {
				logCritical("Connecting to Rabbit: " + err.Error())
				return err
			}
			producerConnection.NotifyClose(producerConnectionChannel)
		}

		return nil
	}()

	if err != nil {
		return err
	}

	//
	ch, qu, err := getQueue(producerConnection, queue, QueueRegister[queue].getQOS())
	if err != nil {
		return err
	}

	defer func() {
		err = ch.Close()
		log.Err(err)
	}()

	return ch.Publish("", qu.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         b,
	})
}

func getConnection() (conn *amqp.Connection, err error) {

	operation := func() (err error) {

		amqpConfig := amqp.Config{}
		if config.IsLocal() {
			amqpConfig.Heartbeat = time.Hour
		}
		conn, err = amqp.DialConfig(config.RabbitDSN(), amqpConfig)

		return err
	}

	policy := backoff.NewExponentialBackOff()
	policy.MaxElapsedTime = 0
	policy.InitialInterval = 5 * time.Second

	err = backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { logInfo(err) })

	return conn, err
}

func getQueue(conn *amqp.Connection, queue queueName, qos int) (ch *amqp.Channel, qu amqp.Queue, err error) {

	ch, err = conn.Channel()
	if err != nil {
		return
	}

	err = ch.Qos(qos, 0, false)
	if err != nil {
		return
	}

	qu, err = ch.QueueDeclare(string(queue), true, false, false, false, nil)

	return ch, qu, err
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

func logCritical(interfaces ...interface{}) {
	log.Critical(append(interfaces, log.LogNameConsumers)...)
}
