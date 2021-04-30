package mongo

import (
	"html/template"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/Jleagle/steam-go/steamid"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	PrivacyState            helpers.Bits = 1 << iota // Friends list, badges, Steam Level, showcases, comments, and group membership.
	PrivacyStateInventory                            // Items you've received in games that use Steam Trading. It also includes any Steam Trading Cards you've collected and extra copies or Steam Gifts.
	PrivacyStateGifts                                // Always keep Steam Gifts private even if users can see my inventory.
	PrivacyStateOwnedGames                           // The list of all games on your Steam account, games you’ve wishlisted, your achievements and your playtime.
	PrivacyStatePlaytime                             // Always keep my total playtime private even if users can see my game details.
	PrivacyStateFriendsList                          // This controls who can see your list of friends on your Steam Community profile.
)

type Player struct {
	AchievementCount         int                        `bson:"achievement_count"`      // Number of achievements
	AchievementCount100      int                        `bson:"achievement_count_100"`  // Number of 100% games
	AchievementCountApps     int                        `bson:"achievement_count_apps"` // Number of games with an achievement
	Avatar                   string                     `bson:"avatar"`
	BackgroundAppID          int                        `bson:"background_app_id"`
	BadgesCount              int                        `bson:"badges_count"`
	BadgesFoilCount          int                        `bson:"badges_foil_count"`
	BadgeStats               ProfileBadgeStats          `bson:"badge_stats"`
	Bans                     PlayerBans                 `bson:"bans"`
	CommentsCount            int                        `bson:"comments_count"`
	CommunityVisibilityState int                        `bson:"community_visibility_state"` // 1=private 3=public
	ContinentCode            string                     `bson:"continent_code"`
	CountryCode              string                     `bson:"country_code"`
	CreatedAt                time.Time                  `bson:"created_at"` // Added late
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
	Permissions              helpers.Bits               `bson:"permissions"`
	PersonaName              string                     `bson:"persona_name"`
	PlayTime                 int                        `bson:"play_time"`
	PlayTimeWindows          int                        `bson:"play_time_windows"`
	PlayTimeMac              int                        `bson:"play_time_mac"`
	PlayTimeLinux            int                        `bson:"play_time_linux"`
	PrimaryGroupID           string                     `bson:"primary_clan_id_string"`
	Private                  bool                       `bson:"private"`
	Ranks                    map[helpers.RankMetric]int `bson:"ranks"`
	RecentAppsCount          int                        `bson:"recent_apps_count"`
	Removed                  bool                       `bson:"removed"` // Removed from Steam
	AwardsGivenCount         int                        `bson:"awards_given_count"`
	AwardsGivenPoints        int                        `bson:"awards_given_points"`
	AwardsReceivedCount      int                        `bson:"awards_received_count"`
	AwardsReceivedPoints     int                        `bson:"awards_received_points"`
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
		player.Ranks = map[helpers.RankMetric]int{}
	}

	player.UpdatedAt = time.Now()

	// we now only set this on new players
	// if player.CreatedAt.IsZero() || player.CreatedAt.Unix() == 0 {
	// 	player.CreatedAt = time.Now()
	// }

	return bson.D{
		{"_id", player.ID},
		{"achievement_count", player.AchievementCount},
		{"achievement_count_100", player.AchievementCount100},
		{"achievement_count_apps", player.AchievementCountApps},
		{"avatar", player.Avatar},
		{"background_app_id", player.BackgroundAppID},
		{"badge_stats", player.BadgeStats},
		{"bans", player.Bans},
		{"community_visibility_state", player.CommunityVisibilityState},
		{"continent_code", player.ContinentCode},
		{"country_code", player.CountryCode},
		{"created_at", player.CreatedAt},
		{"donated", player.Donated},
		{"game_stats", player.GameStats},
		{"games_by_type", player.GamesByType},
		{"bans_last", player.LastBan},
		{"bans_game", player.NumberOfGameBans},
		{"bans_cav", player.NumberOfVACBans},
		{"persona_name", player.PersonaName},
		{"primary_clan_id_string", player.PrimaryGroupID},
		{"private", player.Private},
		{"status_code", player.StateCode},
		{"time_created", player.TimeCreated},
		{"updated_at", player.UpdatedAt},
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
		{"friends_count", player.FriendsCount},
		{"comments_count", player.CommentsCount},

		// Rank Metrics
		{"badges_count", player.BadgesCount},
		{"badges_foil_count", player.BadgesFoilCount},
		{"games_count", player.GamesCount},
		{"level", player.Level},
		{"play_time", player.PlayTime},
		{"awards_given_count", player.AwardsGivenCount},
		{"awards_given_points", player.AwardsGivenPoints},
		{"awards_received_count", player.AwardsReceivedCount},
		{"awards_received_points", player.AwardsReceivedPoints},
	}
}

