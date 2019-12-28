package queue

import (
	"errors"
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue/framework"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/streadway/amqp"
)

const (
	QueueApps            framework.QueueName = "GDB_Apps"
	QueueAppsRegular     framework.QueueName = "GDB_Apps_Regular"
	QueueAppPlayers      framework.QueueName = "GDB_App_Players"
	QueueBundles         framework.QueueName = "GDB_Bundles"
	QueueChanges         framework.QueueName = "GDB_Changes"
	QueueGroups          framework.QueueName = "GDB_Groups"
	QueuePackages        framework.QueueName = "GDB_Packages"
	QueuePackagesRegular framework.QueueName = "GDB_Packages_Regular"
	QueuePlayers         framework.QueueName = "GDB_Players"
	QueuePlayersRegular  framework.QueueName = "GDB_Players_Regular"
	QueuePlayerRanks     framework.QueueName = "GDB_Player_Ranks"
	QueueSteam           framework.QueueName = "GDB_Steam"

	QueueDelay  framework.QueueName = "GDB_Delay"
	QueueFailed framework.QueueName = "GDB_Failed"
	QueueTest   framework.QueueName = "GDB_Test"
)

var (
	Channels = map[framework.ConnType]map[framework.QueueName]*framework.Channel{
		framework.Consumer: {},
		framework.Producer: {},
	}

	QueueDefinitions = []queue{
		{name: QueueApps, consumer: appHandler},
		{name: QueueAppsRegular},
		{name: QueueAppPlayers, consumer: appPlayersHandler},
		{name: QueueBundles, consumer: bundleHandler},
		{name: QueueChanges, consumer: changesHandler},
		{name: QueueGroups, consumer: groupsHandler},
		{name: QueuePackages, consumer: packageHandler},
		{name: QueuePackagesRegular},
		{name: QueuePlayers, consumer: playerHandler},
		{name: QueuePlayersRegular},
		{name: QueuePlayerRanks, consumer: playerRanksHandler},
		{name: QueueDelay, consumer: delayHandler, skipHeaders: true},
		{name: QueueSteam, consumer: nil},
		{name: QueueFailed, consumer: nil},
		{name: QueueTest, consumer: testHandler},
	}

	QueueSteamDefinitions = []queue{
		{name: QueueSteam, consumer: steamHandler},
		{name: QueueApps, consumer: nil},
		{name: QueuePackages, consumer: nil},
		{name: QueuePlayers, consumer: nil},
		{name: QueueChanges, consumer: nil},
	}

	QueueCronsDefinitions = []queue{
		{name: QueueApps, consumer: nil},
		{name: QueueAppPlayers, consumer: nil},
		{name: QueueGroups, consumer: nil},
		{name: QueuePlayers, consumer: nil},
		{name: QueuePlayerRanks, consumer: nil},
		{name: QueueSteam, consumer: nil},
	}
)

type queue struct {
	name        framework.QueueName
	consumer    framework.Handler
	skipHeaders bool
	batchSize   int
}

func Init(definitions []queue, consume bool) {

	heartbeat := time.Minute
	if config.IsLocal() {
		heartbeat = time.Hour
	}

	// Producers
	producerConnection, err := framework.NewConnection(config.RabbitDSN(), framework.Producer, amqp.Config{Heartbeat: heartbeat, Properties: map[string]interface{}{
		"connection_name": config.Config.Environment.Get() + "-" + string(framework.Consumer) + "-" + config.GetSteamKeyTag(),
	}})
	if err != nil {
		log.Info(err)
		return
	}

	for _, queue := range definitions {

		q, err := framework.NewChannel(producerConnection, queue.name, 10, queue.batchSize, queue.consumer, !queue.skipHeaders)
		if err != nil {
			log.Critical(string(queue.name), err)
		} else {
			Channels[framework.Producer][queue.name] = q
		}
	}

	// Consumers
	if consume {

		consumerConnection, err := framework.NewConnection(config.RabbitDSN(), framework.Consumer, amqp.Config{Heartbeat: heartbeat, Properties: map[string]interface{}{
			"connection_name": config.Config.Environment.Get() + "-" + string(framework.Consumer) + "-" + config.GetSteamKeyTag(),
		}})
		if err != nil {
			log.Info(err)
			return
		}

		for _, queue := range definitions {
			if queue.consumer != nil {

				q, err := framework.NewChannel(consumerConnection, queue.name, 10, queue.batchSize, queue.consumer, !queue.skipHeaders)
				if err != nil {
					log.Critical(string(queue.name), err)
					continue
				}

				Channels[framework.Consumer][queue.name] = q

				go q.Consume()
			}
		}
	}
}

