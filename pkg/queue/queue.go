package queue

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/steam-go/steamid"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type QueueMessageInterface interface {
	Queue() rabbit.QueueName
}

const (
	// Apps
	QueueApps                   rabbit.QueueName = "GDB_Apps"
	QueueAppsAchievements       rabbit.QueueName = "GDB_Apps.Achievements"
	QueueAppsItems              rabbit.QueueName = "GDB_Apps.Items"
	QueueAppsArticlesSearch     rabbit.QueueName = "GDB_Apps.Articles.Search"
	QueueAppsAchievementsSearch rabbit.QueueName = "GDB_Apps.Achievements.Search"
	QueueAppsYoutube            rabbit.QueueName = "GDB_Apps.Youtube"
	QueueAppsWishlists          rabbit.QueueName = "GDB_Apps.Wishlists"
	QueueAppsInflux             rabbit.QueueName = "GDB_Apps.Influx"
	QueueAppsDLC                rabbit.QueueName = "GDB_Apps.DLC"
	QueueAppsSameowners         rabbit.QueueName = "GDB_Apps.Sameowners"
	QueueAppsNews               rabbit.QueueName = "GDB_Apps.News"
	QueueAppsFindGroup          rabbit.QueueName = "GDB_Apps.FindGroup"
	QueueAppsReviews            rabbit.QueueName = "GDB_Apps.Reviews"
	QueueAppsTwitch             rabbit.QueueName = "GDB_Apps.Twitch"
	QueueAppsMorelike           rabbit.QueueName = "GDB_Apps.Morelike"
	QueueAppsSteamspy           rabbit.QueueName = "GDB_Apps.Steamspy"
	QueueAppsSearch             rabbit.QueueName = "GDB_Apps.Search"

	// Packages
	QueuePackages       rabbit.QueueName = "GDB_Packages"
	QueuePackagesPrices rabbit.QueueName = "GDB_Packages.Prices"

	// Players
	QueuePlayers             rabbit.QueueName = "GDB_Players"
	QueuePlayersAchievements rabbit.QueueName = "GDB_Players.Achievements"
	QueuePlayersBadges       rabbit.QueueName = "GDB_Players.Badges"
	QueuePlayersSearch       rabbit.QueueName = "GDB_Players.Search"
	QueuePlayersGames        rabbit.QueueName = "GDB_Players.Games"
	QueuePlayersAliases      rabbit.QueueName = "GDB_Players.Aliases"
	QueuePlayersGroups       rabbit.QueueName = "GDB_Players.Groups"

	// Group
	QueueGroups          rabbit.QueueName = "GDB_Groups"
	QueueGroupsSearch    rabbit.QueueName = "GDB_Groups.Search"
	QueueGroupsPrimaries rabbit.QueueName = "GDB_Groups.Primaries"

	// App players
	QueueAppPlayers    rabbit.QueueName = "GDB_App_Players"
	QueueAppPlayersTop rabbit.QueueName = "GDB_App_Players_Top"

	// Other
	QueueBundles     rabbit.QueueName = "GDB_Bundles"
	QueueChanges     rabbit.QueueName = "GDB_Changes"
	QueueDelay       rabbit.QueueName = "GDB_Delay"
	QueueFailed      rabbit.QueueName = "GDB_Failed"
	QueuePlayerRanks rabbit.QueueName = "GDB_Player_Ranks"
	QueueSteam       rabbit.QueueName = "GDB_Steam"
	QueueTest        rabbit.QueueName = "GDB_Test"
	QueueWebsockets  rabbit.QueueName = "GDB_Websockets"
)

