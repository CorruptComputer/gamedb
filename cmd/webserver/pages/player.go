package pages

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Jleagle/influxql"
	"github.com/Jleagle/session-go/session"
	"github.com/gamedb/gamedb/cmd/webserver/middleware"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
	"github.com/justinas/nosurf"
	"go.mongodb.org/mongo-driver/bson"
)

func PlayerRouter() http.Handler {

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Use(middleware.MiddlewareCSRF)
		r.Get("/", playerHandler)
		r.Get("/{slug}", playerHandler)
		r.Get("/update.json", playersUpdateAjaxHandler)
	})

	r.Get("/add-friends", playerAddFriendsHandler)
	r.Get("/badges.json", playerBadgesAjaxHandler)
	r.Get("/friends.json", playerFriendsAjaxHandler)
	r.Get("/games.json", playerGamesAjaxHandler)
	r.Get("/groups.json", playerGroupsAjaxHandler)
	r.Get("/history.json", playersHistoryAjaxHandler)
	r.Get("/recent.json", playerRecentAjaxHandler)
	r.Get("/wishlist.json", playerWishlistAppsAjaxHandler)
	return r
}

func playerHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	idx, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid Player ID: " + id, Error: err})
		return
	}

	if !helpers.IsValidPlayerID(idx) {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid Player ID: " + id})
		return
	}

	var code = helpers.GetProductCC(r)

	// Find the player row
	player, err := mongo.GetPlayer(idx)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			err = queue.ProduceToSteam(queue.SteamPayload{ProfileIDs: []int64{idx}}, false)
			err = helpers.IgnoreErrors(err, queue.ErrInQueue)
			log.Err(err)

			// Template
			tm := playerMissingTemplate{}
			tm.fill(w, r, "Looking for player!", "")
			tm.addAssetHighCharts()
			tm.addToast(Toast{Title: "Update", Message: "Player has been queued for an update"})
			tm.Player = player
			tm.DefaultAvatar = helpers.DefaultPlayerAvatar

			returnTemplate(w, r, "player_missing", tm)
		} else {
			returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the player.", Error: err})
		}
		return
	}

	//
	var wg sync.WaitGroup

	// Number of players
	// var players int64
	// wg.Add(1)
	// go func(player mongo.Player) {
	//
	// 	defer wg.Done()
	//
	// 	var err error
	// 	players, err = mongo.CountPlayers()
	// 	log.Err(err, r)
	//
	// }(player)

	// Get bans
	var bans mongo.PlayerBans
	wg.Add(1)
	go func(player mongo.Player) {

		defer wg.Done()

		var err error
		bans, err = player.GetBans()
		log.Err(err, r)

		if err == nil {
			bans.EconomyBan = strings.Title(bans.EconomyBan)
		}

	}(player)

	// Get badge stats
	var badgeStats mongo.ProfileBadgeStats
	wg.Add(1)
	go func(player mongo.Player) {

		defer wg.Done()

		var err error
		badgeStats, err = player.GetBadgeStats()
		log.Err(err, r)

	}(player)

	// Get background app
	var backgroundApp sql.App
	if player.BackgroundAppID > 0 {
		wg.Add(1)
		go func(player mongo.Player) {

			defer wg.Done()

			var err error
			backgroundApp, err = sql.GetApp(player.BackgroundAppID, []string{"id", "name", "background"})
			err = helpers.IgnoreErrors(err, sql.ErrInvalidAppID)
			if err == sql.ErrRecordNotFound {
				err := queue.ProduceToSteam(queue.SteamPayload{AppIDs: []int{player.BackgroundAppID}}, false)
				log.Err(err, player.BackgroundAppID)
			} else if err != nil {
				log.Err(err, player.BackgroundAppID)
			}
		}(player)
	}

	// Wait
	wg.Wait()

	player.VanintyURL = helpers.TruncateString(player.VanintyURL, 14, "...")

	// Game stats
	gameStats, err := player.GetGameStats(code)
	log.Err(err, r)

	// Make banners
	banners := make(map[string][]string)
	var primary []string

	if player.ID == 76561197960287930 {
		primary = append(primary, "This profile belongs to Gabe Newell, the Co-founder of Valve")
	}

	if len(primary) > 0 {
		banners["primary"] = primary
	}

	//
	t := playerTemplate{}

	// Add to Rabbit
	if player.NeedsUpdate(mongo.PlayerUpdateAuto) && !helpers.IsBot(r.UserAgent()) {

		err = queue.ProduceToSteam(queue.SteamPayload{ProfileIDs: []int64{player.ID}}, false)
		if err != nil && err != queue.ErrInQueue {
			log.Err(err, r)
		} else {
			t.addToast(Toast{Title: "Update", Message: "Player has been queued for an update"})
		}
	}

	// Get ranks
	ranks := []mongo.RankKey{mongo.RankKeyLevel, mongo.RankKeyBadges, mongo.RankKeyFriends, mongo.RankKeyComments, mongo.RankKeyGames, mongo.RankKeyPlaytime}
	cc := player.CountryCode
	if cc == "" {
		cc = mongo.RankCountryNone
	}

	for _, v := range ranks {
		if position, ok := player.Ranks[strconv.Itoa(int(v))+"_"+mongo.RankCountryAll]; ok {
			t.Ranks = append(t.Ranks, playerRankTemplate{
				List:     "Globally",
				Metric:   v,
				Position: position,
			})
		}
		if position, ok := player.Ranks[strconv.Itoa(int(v))+"_"+cc]; ok {
			t.Ranks = append(t.Ranks, playerRankTemplate{
				List:     "In the country",
				Metric:   v,
				Position: position,
			})
		}
		if position, ok := player.Ranks[strconv.Itoa(int(v))+"_s-"+player.StateCode]; ok {
			t.Ranks = append(t.Ranks, playerRankTemplate{
				List:     "In the state",
				Metric:   v,
				Position: position,
			})
		}
		if position, ok := player.Ranks[strconv.Itoa(int(v))+"_c-"+player.StateCode]; ok {
			t.Ranks = append(t.Ranks, playerRankTemplate{
				List:     "In the continent",
				Metric:   v,
				Position: position,
			})
		}
	}

	sort.Slice(t.Ranks, func(i, j int) bool {
		return t.Ranks[i].Position < t.Ranks[j].Position
	})

	// Template
	t.setBackground(backgroundApp, true, false)
	t.fill(w, r, player.PersonaName, "")
	t.addAssetHighCharts()
	t.IncludeSocialJS = true

	t.Badges = player.GetSpecialBadges()
	t.BadgeStats = badgeStats
	t.Banners = banners
	t.Bans = bans
	t.Canonical = player.GetPath()
	t.CSRF = nosurf.Token(r)
	t.DefaultAvatar = helpers.DefaultPlayerAvatar
	t.GameStats = gameStats
	t.Player = player

	returnTemplate(w, r, "player", t)
}