func (player Player) GetPath() string {
	return helpers.GetPlayerPath(player.ID, player.GetName())
}

func (player Player) GetPathAbsolute() string {
	return helpers.GetPlayerPathAbsolute(player.ID, player.GetName())
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
	return helpers.GetPlayerCommunityLink(player.ID)
}

func (player Player) WishlistLink() string {
	return "https://store.steampowered.com/wishlist/profiles/" + strconv.FormatInt(player.ID, 10)
}

func (player Player) FriendsLink() string {
	return "https://steamcommunity.com/profiles/" + strconv.FormatInt(player.ID, 10) + "/friends"
}

func (player Player) GroupsLink() string {
	return "https://steamcommunity.com/profiles/" + strconv.FormatInt(player.ID, 10) + "/groups"
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

func (player Player) GetAvatarAbsolute() string {
	return helpers.GetPlayerAvatarAbsolute(player.Avatar)
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

func (player Player) GetGamesCount() int {
	return player.GamesCount
}

func (player Player) GetAchievements() int {
	return player.AchievementCount
}

func (player Player) GetPlaytime() int {
	return player.PlayTime
}

func (player Player) GetLevel() int {
	return player.Level
}

func (player Player) GetBadges() int {
	return player.BadgesCount
}

func (player Player) GetBadgesFoil() int {
	return player.BadgesFoilCount
}

func (player Player) GetRanks() map[helpers.RankMetric]int {
	return player.Ranks
}

func (player Player) GetVACBans() int {
	return player.NumberOfVACBans
}

func (player Player) GetGameBans() int {
	return player.NumberOfGameBans
}

func (player Player) GetLastBan() time.Time {
	return player.LastBan
}

func ensurePlayerIndexes() {

	var indexModels []mongo.IndexModel

	// These are for the ranking cron
	for col := range helpers.PlayerRankFields {
		indexModels = append(indexModels, mongo.IndexModel{
			Keys: bson.D{
				{col, -1},
			},
		})
		indexModels = append(indexModels, mongo.IndexModel{
			Keys: bson.D{
				{"continent_code", 1},
				{col, -1},
			},
		})
		indexModels = append(indexModels, mongo.IndexModel{
			Keys: bson.D{
				{"country_code", 1},
				{col, -1},
			},
		})
		indexModels = append(indexModels, mongo.IndexModel{
			Keys: bson.D{
				{"country_code", 1},
				{"status_code", 1},
				{col, -1},
			},
		})
	}

	// For last updated task
	indexModels = append(indexModels, mongo.IndexModel{
		Keys: bson.D{
			{"community_visibility_state", 1},
			{"removed", 1},
			{"updated_at", 1},
		},
	})

	// For admin stats
	indexModels = append(indexModels,
		mongo.IndexModel{Keys: bson.D{{"community_visibility_state", 1}}},
		mongo.IndexModel{Keys: bson.D{{"removed", 1}}},
	)

	// Misc
	indexModels = append(indexModels,

		// Asc
		mongo.IndexModel{Keys: bson.D{{"primary_clan_id_string", 1}}},

		// Desc
		mongo.IndexModel{Keys: bson.D{{"achievement_count_100", -1}}},
		mongo.IndexModel{Keys: bson.D{{"bans_cav", -1}}},
		mongo.IndexModel{Keys: bson.D{{"bans_game", -1}}},
		mongo.IndexModel{Keys: bson.D{{"created_at", -1}}},
	)

	//
	client, ctx, err := getMongo()
	if err != nil {
		log.ErrS(err)
		return
	}

	_, err = client.Database(config.C.MongoDatabase).Collection(CollectionPlayers.String()).Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		log.ErrS(err)
	}
}

func GetPlayer(id int64) (player Player, err error) {

	item := memcache.ItemPlayer(id)
	err = memcache.Client().GetSet(item.Key, item.Expiration, &player, func() (interface{}, error) {

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

	cur, ctx, err := find(CollectionPlayers, offset, limit, filter, sort, projection, nil)
	if err != nil {
		return players, err
	}

	defer closeCursor(cur, ctx)

	for cur.Next(ctx) {

		var player Player
		err := cur.Decode(&player)
		if err != nil {
			log.ErrS(err, player.ID)
		} else {
			players = append(players, player)
		}
	}

	return players, cur.Err()
}

func GetPlayerLevels() (counts []Count, err error) {

	item := memcache.ItemPlayerLevels
	err = memcache.Client().GetSet(item.Key, item.Expiration, &counts, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return counts, err
		}

		pipeline := mongo.Pipeline{
			{{Key: "$sort", Value: bson.M{"level": 1}}}, // Just here to hit the index
			{{Key: "$group", Value: bson.M{"_id": "$level", "count": bson.M{"$sum": 1}}}},
		}

		cur, err := client.Database(config.C.MongoDatabase, options.Database()).Collection(CollectionPlayers.String()).Aggregate(ctx, pipeline, options.Aggregate())
		if err != nil {
			return counts, err
		}

		defer closeCursor(cur, ctx)

		var counts []Count
		for cur.Next(ctx) {

			var count Count
			err := cur.Decode(&count)
			if err != nil {
				log.ErrS(err, count.ID)
			}
			counts = append(counts, count)
		}

		sort.Slice(counts, func(i, j int) bool {
			return counts[i].ID < counts[j].ID
		})

		return counts, cur.Err()
	})

	return counts, err
}

func GetPlayerUpdateDays() (counts []DateCount, err error) {

	item := memcache.ItemPlayerUpdateDates
	err = memcache.Client().GetSet(item.Key, item.Expiration, &counts, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return counts, err
		}

		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: helpers.LastUpdatedQuery}},
			{{Key: "$project", Value: bson.M{"yearMonthDayUTC": bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$updated_at"}}}}},
			{{Key: "$group", Value: bson.M{"_id": "$yearMonthDayUTC", "count": bson.M{"$sum": 1}}}},
		}

		cur, err := client.Database(config.C.MongoDatabase, options.Database()).Collection(CollectionPlayers.String()).Aggregate(ctx, pipeline, options.Aggregate())
		if err != nil {
			return counts, err
		}

		defer closeCursor(cur, ctx)

		var counts []DateCount
		for cur.Next(ctx) {

			var count DateCount
			err := cur.Decode(&count)
			if err != nil {
				log.ErrS(err, count.Date)
			}
			counts = append(counts, count)
		}

		sort.Slice(counts, func(i, j int) bool {
			return counts[i].Date < counts[j].Date
		})

		return counts, cur.Err()
	})

	return counts, err
}