var (
	ProducerChannels = map[rabbit.QueueName]*rabbit.Channel{}

	AllProducerDefinitions = []QueueDefinition{
		{name: QueueAppPlayers},
		{name: QueueAppPlayersTop},
		{name: QueueApps},
		{name: QueueAppsDLC},
		{name: QueueAppsYoutube},
		{name: QueueAppsInflux},
		{name: QueueAppsNews},
		{name: QueueAppsFindGroup},
		{name: QueueAppsWishlists, prefetchSize: 1000},
		{name: QueueAppsItems},
		{name: QueueAppsAchievements},
		{name: QueueAppsAchievementsSearch},
		{name: QueueAppsArticlesSearch},
		{name: QueueAppsSameowners},
		{name: QueueAppsReviews},
		{name: QueueAppsMorelike},
		{name: QueueAppsTwitch},
		{name: QueueAppsSteamspy},
		{name: QueueBundles},
		{name: QueueChanges},
		{name: QueueGroups},
		{name: QueueGroupsPrimaries, prefetchSize: 1000},
		{name: QueueGroupsSearch, prefetchSize: 1000},
		{name: QueuePackages},
		{name: QueuePackagesPrices},
		{name: QueuePlayers},
		{name: QueuePlayerRanks},
		{name: QueuePlayersAchievements},
		{name: QueuePlayersBadges},
		{name: QueuePlayersGames},
		{name: QueueDelay, skipHeaders: true},
		{name: QueueAppsSearch, prefetchSize: 1000},
		{name: QueuePlayersSearch, prefetchSize: 1000},
		{name: QueuePlayersAliases},
		{name: QueuePlayersGroups},
		{name: QueueSteam},
		{name: QueueFailed, skipHeaders: true},
		{name: QueueTest},
		{name: QueueWebsockets},
	}

	ConsumersDefinitions = []QueueDefinition{
		{name: QueueAppPlayers, consumer: appPlayersHandler},
		{name: QueueAppPlayersTop, consumer: appPlayersHandler},
		{name: QueueApps, consumer: appHandler},
		{name: QueueAppsInflux, consumer: appInfluxHandler},
		{name: QueueAppsDLC, consumer: appDLCHandler},
		{name: QueueAppsYoutube, consumer: appYoutubeHandler},
		{name: QueueAppsNews, consumer: appNewsHandler},
		{name: QueueAppsWishlists, consumer: appWishlistsHandler, prefetchSize: 1000},
		{name: QueueAppsFindGroup, consumer: appsFindGroupHandler},
		{name: QueueAppsAchievements, consumer: appAchievementsHandler},
		{name: QueueAppsAchievementsSearch, consumer: appsAchievementsSearchHandler},
		{name: QueueAppsArticlesSearch, consumer: appsArticlesSearchHandler},
		{name: QueueAppsItems, consumer: appItemsHandler},
		{name: QueueAppsSameowners, consumer: appSameownersHandler},
		{name: QueueAppsReviews, consumer: appReviewsHandler},
		{name: QueueAppsMorelike, consumer: appMorelikeHandler},
		{name: QueueAppsTwitch, consumer: appTwitchHandler},
		{name: QueueAppsSteamspy, consumer: appSteamspyHandler},
		{name: QueueBundles, consumer: bundleHandler},
		{name: QueueChanges, consumer: changesHandler},
		{name: QueueGroups, consumer: groupsHandler},
		{name: QueueGroupsSearch, consumer: groupsSearchHandler, prefetchSize: 1000},
		{name: QueueGroupsPrimaries, consumer: groupPrimariesHandler, prefetchSize: 1000},
		{name: QueuePackages, consumer: packageHandler},
		{name: QueuePackagesPrices, consumer: packagePriceHandler},
		{name: QueuePlayers, consumer: playerHandler},
		{name: QueuePlayersAliases, consumer: playerAliasesHandler},
		{name: QueuePlayersGroups, consumer: playersGroupsHandler},
		{name: QueuePlayerRanks, consumer: playerRanksHandler},
		{name: QueuePlayersAchievements, consumer: playerAchievementsHandler},
		{name: QueuePlayersGames, consumer: playerGamesHandler},
		{name: QueuePlayersBadges, consumer: playerBadgesHandler},
		{name: QueueDelay, consumer: delayHandler, skipHeaders: true},
		{name: QueueAppsSearch, consumer: appsSearchHandler, prefetchSize: 1000},
		{name: QueuePlayersSearch, consumer: appsPlayersHandler, prefetchSize: 1000},
		{name: QueueSteam},
		{name: QueueFailed, skipHeaders: true},
		{name: QueueTest, consumer: testHandler},
		{name: QueueWebsockets},
	}

	WebserverDefinitions = []QueueDefinition{
		{name: QueueApps},
		{name: QueueAppsAchievements},
		{name: QueueAppsAchievementsSearch},
		{name: QueueAppsArticlesSearch},
		{name: QueueAppsYoutube},
		{name: QueueAppsInflux},
		{name: QueueAppsWishlists, prefetchSize: 1000},
		{name: QueueAppPlayers},
		{name: QueueAppPlayersTop},
		{name: QueueAppsReviews},
		{name: QueueAppsSearch, prefetchSize: 1000},
		{name: QueueBundles},
		{name: QueueChanges},
		{name: QueueGroups},
		{name: QueueGroupsSearch, prefetchSize: 1000},
		{name: QueueGroupsPrimaries, prefetchSize: 1000},
		{name: QueuePackages},
		{name: QueuePackagesPrices},
		{name: QueuePlayers},
		{name: QueuePlayersGroups},
		{name: QueuePlayersSearch, prefetchSize: 1000},
		{name: QueuePlayerRanks},
		{name: QueueDelay, skipHeaders: true},
		{name: QueueSteam},
		{name: QueueFailed, skipHeaders: true},
		{name: QueueTest},
		{name: QueueWebsockets, consumer: websocketHandler},
	}

	QueueSteamDefinitions = []QueueDefinition{
		{name: QueueSteam, consumer: steamHandler},
		{name: QueueApps},
		{name: QueuePackages},
		{name: QueuePlayers},
		{name: QueueChanges},
		{name: QueueDelay, skipHeaders: true},
	}

	QueueCronsDefinitions = []QueueDefinition{
		{name: QueueApps},
		{name: QueueAppsAchievements},
		{name: QueueAppsAchievementsSearch},
		{name: QueueAppsInflux},
		{name: QueueAppsWishlists, prefetchSize: 1000},
		{name: QueueAppsSearch, prefetchSize: 1000},
		{name: QueueAppsYoutube},
		{name: QueueAppsReviews},
		{name: QueueAppPlayers},
		{name: QueueAppPlayersTop},
		{name: QueueGroups},
		{name: QueueGroupsSearch, prefetchSize: 1000},
		{name: QueueGroupsPrimaries, prefetchSize: 1000},
		{name: QueuePackages},
		{name: QueuePlayers},
		{name: QueuePlayersGroups},
		{name: QueuePlayersSearch, prefetchSize: 1000},
		{name: QueuePlayerRanks},
		{name: QueueSteam},
		{name: QueueDelay, skipHeaders: true},
		{name: QueueWebsockets},
	}

	ChatbotDefinitions = []QueueDefinition{
		{name: QueuePlayers},
		{name: QueueWebsockets},
	}
)