type playerTemplate struct {
	GlobalTemplate
	Badges        []mongo.PlayerBadge
	BadgeStats    mongo.ProfileBadgeStats
	Banners       map[string][]string
	Bans          mongo.PlayerBans
	CSRF          string
	DefaultAvatar string
	GameStats     mongo.PlayerAppStatsTemplate
	Player        mongo.Player
	Ranks         []playerRankTemplate
}

type playerMissingTemplate struct {
	GlobalTemplate
	Player        mongo.Player
	DefaultAvatar string
}

type playerRankTemplate struct {
	List     string
	Metric   mongo.RankKey
	Position int
}

func (pr playerRankTemplate) Rank() string {
	return helpers.OrdinalComma(pr.Position)
}

func playerAddFriendsHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	defer func() {

		err := session.Save(w, r)
		log.Err(err)

		http.Redirect(w, r, "/players/"+id+"#friends", http.StatusFound)
	}()

	idx, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return
	}

	if !helpers.IsAdmin(r) {

		user, err := getUserFromSession(r)
		if err != nil {
			return
		}

		if strconv.FormatInt(user.SteamID, 10) != id {
			err = session.SetFlash(r, helpers.SessionBad, "Invalid user")
			log.Err(err)
			return
		}

		if user.PatreonLevel < 2 {
			err = session.SetFlash(r, helpers.SessionBad, "Invalid user level")
			log.Err(err)
			return
		}
	}

	//
	var friendIDs []int64
	var friendIDsMap = map[int64]bool{}

	friends, err := mongo.GetFriends(idx, 0, 0, nil)
	log.Err(err)
	for _, v := range friends {
		friendIDs = append(friendIDs, v.FriendID)
		friendIDsMap[v.FriendID] = true
	}

	// Remove players we already have
	players, err := mongo.GetPlayersByID(friendIDs, bson.M{"_id": 1})
	log.Err(err)
	for _, v := range players {
		delete(friendIDsMap, v.ID)
	}

	// Queue the rest
	var missingPlayerIDs []int64
	for friendID := range friendIDsMap {
		missingPlayerIDs = append(missingPlayerIDs, friendID)
	}

	err = queue.ProduceToSteam(queue.SteamPayload{ProfileIDs: missingPlayerIDs}, false)
	err = helpers.IgnoreErrors(err, queue.ErrInQueue)
	log.Err(err)

	err = session.SetFlash(r, helpers.SessionGood, strconv.Itoa(len(friendIDsMap))+" friends queued")
	log.Err(err)
}

func playerGamesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	playerID := chi.URLParam(r, "id")

	playerIDInt, err := strconv.ParseInt(playerID, 10, 64)
	if err != nil {
		log.Err(err, r)
		return
	}

	query := DataTablesQuery{}
	err = query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	query.limit(r)

	code := helpers.GetProductCC(r)

	//
	var wg sync.WaitGroup

	// Get apps
	var playerApps []mongo.PlayerApp
	wg.Add(1)
	go func() {

		defer wg.Done()

		columns := map[string]string{
			"0": "app_name",
			"1": "app_prices",
			"2": "app_time",
			"3": "app_prices_hour",
		}

		colEdit := func(col string) string {
			if col == "app_prices" || col == "app_prices_hour" {
				col = col + "." + string(code)
			}
			return col
		}

		var err error
		playerApps, err = mongo.GetPlayerApps(playerIDInt, query.getOffset64(), 100, query.getOrderMongo(columns, colEdit))
		log.Err(err)
	}()

	// Get total
	var total int
	wg.Add(1)
	go func() {

		defer wg.Done()

		player, err := mongo.GetPlayer(playerIDInt)
		if err != nil {
			log.Err(err, r)
			return
		}

		total = player.GamesCount

	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = int64(total)
	response.RecordsFiltered = int64(total)
	response.Draw = query.Draw
	response.limit(r)

	for _, pa := range playerApps {
		response.AddRow([]interface{}{
			pa.AppID,
			pa.AppName,
			pa.GetIcon(),
			pa.AppTime,
			pa.GetTimeNice(),
			pa.GetPriceFormatted(code),
			pa.GetPriceHourFormatted(code),
			pa.GetPath(),
		})
	}

	response.output(w, r)

}

func playerRecentAjaxHandler(w http.ResponseWriter, r *http.Request) {

	playerID := chi.URLParam(r, "id")

	playerIDInt, err := strconv.ParseInt(playerID, 10, 64)
	if err != nil {
		log.Err(err, r)
		return
	}

	query := DataTablesQuery{}
	err = query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	query.limit(r)

	//
	var wg sync.WaitGroup

	// Get apps
	var apps []mongo.PlayerRecentApp
	wg.Add(1)
	go func() {

		defer wg.Done()

		columns := map[string]string{
			"0": "name",
			"1": "playtime_2_weeks",
			"2": "playtime_forever",
		}

		var err error
		apps, err = mongo.GetRecentApps(playerIDInt, query.getOffset64(), 100, query.getOrderMongo(columns, nil))
		log.Err(err)
	}()

	// Get total
	var total int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		player, err := mongo.GetPlayer(playerIDInt)
		if err != nil {
			log.Err(err, r)
			return
		}

		total = int64(player.RecentAppsCount)
	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = total
	response.RecordsFiltered = total
	response.Draw = query.Draw
	response.limit(r)

	for _, app := range apps {
		response.AddRow([]interface{}{
			app.AppID,                               // 0
			helpers.GetAppIcon(app.AppID, app.Icon), // 1
			app.AppName,                             // 2
			helpers.GetTimeShort(app.PlayTime2Weeks, 2),  // 3
			helpers.GetTimeShort(app.PlayTimeForever, 2), // 4
			helpers.GetAppPath(app.AppID, app.AppName),   // 5
		})
	}

	response.output(w, r)
}

func playerFriendsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	playerID := chi.URLParam(r, "id")

	playerIDInt, err := strconv.ParseInt(playerID, 10, 64)
	if err != nil {
		log.Err(err, r)
		return
	}

	query := DataTablesQuery{}
	err = query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	query.limit(r)

	//
	var wg sync.WaitGroup

	// Get apps
	var friends []mongo.PlayerFriend
	wg.Add(1)
	go func() {

		defer wg.Done()

		columns := map[string]string{
			"1": "level",
			"2": "games",
			"3": "logged_off",
			"4": "since",
		}

		var err error
		friends, err = mongo.GetFriends(playerIDInt, query.getOffset64(), 100, query.getOrderMongo(columns, nil))
		log.Err(err)
	}()

	// Get total
	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountFriends(playerIDInt)
		if err != nil {
			log.Err(err, r)
			return
		}
	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = count
	response.RecordsFiltered = count
	response.Draw = query.Draw
	response.limit(r)

	for _, friend := range friends {
		response.AddRow([]interface{}{
			strconv.FormatInt(friend.PlayerID, 10), // 0
			friend.GetPath(),                       // 1
			friend.Avatar,                          // 2
			friend.GetName(),                       // 3
			friend.GetLevel(),                      // 4
			friend.Scanned(),                       // 5
			friend.Games,                           // 6
			friend.GetLoggedOff(),                  // 7
			friend.GetFriendSince(),                // 8
		})
	}

	response.output(w, r)
}

func playerBadgesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Err("invalid id: "+id, r)
		return
	}

	idx, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Err(err, r)
		return
	}

	query := DataTablesQuery{}
	err = query.fillFromURL(r.URL.Query())
	if err != nil {
		log.Err(err)
		return
	}

	// Make filter
	var filter = bson.D{{"app_id", bson.M{"$gt": 0}}, {"player_id", idx}}

	//
	var wg sync.WaitGroup

	// Get badges
	var badges []mongo.PlayerBadge
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		badges, err = mongo.GetPlayerEventBadges(query.getOffset64(), filter)
		if err != nil {
			log.Err(err, r)
		}
	}(r)

	// Get total
	var total int64
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionPlayerBadges, filter, 0)
		if err != nil {
			log.Err(err, r)
		}
	}(r)

	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = total
	response.RecordsFiltered = total
	response.Draw = query.Draw

	for _, badge := range badges {
		response.AddRow([]interface{}{
			badge.AppID,        // 0
			badge.AppName,      // 1
			badge.GetAppPath(), // 2
			badge.BadgeCompletionTime.Format("2006-01-02 15:04:05"), // 3
			badge.BadgeFoil,     // 4
			badge.BadgeIcon,     // 5
			badge.BadgeLevel,    // 6
			badge.BadgeScarcity, // 7
			badge.BadgeXP,       // 8
		})
	}

	response.output(w, r)
}

func playerWishlistAppsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Err("invalid id: "+id, r)
		return
	}

	idx, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Err(err, r)
		return
	}

	query := DataTablesQuery{}
	err = query.fillFromURL(r.URL.Query())
	if err != nil {
		log.Err(err)
		return
	}

	//
	var wg sync.WaitGroup

	// Get apps
	var wishlistApps []mongo.PlayerWishlistApp

	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		code := helpers.GetProductCC(r)

		columns := map[string]string{
			"0": "order",
			"1": "app_name",
			"3": "app_release_date",
			"4": "app_prices." + string(code),
		}

		var err error
		wishlistApps, err = mongo.GetPlayerWishlistAppsByPlayer(idx, query.getOffset64(), 0, query.getOrderMongo(columns, nil))
		if err != nil {
			log.Err(err, r)
			return
		}
	}(r)

	// Get total
	var total int64
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		player, err := mongo.GetPlayer(idx)
		if err != nil {
			log.Err(err, r)
		}

		total = int64(player.WishlistAppsCount)
	}(r)

	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = total
	response.RecordsFiltered = total
	response.Draw = query.Draw

	code := helpers.GetProductCC(r)

	for _, app := range wishlistApps {

		var priceFormatted string

		if price, ok := app.AppPrices[code]; ok {
			priceFormatted = helpers.FormatPrice(helpers.GetProdCC(code).CurrencyCode, price)
		} else {
			priceFormatted = "-"
		}

		response.AddRow([]interface{}{
			app.AppID,                // 0
			app.GetName(),            // 1
			app.GetPath(),            // 2
			app.GetIcon(),            // 3
			app.Order,                // 4
			app.GetReleaseState(),    // 5
			app.GetReleaseDateNice(), // 6
			priceFormatted,           // 7
		})
	}

	response.output(w, r)
}

func playerGroupsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Err("invalid id: "+id, r)
		return
	}

	idx, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Err(err, r)
		return
	}

	query := DataTablesQuery{}
	err = query.fillFromURL(r.URL.Query())
	if err != nil {
		log.Err(err)
		return
	}

	//
	var wg sync.WaitGroup

	// Get groups
	var groups []mongo.PlayerGroup
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		columns := map[string]string{
			"0": "group_name",
			"1": "group_members",
		}

		var err error
		groups, err = mongo.GetPlayerGroups(idx, query.getOffset64(), 100, query.getOrderMongo(columns, nil))
		if err != nil {
			log.Err(err, r)
		}
	}(r)

	// Get total
	var total int64

	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		player, err := mongo.GetPlayer(idx)
		if err != nil {
			log.Err(err, r)
		}

		total = int64(player.GroupsCount)
	}(r)

	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = total
	response.RecordsFiltered = total
	response.Draw = query.Draw

	for _, group := range groups {
		response.AddRow([]interface{}{
			group.GroupID64,    // 0
			group.PlayerID,     // 1
			group.GetName(),    // 2
			group.GetPath(),    // 3
			group.GetIcon(),    // 4
			group.GroupMembers, // 5
			group.GetType(),    // 6
			group.GroupPrimary, // 7
			group.GetURL(),     // 8
		})
	}

	response.output(w, r)
}

func playersUpdateAjaxHandler(w http.ResponseWriter, r *http.Request) {

	message, success, err := func(r *http.Request) (string, bool, error) {

		if !nosurf.VerifyToken(nosurf.Token(r), r.URL.Query().Get("csrf")) || r.URL.Query().Get("csrf") == "" {
			return "Invalid CSRF token, please refresh", false, nil
		}

		idx, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			return "Invalid Player ID", false, err
		}

		if !helpers.IsValidPlayerID(idx) {
			return "Invalid Player ID", false, err
		}

		var message string

		player, err := mongo.GetPlayer(idx)
		if err == nil {
			message = "Updating player!"
		} else if err == mongo.ErrNoDocuments {
			message = "Looking for new player!"
		} else {
			log.Err(err, r)
			return "Error looking for player", false, err
		}

		updateType := mongo.PlayerUpdateManual
		if helpers.IsAdmin(r) {
			message = "Admin update!"
			updateType = mongo.PlayerUpdateAdmin
		}

		if !player.NeedsUpdate(updateType) {
			return "Player can't be updated yet", false, nil
		}

		err = queue.ProduceToSteam(queue.SteamPayload{ProfileIDs: []int64{player.ID}}, false)
		if err == queue.ErrInQueue {
			return "Player already queued", false, err
		} else if err != nil {
			log.Err(err, r)
			return "Something has gone wrong", false, err
		}

		return message, true, err
	}(r)

	var response = PlayersUpdateResponse{
		Success: success,
		Toast:   message,
		Log:     err,
	}

	returnJSON(w, r, response)
}

type PlayersUpdateResponse struct {
	Success bool   `json:"success"` // Red or green
	Toast   string `json:"toast"`   // Browser notification
	Log     error  `json:"log"`     // Console log
}

func playersHistoryAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Err("invalid id", r)
		return
	}

	builder := influxql.NewBuilder()
	builder.AddSelect(`mean("level")`, "mean_level")
	builder.AddSelect(`mean("games")`, "mean_games")
	builder.AddSelect(`mean("badges")`, "mean_badges")
	builder.AddSelect(`mean("playtime")`, "mean_playtime")
	builder.AddSelect(`mean("friends")`, "mean_friends")
	builder.AddSelect(`mean("level_rank")`, "mean_level_rank")
	builder.AddSelect(`mean("games_rank")`, "mean_games_rank")
	builder.AddSelect(`mean("badges_rank")`, "mean_badges_rank")
	builder.AddSelect(`mean("playtime_rank")`, "mean_playtime_rank")
	builder.AddSelect(`mean("friends_rank")`, "mean_friends_rank")
	builder.SetFrom(helpers.InfluxGameDB, helpers.InfluxRetentionPolicyAllTime.String(), helpers.InfluxMeasurementPlayers.String())
	builder.AddWhere("player_id", "=", id)
	builder.AddWhere("time", ">", "now()-365d")
	builder.AddGroupByTime("1d")
	builder.SetFillNone()

	resp, err := helpers.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var hc helpers.HighChartsJSON

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

		hc = helpers.InfluxResponseToHighCharts(resp.Results[0].Series[0])
	}

	returnJSON(w, r, hc)
}
