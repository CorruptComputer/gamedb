package mongo

import (
	"errors"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var CountriesWithStates = []string{"AU", "CA", "FR", "GB", "NZ", "PH", "SI", "US"}

type RankKey int

func (rk RankKey) String() string {
	switch rk {
	case RankKeyLevel:
		return "Level"
	case RankKeyBadges:
		return "Badges"
	case RankKeyFriends:
		return "Friends"
	case RankKeyComments:
		return "Comments"
	case RankKeyGames:
		return "Games"
	case RankKeyPlaytime:
		return "Playtime"
	}
	return "xx"
}

const (
	RankKeyLevel    RankKey = 1
	RankKeyBadges   RankKey = 2
	RankKeyFriends  RankKey = 3
	RankKeyComments RankKey = 4
	RankKeyGames    RankKey = 5
	RankKeyPlaytime RankKey = 6
)

const (
	RankCountryAll  = "ALL"
	RankCountryNone = "NONE"
)

var (
	ErrInvalidPlayerID   = errors.New("invalid player id")
	ErrInvalidPlayerName = errors.New("invalid player name")
)

type Player struct {
	ID                  int64          `bson:"_id"`                    //
	Avatar              string         `bson:"avatar"`                 //
	BackgroundAppID     int            `bson:"background_app_id"`      //
	BadgeIDs            []int          `bson:"badge_ids"`              // []int - Only special badges
	BadgeStats          string         `bson:"badge_stats"`            // ProfileBadgeStats
	Bans                string         `bson:"bans"`                   // PlayerBans
	CountryCode         string         `bson:"country_code"`           //
	Donated             int            `bson:"donated"`                //
	GameStats           string         `bson:"game_stats"`             // PlayerAppStatsTemplate
	GamesByType         map[string]int `bson:"games_by_type"`          //
	Ranks               map[string]int `bson:"ranks"`                  //
	LastLogOff          time.Time      `bson:"time_logged_off"`        //
	LastBan             time.Time      `bson:"bans_last"`              //
	NumberOfGameBans    int            `bson:"bans_game"`              //
	NumberOfVACBans     int            `bson:"bans_cav"`               //
	PersonaName         string         `bson:"persona_name"`           //
	PrimaryClanIDString string         `bson:"primary_clan_id_string"` //
	StateCode           string         `bson:"status_code"`            //
	TimeCreated         time.Time      `bson:"time_created"`           //
	UpdatedAt           time.Time      `bson:"updated_at"`             //
	VanintyURL          string         `bson:"vanity_url"`             //
	WishlistAppsCount   int            `bson:"wishlist_apps_count"`    //
	RecentAppsCount     int            `bson:"recent_apps_count"`      //
	GroupsCount         int            `bson:"groups_count"`           //
	CommentsCount       int            `bson:"comments_count"`         //

	// Ranked
	BadgesCount  int `bson:"badges_count"`
	FriendsCount int `bson:"friends_count"`
	GamesCount   int `bson:"games_count"`
	Level        int `bson:"level"`
	PlayTime     int `bson:"play_time"`
}

func (player Player) BSON() bson.D {

	return bson.D{
		{"_id", player.ID},
		{"avatar", player.Avatar},
		{"background_app_id", player.BackgroundAppID},
		{"badge_ids", player.BadgeIDs},
		{"badge_stats", player.BadgeStats},
		{"bans", player.Bans},
		{"country_code", player.CountryCode},
		{"donated", player.Donated},
		{"game_stats", player.GameStats},
		{"games_by_type", player.GamesByType},
		{"time_logged_off", player.LastLogOff},
		{"bans_last", player.LastBan},
		{"bans_game", player.NumberOfGameBans},
		{"bans_cav", player.NumberOfVACBans},
		{"persona_name", player.PersonaName},
		{"primary_clan_id_string", player.PrimaryClanIDString},
		{"status_code", player.StateCode},
		{"time_created", player.TimeCreated},
		{"updated_at", time.Now()},
		{"vanity_url", player.VanintyURL},
		{"wishlist_apps_count", player.WishlistAppsCount},
		{"recent_apps_count", player.RecentAppsCount},
		{"groups_count", player.GroupsCount},
		{"ranks", player.Ranks},

		// Ranked
		{"badges_count", player.BadgesCount},
		{"friends_count", player.FriendsCount},
		{"games_count", player.GamesCount},
		{"level", player.Level},
		{"play_time", player.PlayTime},
		{"comments_count", player.CommentsCount},
	}
}

func (player Player) GetPath() string {
	return helpers.GetPlayerPath(player.ID, player.GetName())
}

func (player Player) GetName() string {
	return helpers.GetPlayerName(player.ID, player.PersonaName)
}

func (player Player) GetSteamTimeUnix() int64 {
	return player.TimeCreated.Unix()
}

func (player Player) GetSteamTimeNice() string {
	return player.TimeCreated.Format(helpers.DateYear)
}

func (player Player) GetLogoffUnix() int64 {
	return player.LastLogOff.Unix()
}

func (player Player) GetLogoffNice() string {
	return player.LastLogOff.Format(helpers.DateYearTime)
}

func (player Player) GetUpdatedUnix() int64 {
	return player.UpdatedAt.Unix()
}

func (player Player) GetUpdatedNice() string {
	return player.UpdatedAt.Format(helpers.DateTime)
}

func (player Player) CommunityLink() string {

	if player.VanintyURL != "" {
		return "https://steamcommunity.com/id/" + player.VanintyURL
	}

	return "https://steamcommunity.com/profiles/" + strconv.FormatInt(player.ID, 10)
}

func (player Player) GetStateName() string {

	if player.CountryCode == "" || player.StateCode == "" {
		return ""
	}

	if val, ok := helpers.States[player.CountryCode][player.StateCode]; ok {
		return val
	}

	return player.StateCode
}

func (player Player) GetMaxFriends() int {
	return helpers.GetPlayerMaxFriends(player.Level)
}

func (player Player) GetAvatar() string {
	return helpers.GetPlayerAvatar(player.Avatar)
}

func (player Player) GetFlag() string {
	return helpers.GetPlayerFlagPath(player.CountryCode)
}

func (player Player) GetCountry() string {
	return helpers.CountryCodeToName(player.CountryCode)
}

func (player Player) GetBadgeStats() (stats ProfileBadgeStats, err error) {

	err = helpers.Unmarshal([]byte(player.BadgeStats), &stats)
	return stats, err
}

func (player Player) GetAvatar2() string {
	return helpers.GetPlayerAvatar2(player.Level)
}

func (player Player) GetTimeShort() (ret string) {
	return helpers.GetTimeShort(player.PlayTime, 2)
}

func (player Player) GetTimeLong() (ret string) {
	return helpers.GetTimeLong(player.PlayTime, 5)
}

//
func (player Player) GetSpecialBadges() (badges []PlayerBadge) {

	if player.BadgeIDs == nil || len(player.BadgeIDs) == 0 {
		return
	}

	for _, v := range player.BadgeIDs {

		if val, ok := GlobalBadges[v]; ok {
			badges = append(badges, val)
		}
	}

	sort.Slice(badges, func(i, j int) bool {
		return badges[i].GetUniqueID() > badges[j].GetUniqueID()
	})

	return badges
}

func (player Player) GetBans() (bans PlayerBans, err error) {

	err = helpers.Unmarshal([]byte(player.Bans), &bans)
	return bans, err
}

func (player Player) GetGameStats(code steam.ProductCC) (stats PlayerAppStatsTemplate, err error) {

	err = helpers.Unmarshal([]byte(player.GameStats), &stats)

	stats.All.ProductCC = code
	stats.Played.ProductCC = code

	return stats, err
}

func (player Player) GetRank(metric RankKey, cc string) (i int, found bool) {

	if val, ok := player.Ranks[strconv.Itoa(int(metric))+"_"+cc]; ok {
		return val, true
	}

	return 0, false
}

type UpdateType string

const (
	PlayerUpdateAuto   UpdateType = "auto"
	PlayerUpdateManual UpdateType = "manual"
	PlayerUpdateAdmin  UpdateType = "admin"
)

func (player Player) NeedsUpdate(updateType UpdateType) bool {

	if !helpers.IsValidPlayerID(player.ID) {
		return false
	}

	switch updateType {
	case PlayerUpdateAdmin:
		return true
	case PlayerUpdateAuto:
		if player.UpdatedAt.Add(time.Hour*24*7).Unix() < time.Now().Unix() { // 1 week
			return true
		}
	case PlayerUpdateManual:
		if player.Donated == 0 {
			if player.UpdatedAt.Add(time.Hour*24).Unix() < time.Now().Unix() { // 1 day
				return true
			}
		} else {
			if player.UpdatedAt.Add(time.Hour*1).Unix() < time.Now().Unix() { // 1 hour
				return true
			}
		}
	}

	return false
}

func GetPlayer(id int64) (player Player, err error) {

	var item = helpers.MemcachePlayer(id)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &player, func() (interface{}, error) {

		if !helpers.IsValidPlayerID(id) {
			return player, ErrInvalidPlayerID
		}

		err = FindOne(CollectionPlayers, bson.D{{"_id", id}}, nil, nil, &player)
		if err != nil {
			return player, err
		}
		if player.ID == 0 {
			return player, ErrNoDocuments
		}

		return player, err
	})

	player.ID = id

	return player, err
}