var discordClient *discordgo.Session

func SetDiscordClient(c *discordgo.Session) {
	discordClient = c
}

type QueueDefinition struct {
	name         rabbit.QueueName
	consumer     rabbit.Handler
	skipHeaders  bool
	prefetchSize int
}

func Init(definitions []QueueDefinition) {

	heartbeat := time.Minute
	if config.IsLocal() {
		heartbeat = time.Hour
	}

	// Producers
	c := rabbit.ConnectionConfig{
		Address:  config.RabbitDSN(),
		ConnType: rabbit.Producer,
		Config: amqp.Config{
			Heartbeat: heartbeat,
			Properties: map[string]interface{}{
				"connection_name": config.Config.Environment.Get() + "-" + string(rabbit.Consumer) + "-" + config.GetSteamKeyTag(),
			},
		},
		LogInfo: func(i ...interface{}) {
			// zap.S().Info(i...)
		},
		LogError: func(i ...interface{}) {
			zap.S().Error(i...)
		},
	}

	producerConnection, err := rabbit.NewConnection(c)
	if err != nil {
		zap.S().Info(err)
		return
	}

	var consume bool

	for _, queue := range definitions {

		if queue.consumer != nil {
			consume = true
		}

		prefetchSize := 50
		if queue.prefetchSize > 0 {
			prefetchSize = queue.prefetchSize
		}

		q, err := rabbit.NewChannel(producerConnection, queue.name, config.Config.Environment.Get(), prefetchSize, queue.consumer, !queue.skipHeaders)
		if err != nil {
			zap.S().Fatal(string(queue.name), err)
		} else {
			ProducerChannels[queue.name] = q
		}
	}

	// Consumers
	if consume {

		c = rabbit.ConnectionConfig{
			Address:  config.RabbitDSN(),
			ConnType: rabbit.Consumer,
			Config: amqp.Config{
				Heartbeat: heartbeat,
				Properties: map[string]interface{}{
					"connection_name": config.Config.Environment.Get() + "-" + string(rabbit.Consumer) + "-" + config.GetSteamKeyTag(),
				},
			},
			LogInfo: func(i ...interface{}) {
				// zap.S().Info(i...)
			},
			LogError: func(i ...interface{}) {
				zap.S().Error(i...)
			},
		}

		consumerConnection, err := rabbit.NewConnection(c)
		if err != nil {
			zap.S().Info(err)
			return
		}

		for _, queue := range definitions {
			if queue.consumer != nil {

				prefetchSize := 50
				if queue.prefetchSize > 0 {
					prefetchSize = queue.prefetchSize
				}

				for k := range make([]int, 2) {

					q, err := rabbit.NewChannel(consumerConnection, queue.name, config.Config.Environment.Get()+"-"+strconv.Itoa(k), prefetchSize, queue.consumer, !queue.skipHeaders)
					if err != nil {
						zap.S().Fatal(string(queue.name), err)
						continue
					}

					go q.Consume()
				}
			}
		}
	}
}

