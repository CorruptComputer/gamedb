package mongo

import (
	"html/template"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/Jleagle/steam-go/steamid"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/steam"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RankMetric string

func (rk RankMetric) String() string {
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
	return ""
}

const (
	RankKeyLevel    RankMetric = "l"
	RankKeyBadges   RankMetric = "b"
	RankKeyFriends  RankMetric = "f"
	RankKeyComments RankMetric = "c"
	RankKeyGames    RankMetric = "g"
	RankKeyPlaytime RankMetric = "p"
)

var PlayerRankFields = map[string]RankMetric{
	"level":          RankKeyLevel,
	"games_count":    RankKeyGames,
	"badges_count":   RankKeyBadges,
	"play_time":      RankKeyPlaytime,
	"friends_count":  RankKeyFriends,
	"comments_count": RankKeyComments,
}

var PlayerRankFieldsInflux = map[RankMetric]string{
	RankKeyLevel:    "level_rank",
	RankKeyGames:    "games_rank",
	RankKeyBadges:   "badges_rank",
	RankKeyPlaytime: "playtime_rank",
	RankKeyFriends:  "friends_rank",
	RankKeyComments: "comments_rank",
}

type Player struct {
	AchievementCount         int                        `bson:"achievement_count"`      // Number of achievements
	AchievementCount100      int                        `bson:"achievement_count_100"`  // Number of 100% games
	AchievementCountApps     int                        `bson:"achievement_count_apps"` // Number of games with an achievement
	Aliases                  []string                   `bson:"aliases"`
	Avatar                   string                     `bson:"avatar"`
	BackgroundAppID          int                        `bson:"background_app_id"`
	BadgesCount              int                        `bson:"badges_count"`
	BadgeStats               ProfileBadgeStats          `bson:"badge_stats"`
	Bans                     PlayerBans                 `bson:"bans"`
	CommentsCount            int                        `bson:"comments_count"`
	CommunityVisibilityState int                        `bson:"community_visibility_state"`
	ContinentCode            string                     `bson:"continent_code"`
	CountryCode              string                     `bson:"country_code"`
	Donated                  int                        `bson:"donated"`
	FriendsCount             int                        `bson:"friends_count"`
	GamesByType              map[string]int             `bson:"games_by_type"`
	GamesCount               int                        `bson:"games_count"`
	GameStats                PlayerAppStatsTemplate     `bson:"game_stats"`
	GroupsCount              int                        `bson:"groups_count"`
	ID                       int64                      `bson:"_id"`
	LastBan                  time.Time                  `bson:"bans_last"`
	Level                    int                        `bson:"level"`
	NumberOfGameBans         int                        `bson:"bans_game"`
	NumberOfVACBans          int                        `bson:"bans_cav"`
	PersonaName              string                     `bson:"persona_name"`
	PlayTime                 int                        `bson:"play_time"`
	PlayTimeWindows          int                        `bson:"play_time_windows"`
	PlayTimeMac              int                        `bson:"play_time_mac"`
	PlayTimeLinux            int                        `bson:"play_time_linux"`
	PrimaryGroupID           string                     `bson:"primary_clan_id_string"`
	Ranks                    map[string]int             `bson:"ranks"`
	RecentAppsCount          int                        `bson:"recent_apps_count"`
	Removed                  bool                       `bson:"removed"` // Removed from Steam
	StateCode                string                     `bson:"status_code"`
	TimeCreated              time.Time                  `bson:"time_created"` // Created on Steam
	UpdatedAt                time.Time                  `bson:"updated_at"`
	VanityURL                string                     `bson:"vanity_url"`
	WishlistAppsCount        int                        `bson:"wishlist_apps_count"`
	WishlistTotalCost        map[steamapi.ProductCC]int `bson:"wishlist_total_cost"`
}

func (player Player) BSON() bson.D {

	// Stops ranks saving as null
	if player.Ranks == nil {
		player.Ranks = map[string]int{}
	}

	return bson.D{
		{"_id", player.ID},
		{"achievement_count", player.AchievementCount},
		{"achievement_count_100", player.AchievementCount100},
		{"achievement_count_apps", player.AchievementCountApps},
		{"aliases", player.Aliases},
		{"avatar", player.Avatar},
		{"background_app_id", player.BackgroundAppID},
		{"badge_stats", player.BadgeStats},
		{"bans", player.Bans},
		{"community_visibility_state", player.CommunityVisibilityState},
		{"continent_code", player.ContinentCode},
		{"country_code", player.CountryCode},
		{"donated", player.Donated},
		{"game_stats", player.GameStats},
		{"games_by_type", player.GamesByType},
		{"bans_last", player.LastBan},
		{"bans_game", player.NumberOfGameBans},
		{"bans_cav", player.NumberOfVACBans},
		{"persona_name", player.PersonaName},
		{"primary_clan_id_string", player.PrimaryGroupID},
		{"status_code", player.StateCode},
		{"time_created", player.TimeCreated},
		{"updated_at", time.Now()},
		{"vanity_url", player.VanityURL},
		{"wishlist_apps_count", player.WishlistAppsCount},
		{"wishlist_total_cost", player.WishlistTotalCost},
		{"recent_apps_count", player.RecentAppsCount},
		{"removed", player.Removed},
		{"groups_count", player.GroupsCount},
		{"ranks", player.Ranks},
		{"play_time_windows", player.PlayTimeWindows},
		{"play_time_mac", player.PlayTimeMac},
		{"play_time_linux", player.PlayTimeLinux},

		// Rank Metrics
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

	if player.TimeCreated.IsZero() || player.TimeCreated.Unix() == 0 {
		return "-"
	}
	return player.TimeCreated.Format(helpers.DateYear)
}

func (player Player) GetUpdatedUnix() int64 {
	return player.UpdatedAt.Unix()
}

func (player Player) GetUpdatedNice() string {
	return player.UpdatedAt.Format(helpers.DateTime)
}

func (player Player) GetFriendLink() template.URL {
	return template.URL("steam://friends/add/" + strconv.FormatInt(player.ID, 10))
}

func (player Player) GetMessageLink() template.URL {
	return template.URL("steam://friends/message/" + strconv.FormatInt(player.ID, 10))
}

func (player Player) CommunityLink() string {
	return helpers.GetPlayerCommunityLink(player.ID, player.VanityURL)
}

func (player Player) GetStateName() string {

	if player.CountryCode == "" || player.StateCode == "" {
		return ""
	}

	if val, ok := i18n.States[player.CountryCode][player.StateCode]; ok {
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
	return i18n.CountryCodeToName(player.CountryCode)
}

func (player Player) GetAvatar2() string {
	return helpers.GetPlayerAvatar2(player.Level)
}

func (player Player) GetPlaytimeShort(platform string, max int) (ret string) {

	switch platform {
	case "windows":
		return helpers.GetTimeShort(player.PlayTimeWindows, max)
	case "mac":
		return helpers.GetTimeShort(player.PlayTimeMac, max)
	case "linux":
		return helpers.GetTimeShort(player.PlayTimeLinux, max)
	default:
		return helpers.GetTimeShort(player.PlayTime, max)
	}
}

func (player Player) GetPlaytimePercent(platform string) (ret string) {

	total := player.PlayTimeWindows + player.PlayTimeMac + player.PlayTimeLinux

	if total == 0 {
		return "-"
	}

	var percent float64

	switch platform {
	case "windows":
		percent = float64(player.PlayTimeWindows) / float64(total)
	case "mac":
		percent = float64(player.PlayTimeMac) / float64(total)
	case "linux":
		percent = float64(player.PlayTimeLinux) / float64(total)
	}

	return helpers.FloatToString(percent*100, 2) + "%"
}

func (player Player) GetWishlistTotal(cc steamapi.ProductCC) string {

	if val, ok := player.WishlistTotalCost[cc]; ok {
		return i18n.FormatPrice(i18n.GetProdCC(cc).CurrencyCode, val)
	}

	return "-"
}

type UpdateType string

const (
	PlayerUpdateAuto   UpdateType = "auto"
	PlayerUpdateManual UpdateType = "manual"
	PlayerUpdateAdmin  UpdateType = "admin"
)

func (player Player) NeedsUpdate(updateType UpdateType) bool {

	if player.Removed {
		return false
	}

	var err error
	player.ID, err = helpers.IsValidPlayerID(player.ID)
	if err != nil {
		return false
	}

	switch updateType {
	case PlayerUpdateAdmin:
		return true
	case PlayerUpdateAuto:
		// On page requests
		if player.UpdatedAt.Add(time.Hour * 6).Before(time.Now()) {
			return true
		}
	case PlayerUpdateManual:
		// Non donators
		if player.Donated == 0 {
			if player.UpdatedAt.Add(time.Minute * 10).Before(time.Now()) {
				return true
			}
		} else {
			// Donators
			if player.UpdatedAt.Add(time.Minute * 1).Before(time.Now()) {
				return true
			}
		}
	}

	return false
}

//noinspection GoUnusedExportedFunction
func CreatePlayerIndexes() {

	var indexModels []mongo.IndexModel

	// These are for the ranking cron
	// And for players table  filtering
	for col := range PlayerRankFields {
		indexModels = append(indexModels, mongo.IndexModel{
			Keys: bson.D{{col, -1}},
		})
		indexModels = append(indexModels, mongo.IndexModel{
			Keys: bson.D{{"continent_code", 1}, {col, -1}},
		})
		indexModels = append(indexModels, mongo.IndexModel{
			Keys: bson.D{{"country_code", 1}, {col, -1}},
		})
		indexModels = append(indexModels, mongo.IndexModel{
			Keys: bson.D{{"country_code", 1}, {"status_code", 1}, {col, -1}},
		})
	}

	// For sorting main players table
	cols := []string{
		"achievement_count",
		"badges_count",
		"bans_cav",
		"bans_game",
		"bans_last",
		"comments_count",
		"friends_count",
		"games_count",
		"level",
		"play_time",
	}

	for _, col := range cols {
		indexModels = append(indexModels, mongo.IndexModel{
			Keys: bson.D{{col, -1}},
		})
	}

	// Text index
	indexModels = append(indexModels, mongo.IndexModel{
		Keys:    bson.D{{"persona_name", "text"}, {"vanity_url", "text"}},
		Options: options.Index().SetName("text").SetWeights(bson.D{{"persona_name", 1}, {"vanity_url", 1}}),
	})

	// For player search in chatbot
	indexModels = append(indexModels, mongo.IndexModel{
		Keys: bson.D{{"persona_name", 1}},
		Options: options.Index().SetCollation(&options.Collation{
			Locale:   "en",
			Strength: 2, // Case insensitive
		}),
	})
	indexModels = append(indexModels, mongo.IndexModel{
		Keys: bson.D{{"vanity_url", 1}},
		Options: options.Index().SetCollation(&options.Collation{
			Locale:   "en",
			Strength: 2, // Case insensitive
		}),
	})

	//
	client, ctx, err := getMongo()
	if err != nil {
		log.Err(err)
		return
	}

	_, err = client.Database(MongoDatabase).Collection(CollectionPlayers.String()).Indexes().CreateMany(ctx, indexModels)
	log.Err(err)
}

func GetPlayer(id int64) (player Player, err error) {

	var item = memcache.MemcachePlayer(id)

	err = memcache.GetSetInterface(item.Key, item.Expiration, &player, func() (interface{}, error) {

		id, err := helpers.IsValidPlayerID(id)
		if err != nil {
			return player, steamid.ErrInvalidPlayerID
		}

		err = FindOne(CollectionPlayers, bson.D{{"_id", id}}, nil, nil, &player)
		return player, err
	})

	player.ID = id

	return player, err
}

func SearchPlayer(search string, projection bson.M) (player Player, queue bool, err error) {

	search = strings.TrimSpace(search)

	if search == "" {
		return player, false, steamid.ErrInvalidPlayerID
	}

	//
	var ops = options.FindOne()

	// Set to case insensitive
	ops.SetCollation(&options.Collation{
		Locale:   "en",
		Strength: 2,
	})

	if projection != nil {
		ops.SetProjection(projection)
	}

	client, ctx, err := getMongo()
	if err != nil {
		return player, false, err
	}

	c := client.Database(MongoDatabase).Collection(CollectionPlayers.String())

	// Get by ID
	id, err := steamid.ParsePlayerID(search)
	if err == nil {

		err = c.FindOne(ctx, bson.D{{"_id", id}}, ops).Decode(&player)
		err = helpers.IgnoreErrors(err, ErrNoDocuments)
		if err != nil {
			log.Err(err)
		}
	}

	if player.ID == 0 {

		err = c.FindOne(ctx, bson.D{{"persona_name", search}}, ops).Decode(&player)
		err = helpers.IgnoreErrors(err, ErrNoDocuments)
		if err != nil {
			log.Err(err)
		}
	}

	// if player.ID == 0 {
	//
	// 	err = c.FindOne(ctx, bson.D{{"vanity_url", search}}, ops).Decode(&player)
	// 	err = helpers.IgnoreErrors(err, ErrNoDocuments)
	// 	if err != nil {
	// 		log.Err(err)
	// 	}
	// }

	if player.ID == 0 {

		resp, err := steam.GetSteam().ResolveVanityURL(search, steamapi.VanityURLProfile)
		if err == nil && resp.Success > 0 {

			player.ID = int64(resp.SteamID)

			var wg sync.WaitGroup
			for k := range projection {

				switch k {
				case "level":

					wg.Add(1)
					go func() {

						defer wg.Done()

						resp, err := steam.GetSteam().GetSteamLevel(player.ID)
						err = steam.AllowSteamCodes(err)
						if err != nil {
							log.Err(err)
							return
						}

						player.Level = resp
					}()

				case "persona_name", "avatar":

					wg.Add(1)
					go func() {

						defer wg.Done()

						if player.PersonaName == "" {

							summary, err := steam.GetSteam().GetPlayer(player.ID)
							if err == steamapi.ErrProfileMissing {
								return
							}
							if err = steam.AllowSteamCodes(err); err != nil {
								log.Err(err)
								return
							}

							player.PersonaName = summary.PersonaName
							player.Avatar = summary.AvatarHash
						}
					}()

				case "games_count", "play_time":

					wg.Add(1)
					go func() {

						defer wg.Done()

						if player.GamesCount == 0 {

							resp, err := steam.GetSteam().GetOwnedGames(player.ID)
							err = steam.AllowSteamCodes(err)
							if err != nil {
								log.Err(err)
								return
							}

							var playtime = 0
							for _, v := range resp.Games {
								playtime += v.PlaytimeForever
							}

							player.PlayTime = playtime
							player.GamesCount = len(resp.Games)
						}
					}()

				case "friends_count":

					wg.Add(1)
					go func() {

						defer wg.Done()

						resp, err := steam.GetSteam().GetFriendList(player.ID)
						err = steam.AllowSteamCodes(err, 401, 404)
						if err != nil {
							log.Err(err)
							return
						}

						player.FriendsCount = len(resp)
					}()
				}
			}
			wg.Wait()
		}
	}

	if player.ID == 0 {
		return player, false, mongo.ErrNoDocuments
	}

	return player, true, nil
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
		} else {
			players = append(players, player)
		}
	}

	return players, cur.Err()
}

func GetPlayerLevels() (counts []count, err error) {

	var item = memcache.MemcachePlayerLevels

	err = memcache.GetSetInterface(item.Key, item.Expiration, &counts, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return counts, err
		}

		pipeline := mongo.Pipeline{
			{{Key: "$group", Value: bson.M{"_id": "$level", "count": bson.M{"$sum": 1}}}},
		}

		cur, err := client.Database(MongoDatabase, options.Database()).Collection(CollectionPlayers.String()).Aggregate(ctx, pipeline, options.Aggregate())
		if err != nil {
			return counts, err
		}

		defer func() {
			err = cur.Close(ctx)
			log.Err(err)
		}()

		var counts []count
		for cur.Next(ctx) {

			var level count
			err := cur.Decode(&level)
			if err != nil {
				log.Err(err, level.ID)
			}
			counts = append(counts, level)
		}

		sort.Slice(counts, func(i, j int) bool {
			return counts[i].ID < counts[j].ID
		})

		return counts, cur.Err()
	})

	return counts, err
}