// Message helpers
func sendToFailQueue(message *framework.Message) {
	message.SendToQueue(Channels[framework.Producer][QueueFailed])
}

func sendToRetryQueue(message *framework.Message) {
	message.SendToQueue(Channels[framework.Producer][QueueDelay])
}

func sendToLastQueue(message *framework.Message) {

	queue := message.LastQueue()

	if queue == "" {
		queue = QueueFailed
	}

	message.SendToQueue(Channels[framework.Producer][queue])
}

// Producers
func ProduceApp(payload AppMessage) (err error) {

	if !helpers.IsValidAppID(payload.ID) {
		return sql.ErrInvalidAppID
	}

	mc := memcache.GetClient()
	item := memcache.MemcacheAppInQueue(payload.ID)

	if payload.ChangeNumber == 0 {
		_, err = mc.Get(item.Key)
		if err == nil {
			return memcache.ErrInQueue
		}
	}

	err = produce(QueueApps, payload)
	if err == nil {
		err = mc.Set(&item)
	}

	return err
}

func ProduceAppRegular(payload AppMessage) (err error) {
	return produce(QueueAppsRegular, payload)
}

func ProduceAppPlayers(payload AppPlayerMessage) (err error) {

	if len(payload.IDs) == 0 {
		return nil
	}

	return produce(QueueAppPlayers, payload)
}

func ProduceBundle(id int) (err error) {

	mc := memcache.GetClient()

	item := memcache.MemcacheBundleInQueue(id)
	_, err = mc.Get(item.Key)
	if err == nil {
		return memcache.ErrInQueue
	}

	err = produce(QueueBundles, BundleMessage{ID: id})
	if err == nil {
		err = mc.Set(&item)
	}

	return err
}

func ProduceChanges(payload ChangesMessage) (err error) {
	return produce(QueueChanges, payload)
}

func ProduceGroup(payload GroupMessage) (err error) {

	if !helpers.IsValidGroupID(payload.ID) {
		return errors.New("invalid group id: " + payload.ID)
	}

	if payload.UserAgent != nil && helpers.IsBot(*payload.UserAgent) {
		return ErrIsBot
	}

	payload.ID, err = helpers.UpgradeGroupID(payload.ID)
	if err != nil {
		return err
	}

	return produce(QueueGroups, payload)
}

func ProducePackage(payload PackageMessage) (err error) {

	if !sql.IsValidPackageID(payload.ID) {
		return sql.ErrInvalidPackageID
	}

	mc := memcache.GetClient()
	item := memcache.MemcachePackageInQueue(payload.ID)

	if payload.ChangeNumber == 0 {
		_, err = mc.Get(item.Key)
		if err == nil {
			return memcache.ErrInQueue
		}
	}

	err = produce(QueuePackages, payload)
	if err == nil {
		err = mc.Set(&item)
	}

	return err
}

func ProducePackageRegular(payload PackageMessage) (err error) {
	return produce(QueuePackagesRegular, payload)
}

var ErrIsBot = errors.New("bots can't update players")

func ProducePlayer(payload PlayerMessage) (err error) {

	if !helpers.IsValidPlayerID(payload.ID) {
		return errors.New("invalid player id: " + strconv.FormatInt(payload.ID, 10))
	}

	if payload.UserAgent != nil && helpers.IsBot(*payload.UserAgent) {
		return ErrIsBot
	}

	mc := memcache.GetClient()

	item := memcache.MemcachePlayerInQueue(payload.ID)
	_, err = mc.Get(item.Key)
	if err == nil {
		return memcache.ErrInQueue
	}

	err = produce(QueuePlayers, payload)
	if err == nil {
		err = mc.Set(&item)
	}

	return err
}

func ProducePlayerRegular(id int64) (err error) {
	return produce(QueuePlayersRegular, PlayerMessage{ID: id})
}

func ProducePlayerRank(payload PlayerRanksMessage) (err error) {
	return produce(QueuePlayerRanks, payload)
}

func ProduceSteam(payload SteamMessage) (err error) {

	if len(payload.AppIDs) == 0 && len(payload.PackageIDs) == 0 {
		return nil
	}

	return produce(QueueSteam, payload)
}

func ProduceTest(id int) (err error) {
	return produce(QueueTest, TestMessage{ID: id})
}

func produce(q framework.QueueName, payload interface{}) error {

	if val, ok := Channels[framework.Producer][q]; ok {
		return val.Produce(payload)
	}

	return errors.New("channel does not exist")
}