// Message helpers
func sendToFailQueue(message *rabbit.Message) {

	err := message.SendToQueueAndAck(ProducerChannels[QueueFailed], nil)
	if err != nil {
		zap.S().Error(err)
	}
}

func sendToRetryQueue(message *rabbit.Message) {

	sendToRetryQueueWithDelay(message, 0)
}

func sendToRetryQueueWithDelay(message *rabbit.Message, delay time.Duration) {

	var po rabbit.ProduceOptions
	if delay > 0 {
		po = func(p amqp.Publishing) amqp.Publishing {
			p.Headers["delay-until"] = time.Now().Add(delay).Unix()
			return p
		}
	}

	err := message.SendToQueueAndAck(ProducerChannels[QueueDelay], po)
	if err != nil {
		zap.S().Error(err)
	}
}

func sendToLastQueue(message *rabbit.Message) {

	queue := message.LastQueue()

	if queue == "" {
		queue = QueueFailed
	}

	err := message.SendToQueueAndAck(ProducerChannels[queue], nil)
	if err != nil {
		zap.S().Error(err)
	}
}

// Producers
func ProduceApp(payload AppMessage) (err error) {

	if !helpers.IsValidAppID(payload.ID) {
		return mongo.ErrInvalidAppID
	}

	item := memcache.MemcacheAppInQueue(payload.ID)

	if payload.ChangeNumber == 0 {
		_, err = memcache.Get(item.Key)
		if err == nil {
			return memcache.ErrInQueue
		}
	}

	err = produce(QueueApps, payload)
	if err == nil {
		err = memcache.Set(item.Key, item.Value, item.Expiration)
	}

	return err
}

func ProduceAppsInflux(appIDs []int) (err error) {
	m := AppInfluxMessage{AppIDs: appIDs}
	return produce(m.Queue(), m)
}

func ProduceAppsReviews(id int) (err error) {
	m := AppReviewsMessage{AppID: id}
	return produce(m.Queue(), m)
}

func ProduceAppsYoutube(id int, name string) (err error) {
	return produce(QueueAppsYoutube, AppYoutubeMessage{ID: id, Name: name})
}

func ProduceAppsWishlists(id int) (err error) {
	return produce(QueueAppsWishlists, AppWishlistsMessage{AppID: id})
}

func ProduceAppPlayers(appIDs []int) (err error) {

	if len(appIDs) == 0 {
		return nil
	}

	return produce(QueueAppPlayers, AppPlayerMessage{IDs: appIDs})
}

func ProduceAppPlayersTop(appIDs []int) (err error) {

	if len(appIDs) == 0 {
		return nil
	}

	return produce(QueueAppPlayersTop, AppPlayerMessage{IDs: appIDs})
}

func ProduceBundle(id int) (err error) {

	item := memcache.MemcacheBundleInQueue(id)

	_, err = memcache.Get(item.Key)
	if err == nil {
		return memcache.ErrInQueue
	}

	err = produce(QueueBundles, BundleMessage{ID: id})
	if err == nil {
		err = memcache.Set(item.Key, item.Value, item.Expiration)
	}

	return err
}

func ProduceChanges(payload ChangesMessage) (err error) {

	return produce(QueueChanges, payload)
}

func ProduceDLC(appID int, DLCIDs []int) (err error) {

	return produce(QueueAppsDLC, DLCMessage{AppID: appID, DLCIDs: DLCIDs})
}

func ProducePlayerAchievements(playerID int64, appID int, force bool) (err error) {

	return produce(QueuePlayersAchievements, PlayerAchievementsMessage{PlayerID: playerID, AppID: appID, Force: force})
}