func SearchPlayer(s string, projection bson.M) (player Player, err error) {

	if s == "" {
		return player, ErrInvalidPlayerID
	}

	client, ctx, err := getMongo()
	if err != nil {
		return player, err
	}

	var filter bson.M

	i, _ := strconv.ParseInt(s, 10, 64)
	if helpers.IsValidPlayerID(i) {
		filter = bson.M{"_id": s}
	} else {
		filter = bson.M{"$text": bson.M{"$search": s}}
	}

	ops := options.FindOne()
	if projection != nil {
		ops.SetProjection(projection)
	}

	c := client.Database(MongoDatabase).Collection(CollectionPlayers.String())
	result := c.FindOne(ctx, filter, ops)

	err = result.Decode(&player)
	return player, err
}

func GetPlayersByID(ids []int64, projection bson.M) (players []Player, err error) {

	if len(ids) < 1 {
		return players, nil
	}

	var idsBSON bson.A
	for _, v := range ids {
		idsBSON = append(idsBSON, v)
	}

	return GetPlayers(0, 0, nil, bson.D{{"_id", bson.M{"$in": idsBSON}}}, projection)
}

func GetPlayers(offset int64, limit int64, sort bson.D, filter bson.D, projection bson.M) (players []Player, err error) {

	cur, ctx, err := Find(CollectionPlayers, offset, limit, sort, filter, projection, nil)
	if err != nil {
		return players, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var player Player
		err := cur.Decode(&player)
		if err != nil {
			log.Err(err, player.ID)
		}
		players = append(players, player)
	}

	return players, cur.Err()
}

