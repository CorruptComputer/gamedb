package mongo

import (
	"strconv"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
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
	AppPriceHour map[string]float32 `bson:"app_prices_hour"`
}

func (pa PlayerApp) BSON() (ret interface{}) {

	var prices = bson.M{}
	for k, v := range pa.AppPrices {
		prices[k] = v
	}

	var pricesHour = bson.M{}
	for k, v := range pa.AppPriceHour {
		pricesHour[k] = v
	}

	return bson.M{
		"_id":             pa.getKey(),
		"player_id":       pa.PlayerID,
		"app_id":          pa.AppID,
		"app_name":        pa.AppName,
		"app_icon":        pa.AppIcon,
		"app_time":        pa.AppTime,
		"app_prices":      prices,
		"app_prices_hour": pricesHour,
	}
}

func (pa PlayerApp) getKey() string {
	return strconv.FormatInt(pa.PlayerID, 10) + "-" + strconv.Itoa(pa.AppID)
}

func (pa PlayerApp) GetPath() string {
	return helpers.GetAppPath(pa.AppID, pa.AppName)
}

func (pa PlayerApp) GetIcon() string {

	if pa.AppIcon == "" {
		return "/assets/img/no-player-image.jpg"
	}
	return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(pa.AppID) + "/" + pa.AppIcon + ".jpg"
}

func (pa PlayerApp) GetTimeNice() string {

	return helpers.GetTimeShort(pa.AppTime, 2)
}

func (pa PlayerApp) GetPriceFormatted(code steam.CountryCode) string {

	val, ok := pa.AppPrices[string(code)]
	if ok {

		locale, err := helpers.GetLocaleFromCountry(code)
		log.Err(err)
		return locale.Format(val)

	} else {
		return ""
	}
}

func (pa PlayerApp) GetPriceHourFormatted(code steam.CountryCode) string {

	val, ok := pa.AppPriceHour[string(code)]
	if ok {

		if val < 0 {
			return "∞"
		}

		locale, err := helpers.GetLocaleFromCountry(code)
		log.Err(err)
		return locale.FormatFloat(float64(val))

	} else {
		return ""
	}
}

func (pa PlayerApp) OutputForJSON(code steam.CountryCode) (output []interface{}) {

	return []interface{}{
		pa.AppID,
		pa.AppName,
		pa.GetIcon(),
		pa.AppTime,
		pa.GetTimeNice(),
		pa.GetPriceFormatted(code),
		pa.GetPriceHourFormatted(code),
		pa.GetPath(),
	}
}

func GetPlayerAppsByPlayers(playerIDs []int64) (apps []PlayerApp, err error) {

	playersFilter := bson.A{}
	for _, v := range playerIDs {
		playersFilter = append(playersFilter, v)
	}

	return getPlayerApps(0, 0, bson.M{"$or": playersFilter}, nil)
}

func GetPlayerAppsByPlayer(playerID int64, offset int64, limit bool, sort D) (apps []PlayerApp, err error) {

	return getPlayerApps(offset, 100, bson.M{"player_id": playerID}, sort)
}

func getPlayerApps(offset int64, limit int64, filter interface{}, sort D) (apps []PlayerApp, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return apps, err
	}

	ops := options.Find().SetSort(sort)
	if offset > 0 {
		ops.SetSkip(offset)
	}
	if limit > 0 {
		ops.SetLimit(limit)
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionPlayerApps.String())
	cur, err := c.Find(ctx, filter, ops)
	if err != nil {
		return apps, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var app PlayerApp
		err := cur.Decode(&app)
		log.Err(err)
		apps = append(apps, app)
	}

	return apps, cur.Err()
}

func UpdatePlayerApps(apps map[int]*PlayerApp) (err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	var writes []mongo.WriteModel
	for _, v := range apps {

		write := mongo.NewReplaceOneModel()
		write.SetFilter(bson.M{"_id": v.getKey()})
		write.SetReplacement(v.BSON())
		write.SetUpsert(true)

		writes = append(writes, write)
	}

	c := client.Database(MongoDatabase).Collection(CollectionPlayerApps.String())

	_, err = c.BulkWrite(ctx, writes, options.BulkWrite())
	log.Err(err)

	return err
}