func ProduceGroup(payload GroupMessage) (err error) {

	if payload.UserAgent != nil && helpers.IsBot(*payload.UserAgent) {
		return ErrIsBot
	}

	item := memcache.MemcacheGroupInQueue(payload.ID)

	_, err = memcache.Get(item.Key)
	if err == nil {
		return memcache.ErrInQueue
	}

	err = produce(QueueGroups, payload)
	if err == nil {
		err = memcache.Set(item.Key, item.Value, item.Expiration)
	}

	return err
}

func ProducePackage(payload PackageMessage) (err error) {

	if !helpers.IsValidPackageID(payload.ID) {
		return mongo.ErrInvalidPackageID
	}

	item := memcache.MemcachePackageInQueue(payload.ID)

	if payload.ChangeNumber == 0 {
		_, err = memcache.Get(item.Key)
		if err == nil {
			return memcache.ErrInQueue
		}
	}

	err = produce(QueuePackages, payload)
	if err == nil {
		err = memcache.Set(item.Key, item.Value, item.Expiration)
	}

	return err
}

func producePackagePrice(payload PackagePriceMessage) (err error) {
	return produce(QueuePackagesPrices, payload)
}

var ErrIsBot = errors.New("bots can't update players")

func ProducePlayer(payload PlayerMessage) (err error) {

	if payload.UserAgent != nil && helpers.IsBot(*payload.UserAgent) {
		return ErrIsBot
	}

	payload.ID, err = helpers.IsValidPlayerID(payload.ID)
	if err != nil {
		return steamid.ErrInvalidPlayerID
	}

	item := memcache.MemcachePlayerInQueue(payload.ID)

	_, err = memcache.Get(item.Key)
	if err == nil {
		return memcache.ErrInQueue
	}

	err = produce(QueuePlayers, payload)
	if err == nil {
		err = memcache.Set(item.Key, item.Value, item.Expiration)
	}

	return err
}

func ProducePlayerRank(payload PlayerRanksMessage) (err error) {

	return produce(QueuePlayerRanks, payload)
}

func ProduceGroupSearch(group *mongo.Group, groupID string, groupType string) (err error) {

	return produce(QueueGroupsSearch, GroupSearchMessage{Group: group, GroupID: groupID, GroupType: groupType})
}

func ProduceGroupPrimaries(groupID string, groupType string, prims int) (err error) {

	m := GroupPrimariesMessage{GroupID: groupID, GroupType: groupType, CurrentPrimaries: prims}
	return produce(m.Queue(), m)
}

func ProduceAchievementSearch(achievement mongo.AppAchievement, appName string, appOwners int64) (err error) {

	return produce(QueueAppsAchievementsSearch, AppsAchievementsSearchMessage{
		AppAchievement: achievement,
		AppName:        appName,
		AppOwners:      appOwners,
	})
}

func ProduceArticlesSearch(payload AppsArticlesSearchMessage) (err error) {

	return produce(QueueAppsArticlesSearch, payload)
}

//goland:noinspection GoUnusedExportedFunction
func ProducePlayerAlias(id int64) (err error) {

	return produce(QueuePlayersAliases, PlayersAliasesMessage{PlayerID: id})
}

func ProduceAppAchievement(appID int, appName string, appOwners int64) (err error) {

	return produce(QueueAppsAchievements, AppAchievementsMessage{AppID: appID, AppName: appName, AppOwners: appOwners})
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

func ProducePlayerGroup(player mongo.Player, skipGroupUpdate bool, force bool) (err error) {

	return produce(QueuePlayersGroups, PlayersGroupsMessage{
		Player:                    player,
		SkipGroupUpdate:           skipGroupUpdate,
		ForceResavingPlayerGroups: force,
	})
}

func ProduceAppSearch(app *mongo.App, appID int) (err error) {

	m := AppsSearchMessage{App: app, AppID: appID}
	return produce(m.Queue(), m)
}

func ProducePlayerSearch(player *mongo.Player, playerID int64) (err error) {

	return produce(QueuePlayersSearch, PlayersSearchMessage{Player: player, PlayerID: playerID})
}

func ProduceWebsocket(payload interface{}, pages ...websockets.WebsocketPage) (err error) {

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return produce(QueueWebsockets, WebsocketMessage{
		Pages:   pages,
		Message: b,
	})
}

func produce(q rabbit.QueueName, payload interface{}) error {

	if !config.IsLocal() {
		time.Sleep(time.Second / 1000)
	}

	if val, ok := ProducerChannels[q]; ok {
		return val.Produce(payload, nil)
	}

	return errors.New("channel not in register")
}
