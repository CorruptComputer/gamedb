package mongo

import (
	"math"
	"strconv"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PlayerApp struct {
	PlayerID     int64              `bson:"player_id"`
	AppID        int                `bson:"app_id"`
	AppName      string             `bson:"app_name"`
	AppIcon      string             `bson:"app_icon"`
	AppTime      int                `bson:"app_time"`
	AppPrices    map[string]int     `bson:"app_prices"`
	AppPriceHour map[string]float64 `bson:"app_prices_hour"`

	AppAchievementsTotal   int     `bson:"app_achievements_total"`
	AppAchievementsHave    int     `bson:"app_achievements_have"`
	AppAchievementsPercent float64 `bson:"app_achievements_percent"`
}

func (app PlayerApp) BSON() bson.D {

	var prices = bson.M{}
	for k, v := range app.AppPrices {
		prices[k] = v
	}

	var pricesHour = bson.M{}
	for k, v := range app.AppPriceHour {
		pricesHour[k] = v
	}

	return bson.D{
		{"_id", app.GetKey()},
		{"player_id", app.PlayerID},
		{"app_id", app.AppID},
		{"app_name", app.AppName},
		{"app_icon", app.AppIcon},
		{"app_time", app.AppTime},
		{"app_prices", prices},
		{"app_prices_hour", pricesHour},

		{"app_achievements_total", app.AppAchievementsTotal},
		{"app_achievements_have", app.AppAchievementsHave},
		{"app_achievements_percent", app.AppAchievementsPercent},
	}
}

func (app PlayerApp) GetKey() string {
	return strconv.FormatInt(app.PlayerID, 10) + "-" + strconv.Itoa(app.AppID)
}

func (app PlayerApp) GetPath() string {
	return helpers.GetAppPath(app.AppID, app.AppName)
}

func (app PlayerApp) GetIcon() string {

	if app.AppIcon == "" {
		return helpers.DefaultPlayerAvatar
	}
	return helpers.AppIconBase + strconv.Itoa(app.AppID) + "/" + app.AppIcon + ".jpg"
}

func (app PlayerApp) GetTimeNice() string {

	return helpers.GetTimeShort(app.AppTime, 2)
}

func (app PlayerApp) GetPriceFormatted(code steamapi.ProductCC) string {

	val, ok := app.AppPrices[string(code)]
	if ok {
		return i18n.FormatPrice(i18n.GetProdCC(code).CurrencyCode, val)
	} else {
		return "-"
	}
}

func (app PlayerApp) GetPriceHourFormatted(code steamapi.ProductCC) string {

	val, ok := app.AppPriceHour[string(code)]
	if ok {
		if val < 0 {
			return "∞"
		}
		return i18n.FormatPrice(i18n.GetProdCC(code).CurrencyCode, int(math.Round(float64(val))))
	} else {
		return "-"
	}
}

func (app PlayerApp) GetPercent() float64 {
	return math.Round(app.AppAchievementsPercent)
}

func GetPlayerAppsByApp(offset int64, filter bson.D) (apps []PlayerApp, err error) {

	return getPlayerApps(offset, 100, filter, bson.D{{"app_time", -1}}, bson.M{"_id": -1, "player_id": 1, "app_time": 1})
}

func GetPlayerAppByKey(playerID int64, appID int) (playerApp PlayerApp, err error) {

	playerApp.PlayerID = playerID
	playerApp.AppID = appID

	err = FindOne(CollectionPlayerApps, bson.D{{"_id", playerApp.GetKey()}}, nil, nil, &playerApp)

	return playerApp, err
}

func GetPlayerApps(offset int64, limit int64, filter bson.D, sort bson.D) (apps []PlayerApp, err error) {

	return getPlayerApps(offset, limit, filter, sort, nil)
}

func GetPlayersApps(playerIDs []int64, projection bson.M) (apps []PlayerApp, err error) {

	if len(playerIDs) < 1 {
		return apps, err
	}

	playersFilter := bson.A{}
	for _, v := range playerIDs {
		playersFilter = append(playersFilter, v)
	}

	return getPlayerApps(0, 0, bson.D{{"player_id", bson.M{"$in": playersFilter}}}, nil, projection)
}

func GetAppPlayTimes(appID int) ([]PlayerApp, error) {

	return getPlayerApps(0, 0, bson.D{{"app_id", appID}}, nil, bson.M{"_id": -1, "app_time": 1})
}

func GetAppOwners(appID int) ([]PlayerApp, error) {

	return getPlayerApps(0, 0, bson.D{{"app_id", appID}}, nil, bson.M{"_id": -1, "player_id": 1})
}

func getPlayerApps(offset int64, limit int64, filter bson.D, sort bson.D, projection bson.M) (apps []PlayerApp, err error) {

	cur, ctx, err := Find(CollectionPlayerApps, offset, limit, sort, filter, projection, nil)
	if err != nil {
		return apps, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var playerApp PlayerApp
		err := cur.Decode(&playerApp)
		if err != nil {
			log.Err(err, playerApp.GetKey())
		} else {
			apps = append(apps, playerApp)
		}
	}

	return apps, cur.Err()
}

func UpdatePlayerApps(apps map[int]*PlayerApp) (err error) {

	if len(apps) == 0 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	var writes []mongo.WriteModel
	for _, v := range apps {

		write := mongo.NewReplaceOneModel()
		write.SetFilter(bson.M{"_id": v.GetKey()})
		write.SetReplacement(v.BSON())
		write.SetUpsert(true)

		writes = append(writes, write)
	}

	c := client.Database(MongoDatabase).Collection(CollectionPlayerApps.String())

	_, err = c.BulkWrite(ctx, writes, options.BulkWrite())

	return err
}