func GetPlayerLevelsRounded() (counts []Count, err error) {

	item := memcache.ItemPlayerLevelsRounded
	err = memcache.Client().GetSet(item.Key, item.Expiration, &counts, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return counts, err
		}

		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.M{"level": bson.M{"$lte": 1000}}}},
			{{Key: "$group", Value: bson.M{"_id": bson.M{"$trunc": bson.A{"$level", -1}}, "count": bson.M{"$sum": 1}}}},
		}

		cur, err := client.Database(config.C.MongoDatabase, options.Database()).Collection(CollectionPlayers.String()).Aggregate(ctx, pipeline, options.Aggregate())
		if err != nil {
			return counts, err
		}

		defer closeCursor(cur, ctx)

		var maxCount int
		var countsMap = map[int]Count{}

		for cur.Next(ctx) {

			var level Count
			err := cur.Decode(&level)
			if err != nil {
				log.ErrS(err, level.ID)
			}

			countsMap[level.ID] = level

			if level.ID > maxCount {
				maxCount = level.ID
			}
		}

		var counts []Count
		for i := 0; i <= maxCount; i = i + 10 {
			if val, ok := countsMap[i]; ok {
				counts = append(counts, val)
			} else {
				counts = append(counts, Count{ID: i, Count: 0})
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

	c := client.Database(config.C.MongoDatabase).Collection(CollectionPlayers.String())

	_, err = c.BulkWrite(ctx, writes, options.BulkWrite().SetOrdered(false))
	return err
}

type ProfileBadgeStats struct {
	PlayerXP                   int
	PlayerLevel                int
	PlayerXPNeededToLevelUp    int
	PlayerXPNeededCurrentLevel int
	PercentOfLevel             int
}

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