func GetPlayerLevelsRounded() (counts []count, err error) {

	var item = memcache.MemcachePlayerLevelsRounded

	err = memcache.GetSetInterface(item.Key, item.Expiration, &counts, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return counts, err
		}

		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.M{"level": bson.M{"$lte": 2000}}}},
			{{Key: "$group", Value: bson.M{"_id": bson.M{"$trunc": bson.A{"$level", -1}}, "count": bson.M{"$sum": 1}}}},
		}

		cur, err := client.Database(MongoDatabase, options.Database()).Collection(CollectionPlayers.String()).Aggregate(ctx, pipeline, options.Aggregate())
		if err != nil {
			return counts, err
		}

		defer func() {
			err = cur.Close(ctx)
			log.Err(err)
		}()

		var maxCount int
		var countsMap = map[int]count{}

		for cur.Next(ctx) {

			var level count
			err := cur.Decode(&level)
			if err != nil {
				log.Err(err, level.ID)
			}

			countsMap[level.ID] = level

			if level.ID > maxCount {
				maxCount = level.ID
			}
		}

		var counts []count
		for i := 0; i <= maxCount; i = i + 10 {
			if val, ok := countsMap[i]; ok {
				counts = append(counts, val)
			} else {
				counts = append(counts, count{ID: i, Count: 0})
			}
		}

		return counts, cur.Err()
	})

	return counts, err
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