func GetUniquePlayerCountries() (codes []string, err error) {

	var item = helpers.MemcacheUniquePlayerCountryCodes

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &codes, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return codes, err
		}

		c := client.Database(MongoDatabase, options.Database()).Collection(CollectionPlayers.String())

		resp, err := c.Distinct(ctx, "country_code", bson.M{}, options.Distinct())
		if err != nil {
			return codes, err
		}

		for _, v := range resp {
			if code, ok := v.(string); ok {
				codes = append(codes, code)
			}
		}

		return codes, err
	})

	return codes, err
}

func GetUniquePlayerStates(country string) (codes []helpers.Tuple, err error) {

	var item = helpers.MemcacheUniquePlayerStateCodes(country)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &codes, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return codes, err
		}

		c := client.Database(MongoDatabase, options.Database()).Collection(CollectionPlayers.String())

		resp, err := c.Distinct(ctx, "status_code", bson.M{"country_code": country}, options.Distinct())
		if err != nil {
			return codes, err
		}

		for _, v := range resp {
			if stateCode, ok := v.(string); stateCode != "" && ok {

				name := stateCode
				if val, ok := helpers.States[country][stateCode]; ok {
					name = val
				}

				codes = append(codes, helpers.Tuple{Key: stateCode, Value: name})
			}
		}

		sort.Slice(codes, func(i, j int) bool {
			return codes[i].Value < codes[j].Value
		})

		return codes, err
	})

	return codes, err
}

