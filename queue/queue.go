package queue

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

type RabbitQueue string

func (rq RabbitQueue) String() string {
	return string(rq)
}

const (
	QueueApps         RabbitQueue = "Steam_Apps"          // Only takes IDs
	QueueAppsData     RabbitQueue = "Steam_Apps_Data"     //
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

	rabbitDSN string

	consumerConnection   *amqp.Connection
	consumerCloseChannel chan *amqp.Error

	producerConnection   *amqp.Connection
	producerCloseChannel chan *amqp.Error
)

type queueInterface interface {
	getProduceQueue() RabbitQueue
	getConsumeQueue() RabbitQueue
	getRetryData() RabbitMessageDelay
	process(msg amqp.Delivery) (requeue bool, err error)
}

func init() {

	consumerCloseChannel = make(chan *amqp.Error)
	producerCloseChannel = make(chan *amqp.Error)

	qs := []rabbitConsumer{
		{Message: RabbitMessageApp{}},
		{Message: RabbitMessageChanges{}},
		//{Message: RabbitMessageDelay{}},
		{Message: RabbitMessagePackage{}},
		{Message: RabbitMessageProfile{}},
	}

	for _, v := range qs {
		consumers[v.Message.getConsumeQueue()] = v
	}
}

func Init() {

	user := viper.GetString("RABBIT_USER")
	pass := viper.GetString("RABBIT_PASS")
	host := viper.GetString("RABBIT_HOST")
	port := viper.GetString("RABBIT_PORT")

	rabbitDSN = "amqp://" + user + ":" + pass + "@" + host + ":" + port
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

func (s rabbitConsumer) getQueue(conn *amqp.Connection, queue RabbitQueue) (ch *amqp.Channel, qu amqp.Queue, err error) {

	ch, err = conn.Channel()
	log.Log(err)

	err = ch.Qos(10, 0, true)
	log.Log(err)

	qu, err = ch.QueueDeclare(queue.String(), true, false, false, false, nil)
	log.Log(err)

	return ch, qu, err
}

func (s rabbitConsumer) produce(data []byte) (err error) {

	log.Info("Producing to: " + s.Message.getProduceQueue().String())

	// Connect
	if producerConnection == nil {

		producerConnection, err = amqp.Dial(rabbitDSN)
		if err != nil {
			log.Critical("Connecting to Rabbit: " + err.Error())
			return err
		}
		producerConnection.NotifyClose(producerCloseChannel)
	}

	//
	ch, qu, err := s.getQueue(producerConnection, s.Message.getProduceQueue())
	if err != nil {
		return err
	}

	// Close read
	if ch != nil {
		defer func(ch *amqp.Channel) {
			err := ch.Close()
			log.Log(err)
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
		if consumerConnection == nil {

			consumerConnection, err = amqp.Dial(rabbitDSN)
			if err != nil {
				log.Critical("Connecting to Rabbit: " + err.Error())
				return
			}
			consumerConnection.NotifyClose(consumerCloseChannel)
		}

		//
		ch, qu, err := s.getQueue(consumerConnection, s.Message.getConsumeQueue())
		if err != nil {
			log.Log(err)
			return
		}

		msgs, err := ch.Consume(qu.Name, "", false, false, false, false, nil)
		if err != nil {
			log.Log(err)
			return
		}

		// In a anon function so can return at anytime
		func(msgs <-chan amqp.Delivery, s rabbitConsumer) {

			for {
				select {
				case err = <-consumerCloseChannel:
					log.Log(err)
					return
				case msg := <-msgs:

					requeue, err := s.Message.process(msg)
					if err != nil {
						logError(err)
					}

					// Might be getting rate limited
					if err == steam.ErrNullResponse {
						logInfo("Null response, sleeping for 10 seconds")
						time.Sleep(time.Second * 10)
					}

					if requeue {
						logInfo("Requeuing")
						err = s.requeueMessage(msg)
						logError(err)
					}

					err = msg.Ack(false)
					logError(err)
				}
			}

		}(msgs, s)

		// We only get here if the amqp connection gets closed

		err = ch.Close()
		log.Log(err)
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

	data, err := json.Marshal(delayeMessage)
	if err != nil {
		return err
	}

	err = Produce(QueueDelaysData, data)
	log.Log(err)

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

type SteamKitJob struct {
	SequentialCount int    `json:"SequentialCount"`
	StartTime       string `json:"StartTime"`
	ProcessID       int    `json:"ProcessID"`
	BoxID           int    `json:"BoxID"`
	Value           int64  `json:"Value"`
}

func QueueApp(IDs []int) (err error) {

	b, err := json.Marshal(produceAppPayload{
		ID:   IDs,
		Time: time.Now().Unix(),
	})

	return Produce(QueueApps, b)
}

// JSON must match the Updater app
type produceAppPayload struct {
	ID   []int `json:"IDs"`
	Time int64 `json:"Time"`
}

func QueuePackage(IDs []int) (err error) {

	b, err := json.Marshal(producePackagePayload{
		ID:   IDs,
		Time: time.Now().Unix(),
	})

	return Produce(QueuePackages, b)
}

// JSON must match the Updater app
type producePackagePayload struct {
	ID   []int `json:"IDs"`
	Time int64 `json:"Time"`
}

func QueuePlayer(r *http.Request, player db.Player, updateType db.UpdateType) (err error) {

	err = player.ShouldUpdate(r, updateType)
	if err == nil {

		b, err := json.Marshal(producePlayerPayload{
			ID:   player.PlayerID,
			Time: time.Now().Unix(),
		})

		err = Produce(QueueProfiles, b)
		if err == nil {

			err = helpers.GetMemcache().SetItem(helpers.MemcachePlayerRefreshed(player.PlayerID))
		}
	}

	return err
}

// JSON must match the Updater app
type producePlayerPayload struct {
	ID   int64 `json:"ID"`
	Time int64 `json:"Time"`
}

func logInfo(interfaces ...interface{}) {
	log.Info(append(interfaces, log.LogNameConsumers)...)
}

func logError(interfaces ...interface{}) {
	log.Log(append(interfaces, log.LogNameConsumers)...)
}

func logWarning(interfaces ...interface{}) {
	log.Warning(append(interfaces, log.LogNameConsumers)...)
}