func (pb PlayerBans) History() bool {
	return pb.CommunityBanned || pb.VACBanned || pb.NumberOfVACBans > 0 || pb.DaysSinceLastBan > 0 || pb.NumberOfGameBans > 0 || pb.EconomyBan != "none"
}

// PlayerAppStatsTemplate
type PlayerAppStatsTemplate struct {
	Played playerAppStatsInnerTemplate
	All    playerAppStatsInnerTemplate
}

type playerAppStatsInnerTemplate struct {
	Count     int
	Price     map[steamapi.ProductCC]int
	PriceHour map[steamapi.ProductCC]float64
	Time      int
	ProductCC steamapi.ProductCC
}

func (p *playerAppStatsInnerTemplate) AddApp(appTime int, prices map[string]int, priceHours map[string]float64) {

	p.Count++

	if p.Price == nil {
		p.Price = map[steamapi.ProductCC]int{}
	}

	if p.PriceHour == nil {
		p.PriceHour = map[steamapi.ProductCC]float64{}
	}

	for _, code := range i18n.GetProdCCs(true) {

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

	return i18n.FormatPrice(i18n.GetProdCC(p.ProductCC).CurrencyCode, int(math.Round(float64(p.Price[p.ProductCC])/float64(p.Count))), true)
}

func (p playerAppStatsInnerTemplate) GetTotalPrice() string {

	if p.Count == 0 {
		return "-"
	}

	return i18n.FormatPrice(i18n.GetProdCC(p.ProductCC).CurrencyCode, p.Price[p.ProductCC], true)
}

func (p playerAppStatsInnerTemplate) GetAveragePriceHour() string {

	if p.Count == 0 {
		return "-"
	}

	return i18n.FormatPrice(i18n.GetProdCC(p.ProductCC).CurrencyCode, int(p.PriceHour[p.ProductCC]/float64(p.Count)), true)
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