func CountPlayers() (count int64, err error) {

	var item = helpers.MemcachePlayersCount

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		return CountDocuments(CollectionPlayers, nil, 0)
	})

	return count, err
}

func BulkUpdatePlayers(writes []mongo.WriteModel) (err error) {

	if len(writes) == 0 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	c := client.Database(MongoDatabase).Collection(CollectionPlayers.String())

	_, err = c.BulkWrite(ctx, writes, options.BulkWrite().SetOrdered(false))
	return err
}

// ProfileBadgeStats
type ProfileBadgeStats struct {
	PlayerXP                   int
	PlayerLevel                int
	PlayerXPNeededToLevelUp    int
	PlayerXPNeededCurrentLevel int
	PercentOfLevel             int
}

// PlayerBans
type PlayerBans struct {
	CommunityBanned  bool   `json:"community_banned"`
	VACBanned        bool   `json:"vac_banned"`
	NumberOfVACBans  int    `json:"number_of_vac_bans"`
	DaysSinceLastBan int    `json:"days_since_last_ban"`
	NumberOfGameBans int    `json:"number_of_game_bans"`
	EconomyBan       string `json:"economy_ban"`
}

// PlayerAppStatsTemplate
type PlayerAppStatsTemplate struct {
	Played playerAppStatsInnerTemplate
	All    playerAppStatsInnerTemplate
}

type playerAppStatsInnerTemplate struct {
	Count     int
	Price     map[steam.ProductCC]int
	PriceHour map[steam.ProductCC]float64
	Time      int
	ProductCC steam.ProductCC
}

func (p *playerAppStatsInnerTemplate) AddApp(appTime int, prices map[string]int, priceHours map[string]float64) {

	p.Count++

	if p.Price == nil {
		p.Price = map[steam.ProductCC]int{}
	}

	if p.PriceHour == nil {
		p.PriceHour = map[steam.ProductCC]float64{}
	}

	for _, code := range helpers.GetProdCCs(true) {

		// Sometimes priceHour can be -1, meaning infinite
		var priceHour = priceHours[string(code.ProductCode)]
		if priceHour < 0 {
			priceHour = 0
		}

		p.Price[code.ProductCode] = p.Price[code.ProductCode] + prices[string(code.ProductCode)]
		p.PriceHour[code.ProductCode] = p.PriceHour[code.ProductCode] + priceHour
		p.Time = p.Time + appTime
	}
}

func (p playerAppStatsInnerTemplate) GetAveragePrice() string {

	if p.Count == 0 {
		return "-"
	}

	return helpers.FormatPrice(helpers.GetProdCC(p.ProductCC).CurrencyCode, int(math.Round(float64(p.Price[p.ProductCC])/float64(p.Count))), true)
}

func (p playerAppStatsInnerTemplate) GetTotalPrice() string {

	if p.Count == 0 {
		return "-"
	}

	return helpers.FormatPrice(helpers.GetProdCC(p.ProductCC).CurrencyCode, p.Price[p.ProductCC], true)
}

func (p playerAppStatsInnerTemplate) GetAveragePriceHour() string {

	if p.Count == 0 {
		return "-"
	}

	return helpers.FormatPrice(helpers.GetProdCC(p.ProductCC).CurrencyCode, int(p.PriceHour[p.ProductCC]/float64(p.Count)), true)
}

func (p playerAppStatsInnerTemplate) GetAverageTime() string {

	if p.Count == 0 {
		return "-"
	}

	return helpers.GetTimeShort(int(float64(p.Time)/float64(p.Count)), 2)
}

func (p playerAppStatsInnerTemplate) GetTotalTime() string {

	if p.Count == 0 {
		return "-"
	}

	return helpers.GetTimeShort(p.Time, 2)
}
