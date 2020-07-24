package queue

import (
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/Jleagle/steam-go/steamid"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/steam"
	"github.com/gamedb/gamedb/pkg/websockets"
	influx "github.com/influxdata/influxdb1-client"
	"go.mongodb.org/mongo-driver/bson"
)

type PlayerMessage struct {
	ID                       int64   `json:"id"`
	SkipPlayerGroups         bool    `json:"dont_queue_groups"`
	SkipGroupUpdate          bool    `json:"dont_queue_group"`
	SkipAchievements         bool    `json:"skip_achievements"`
	SkipExistingPlayer       bool    `json:"skip_existing_player"`
	ForceAchievementsRefresh bool    `json:"force_achievements_refresh"`
	UserAgent                *string `json:"user_agent"`
}

func playerHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := PlayerMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		if payload.ID == 0 {
			message.Ack(false)
			continue
		}

		if payload.UserAgent != nil && helpers.IsBot(*payload.UserAgent) {
			message.Ack(false)
			continue
		}

		payload.ID, err = helpers.IsValidPlayerID(payload.ID)
		if err != nil {
			message.Ack(false)
			continue
		}

		// Update player
		player, err := mongo.GetPlayer(payload.ID)
		if err == nil && payload.SkipExistingPlayer {
			message.Ack(false)
			continue
		}
		newPlayer := err == mongo.ErrNoDocuments
		err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
		if err != nil {

			log.Err(err, payload.ID)
			if err == steamid.ErrInvalidPlayerID {
				sendToFailQueue(message)
			} else {
				sendToRetryQueue(message)
			}
			continue
		}

		player.ID = payload.ID

		if player.Removed {
			message.Ack(false)
			continue
		}

		//
		var wg sync.WaitGroup

		// Calls to api.steampowered.com
		wg.Add(1)
		go func() {

			defer wg.Done()

			err = updatePlayerSummary(&player)
			if err != nil {

				if err == steamapi.ErrProfileMissing {
					message.Ack(false)
				} else {
					steam.LogSteamError(err, payload.ID)
					sendToRetryQueue(message)
				}
				return
			}

			err = updatePlayerGames(&player, payload)
			if err != nil {
				steam.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			err = updatePlayerRecentGames(&player, payload)
			if err != nil {
				steam.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			err = updatePlayerBadges(&player)
			if err != nil {
				steam.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			err = updatePlayerFriends(&player)
			if err != nil {
				steam.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			err = updatePlayerLevel(&player)
			if err != nil {
				steam.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			err = updatePlayerBans(&player)
			if err != nil {
				steam.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		// Calls to store.steampowered.com
		wg.Add(1)
		go func() {

			defer wg.Done()

			err = updatePlayerWishlistApps(&player)
			if err != nil {
				steam.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			err = updatePlayerComments(&player)
			if err != nil {
				steam.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		wg.Wait()

		if message.ActionTaken {
			continue
		}

		// Read from Mongo databases
		wg.Add(1)
		go func() {

			defer wg.Done()

			apps, err := mongo.GetPlayerWishlistAppsByPlayer(player.ID, 0, 0, nil, bson.M{"app_prices": 1})
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			var total = map[steamapi.ProductCC]int{}
			for _, app := range apps {

				for code, price := range app.AppPrices {
					total[code] += price
				}
			}

			player.WishlistTotalCost = total
		}()

		wg.Wait()

		if message.ActionTaken {
			continue
		}

		// Write to databases
		wg.Add(1)
		go func() {

			defer wg.Done()

			err = savePlayerRow(player)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		if newPlayer {
			wg.Add(1)
			go func() {

				defer wg.Done()

				err = updatePlayerFriendRows(player)
				if err != nil {
					log.Err(err, payload.ID)
					sendToRetryQueue(message)
					return
				}
			}()
		}

		wg.Add(1)
		go func() {

			defer wg.Done()

			err = savePlayerToInflux(player)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		wg.Add(1)
		go func() {

			defer wg.Done()

			user, err := mysql.GetUserByKey("steam_id", player.ID, 0)
			if err == mysql.ErrRecordNotFound {
				return
			}
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			err = mongo.CreateUserEvent(nil, user.ID, mongo.EventRefresh)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		wg.Wait()

		if message.ActionTaken {
			continue
		}

		// Clear caches
		wg.Add(1)
		go func() {

			defer wg.Done()

			var items = []string{
				memcache.MemcachePlayer(player.ID).Key,
				memcache.MemcachePlayerInQueue(player.ID).Key,
			}

			err = memcache.Delete(items...)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		// Websocket
		wg.Add(1)
		go func() {

			defer wg.Done()

			wsPayload := PlayerPayload{
				ID:        strconv.FormatInt(player.ID, 10),
				Name:      player.GetName(),
				Link:      player.GetPath(),
				Avatar:    player.GetAvatar(),
				UpdatedAt: time.Now().Unix(),
				Queue:     "player",
			}

			err = ProduceWebsocket(wsPayload, websockets.PagePlayer)
			if err != nil {
				log.Err(err, payload.ID)
			}
		}()

		wg.Wait()

		if message.ActionTaken {
			continue
		}

		// Produce to sub queues
		var produces = []QueueMessageInterface{
			PlayersSearchMessage{Player: player},
		}

		if !player.Removed {
			produces = append(produces,
				PlayersAliasesMessage{PlayerID: player.ID},
				PlayersGroupsMessage{PlayerID: player.ID, PlayerPersonaName: player.PersonaName, PlayerAvatar: player.Avatar, SkipPlayerGroups: payload.SkipPlayerGroups, SkipGroupUpdate: payload.SkipGroupUpdate, UserAgent: payload.UserAgent},
			)
		}

		for _, v := range produces {
			err = produce(v.Queue(), v)
			if err != nil {
				log.Err(err)
				sendToRetryQueue(message)
				break
			}
		}

		if message.ActionTaken {
			continue
		}

		//
		message.Ack(false)
	}
}

func updatePlayerSummary(player *mongo.Player) error {

	summary, err := steam.GetSteam().GetPlayer(player.ID)
	if err == steamapi.ErrProfileMissing {
		player.Removed = true
		return nil
	}

	err = steam.AllowSteamCodes(err)
	if err != nil {
		return err
	}

	// Avatar
	if strings.Contains(summary.ProfileURL, "/id/") {
		player.VanityURL = path.Base(summary.ProfileURL)
	}

	player.Avatar = summary.AvatarHash
	player.CountryCode = summary.CountryCode
	player.ContinentCode = i18n.CountryCodeToContinent(summary.CountryCode)
	player.StateCode = summary.StateCode
	player.PersonaName = summary.PersonaName
	player.TimeCreated = time.Unix(summary.TimeCreated, 0)
	player.PrimaryGroupID = summary.PrimaryClanID
	player.CommunityVisibilityState = summary.CommunityVisibilityState // 1 is private, 3 s public

	return err
}

func updatePlayerGames(player *mongo.Player, payload PlayerMessage) error {

	// Grab games from Steam
	resp, err := steam.GetSteam().GetOwnedGames(player.ID)
	err = steam.AllowSteamCodes(err)
	if err != nil {
		return err
	}

	// Save count
	player.GamesCount = len(resp.Games)

	// Start creating PlayerApp's
	var playerApps = map[int]*mongo.PlayerApp{}
	var appPrices = map[int]map[string]int{}
	var appPriceHour = map[int]map[string]float64{}
	var appIDs []int
	var playtime = 0
	var playtimeWindows = 0
	var playtimeMac = 0
	var playtimeLinux = 0

	for _, v := range resp.Games {

		playtime += v.PlaytimeForever
		playtimeWindows += v.PlaytimeWindows
		playtimeMac += v.PlaytimeMac
		playtimeLinux += v.PlaytimeLinux

		appIDs = append(appIDs, v.AppID)
		playerApps[v.AppID] = &mongo.PlayerApp{
			PlayerID:      player.ID,
			PlayerCountry: player.CountryCode,
			AppID:         v.AppID,
			AppName:       v.Name,
			AppIcon:       v.ImgIconURL,
			AppTime:       v.PlaytimeForever,
		}
		appPrices[v.AppID] = map[string]int{}
		appPriceHour[v.AppID] = map[string]float64{}
	}

	// Save playtime
	player.PlayTime = playtime
	player.PlayTimeWindows = playtimeWindows
	player.PlayTimeMac = playtimeMac
	player.PlayTimeLinux = playtimeLinux

	// Getting missing price info from MySQL
	gameRows, err := mongo.GetAppsByID(appIDs, bson.M{"_id": 1, "prices": 1, "type": 1})
	if err != nil {
		return err
	}

	player.GamesByType = map[string]int{}

	for _, gameRow := range gameRows {

		// Set games by type
		if _, ok := player.GamesByType[gameRow.GetType()]; ok {
			player.GamesByType[gameRow.GetType()]++
		} else {
			player.GamesByType[gameRow.GetType()] = 1
		}

		//
		for code, vv := range gameRow.Prices {

			vv = gameRow.Prices.Get(code)

			appPrices[gameRow.ID][string(code)] = vv.Final
			if appPrices[gameRow.ID][string(code)] > 0 && playerApps[gameRow.ID].AppTime == 0 {
				appPriceHour[gameRow.ID][string(code)] = -1 // Infinite
			} else if appPrices[gameRow.ID][string(code)] > 0 && playerApps[gameRow.ID].AppTime > 0 {
				appPriceHour[gameRow.ID][string(code)] = (float64(appPrices[gameRow.ID][string(code)]) / 100) / (float64(playerApps[gameRow.ID].AppTime) / 60) * 100
			} else {
				appPriceHour[gameRow.ID][string(code)] = 0 // Free
			}
		}

		//
		playerApps[gameRow.ID].AppPrices = appPrices[gameRow.ID]
		log.Err(err)

		//
		playerApps[gameRow.ID].AppPriceHour = appPriceHour[gameRow.ID]
		log.Err(err)
	}

	// Save playerApps to Mongo
	err = mongo.UpdatePlayerApps(playerApps)
	if err != nil {
		return err
	}

	// Get top game for background
	if len(appIDs) > 0 {

		sort.Slice(appIDs, func(i, j int) bool {

			var appID1 = appIDs[i]
			var appID2 = appIDs[j]

			return playerApps[appID1].AppTime > playerApps[appID2].AppTime
		})

		player.BackgroundAppID = appIDs[0]
	}

	// Save stats to player
	var gameStats = mongo.PlayerAppStatsTemplate{}
	for _, v := range playerApps {

		gameStats.All.AddApp(v.AppTime, appPrices[v.AppID], appPriceHour[v.AppID])
		if v.AppTime > 0 {
			gameStats.Played.AddApp(v.AppTime, appPrices[v.AppID], appPriceHour[v.AppID])
		}
	}

	player.GameStats = gameStats

	if !payload.SkipAchievements || payload.ForceAchievementsRefresh {
		if player.UpdatedAt.Before(time.Now().Add(time.Hour * 24 * 13 * -1)) { // Just under 2 weeks
			for _, v := range resp.Games {
				if v.PlaytimeForever > 0 {
					err = ProducePlayerAchievements(player.ID, v.AppID, payload.ForceAchievementsRefresh)
					log.Err(err)
				}
			}
			err = ProducePlayerAchievements(player.ID, 0, false)
			log.Err(err)
		}
	}

	return err
}

func updatePlayerRecentGames(player *mongo.Player, payload PlayerMessage) error {

	// Get data
	oldAppsSlice, err := mongo.GetRecentApps(player.ID, 0, 0, nil)
	if err != nil {
		return err
	}

	newAppsSlice, err := steam.GetSteam().GetRecentlyPlayedGames(player.ID)
	err = steam.AllowSteamCodes(err)
	if err != nil {
		return err
	}

	player.RecentAppsCount = len(newAppsSlice)

	newAppsMap := map[int]steamapi.RecentlyPlayedGame{}
	for _, app := range newAppsSlice {
		newAppsMap[app.AppID] = app
	}

	// Apps to update
	var appsToAdd []mongo.PlayerRecentApp
	for _, v := range newAppsSlice {
		appsToAdd = append(appsToAdd, mongo.PlayerRecentApp{
			PlayerID:        player.ID,
			AppID:           v.AppID,
			AppName:         helpers.GetAppName(v.AppID, v.Name),
			PlayTime2Weeks:  v.PlayTime2Weeks,
			PlayTimeForever: v.PlayTimeForever,
			Icon:            v.ImgIconURL,
			// Logo:            v.ImgLogoURL,
		})
	}

	// Apps to remove
	var appsToRem []int
	for _, v := range oldAppsSlice {
		if _, ok := newAppsMap[v.AppID]; !ok {
			appsToRem = append(appsToRem, v.AppID)
		}
	}

	// Update DB
	err = mongo.DeleteRecentApps(player.ID, appsToRem)
	if err != nil {
		return err
	}

	err = mongo.UpdateRecentApps(appsToAdd)
	if err != nil {
		return err
	}

	//
	if !payload.SkipAchievements && !payload.ForceAchievementsRefresh {
		if player.UpdatedAt.After(time.Now().Add(time.Hour * 24 * 13 * -1)) { // Just under 2 weeks
			for _, v := range newAppsSlice {
				err = ProducePlayerAchievements(player.ID, v.AppID, false)
				log.Err(err)
			}
			err = ProducePlayerAchievements(player.ID, 0, false)
			log.Err(err)
		}
	}

	return nil
}

func updatePlayerBadges(player *mongo.Player) error {

	response, err := steam.GetSteam().GetBadges(player.ID)
	err = steam.AllowSteamCodes(err)
	if err != nil {
		return err
	}

	// Save count
	player.BadgesCount = len(response.Badges)

	// Save stats
	player.BadgeStats = mongo.ProfileBadgeStats{
		PlayerXP:                   response.PlayerXP,
		PlayerLevel:                response.PlayerLevel,
		PlayerXPNeededToLevelUp:    response.PlayerXPNeededToLevelUp,
		PlayerXPNeededCurrentLevel: response.PlayerXPNeededCurrentLevel,
		PercentOfLevel:             response.GetPercentOfLevel(),
	}

	// Save badges
	var playerBadgeSlice []mongo.PlayerBadge
	var appIDSlice []int

	for _, badge := range response.Badges {

		appIDSlice = append(appIDSlice, badge.AppID)
		playerBadgeSlice = append(playerBadgeSlice, mongo.PlayerBadge{
			AppID:               badge.AppID,
			BadgeCompletionTime: time.Unix(badge.CompletionTime, 0),
			BadgeFoil:           bool(badge.BorderColor),
			BadgeID:             badge.BadgeID,
			BadgeItemID:         int64(badge.CommunityItemID),
			BadgeLevel:          badge.Level,
			BadgeScarcity:       badge.Scarcity,
			BadgeXP:             badge.XP,
			PlayerID:            player.ID,
			PlayerIcon:          player.Avatar,
			PlayerName:          player.PersonaName,
		})
	}
	appIDSlice = helpers.UniqueInt(appIDSlice)

	// Make map of app rows
	var appRowsMap = map[int]mongo.App{}
	appRows, err := mongo.GetAppsByID(appIDSlice, bson.M{"_id": 1, "name": 1, "icon": 1})
	if err != nil {
		return err
	}

	for _, v := range appRows {
		appRowsMap[v.ID] = v
	}

	// Finish badges slice
	for k, v := range playerBadgeSlice {

		if v.IsSpecial() {
			if badge, ok := helpers.BuiltInSpecialBadges[v.BadgeID]; ok {
				playerBadgeSlice[k].AppName = badge.Name
			}
		} else {
			if app, ok := appRowsMap[v.AppID]; ok {
				playerBadgeSlice[k].AppName = app.Name
				playerBadgeSlice[k].BadgeIcon = app.Icon
			}
		}
	}

	// Save to Mongo
	return mongo.UpdatePlayerBadges(playerBadgeSlice)
}

func updatePlayerFriends(player *mongo.Player) error {

	newFriendsSlice, err := steam.GetSteam().GetFriendList(player.ID)
	err = steam.AllowSteamCodes(err, 401, 404)
	if err != nil {
		return err
	}

	//
	player.FriendsCount = len(newFriendsSlice)

	// Get data
	oldFriendsSlice, err := mongo.GetFriends(player.ID, 0, 0, nil)
	if err != nil {
		return err
	}

	newFriendsMap := map[int64]steamapi.Friend{}
	for _, friend := range newFriendsSlice {
		newFriendsMap[int64(friend.SteamID)] = friend
	}

	// Friends to add
	var friendIDsToAdd []int64
	var friendsToAdd = map[int64]*mongo.PlayerFriend{}
	for _, v := range newFriendsSlice {
		friendIDsToAdd = append(friendIDsToAdd, int64(v.SteamID))
		friendsToAdd[int64(v.SteamID)] = &mongo.PlayerFriend{
			PlayerID:     player.ID,
			FriendID:     int64(v.SteamID),
			Relationship: v.Relationship,
			FriendSince:  time.Unix(v.FriendSince, 0),
		}
	}

	// Friends to remove
	var friendsToRem []int64
	for _, v := range oldFriendsSlice {
		if _, ok := newFriendsMap[v.FriendID]; !ok {
			friendsToRem = append(friendsToRem, v.FriendID)
		}
	}

	// Fill in missing map the map
	friendRows, err := mongo.GetPlayersByID(friendIDsToAdd, bson.M{
		"_id":          1,
		"avatar":       1,
		"games_count":  1,
		"persona_name": 1,
		"level":        1,
	})
	if err != nil {
		return err
	}

	for _, friend := range friendRows {
		if friend.ID != 0 {

			friendsToAdd[friend.ID].Avatar = friend.Avatar
			friendsToAdd[friend.ID].Games = friend.GamesCount
			friendsToAdd[friend.ID].Name = friend.GetName()
			friendsToAdd[friend.ID].Level = friend.Level
		}
	}

	// Update DB
	err = mongo.DeleteFriends(player.ID, friendsToRem)
	if err != nil {
		return err
	}

	var friendsToAddSlice []*mongo.PlayerFriend
	for _, v := range friendsToAdd {
		friendsToAddSlice = append(friendsToAddSlice, v)
	}

	return mongo.UpdateFriends(friendsToAddSlice)
}

func updatePlayerLevel(player *mongo.Player) error {

	level, err := steam.GetSteam().GetSteamLevel(player.ID)
	err = steam.AllowSteamCodes(err)
	if err != nil {
		return err
	}

	player.Level = level

	return nil
}

func updatePlayerBans(player *mongo.Player) error {

	response, err := steam.GetSteam().GetPlayerBans(player.ID)
	err = steam.AllowSteamCodes(err)
	if err == steamapi.ErrProfileMissing {
		return nil
	}
	if err != nil {
		return err
	}

	player.NumberOfGameBans = response.NumberOfGameBans
	player.NumberOfVACBans = response.NumberOfVACBans

	if response.NumberOfVACBans > 0 {
		player.LastBan = time.Now().Add(time.Hour * 24 * time.Duration(response.DaysSinceLastBan) * -1)
	} else {
		player.LastBan = time.Unix(0, 0)
	}

	//
	player.Bans = mongo.PlayerBans{
		CommunityBanned:  response.CommunityBanned,
		VACBanned:        response.VACBanned,
		NumberOfVACBans:  response.NumberOfVACBans,
		DaysSinceLastBan: response.DaysSinceLastBan,
		NumberOfGameBans: response.NumberOfGameBans,
		EconomyBan:       response.EconomyBan,
	}

	return nil
}

func updatePlayerWishlistApps(player *mongo.Player) error {

	// New
	resp, err := steam.GetSteam().GetWishlist(player.ID)
	err = steam.AllowSteamCodes(err, 500)
	if err == steamapi.ErrWishlistNotFound {
		return nil
	} else if err != nil {
		return err
	}

	var newAppSlice = resp.Items

	player.WishlistAppsCount = len(resp.Items)

	var newAppMap = map[int]steamapi.WishlistItem{}
	for k, v := range newAppSlice {
		newAppMap[int(k)] = v
	}

	// Old
	oldAppsSlice, err := mongo.GetPlayerWishlistAppsByPlayer(player.ID, 0, 0, nil, nil)
	if err != nil {
		return err
	}

	oldAppsMap := map[int]mongo.PlayerWishlistApp{}
	for _, v := range oldAppsSlice {
		oldAppsMap[v.AppID] = v
	}

	// Delete
	var toDelete []int
	for _, v := range oldAppsSlice {
		if _, ok := newAppMap[v.AppID]; !ok {
			toDelete = append(toDelete, v.AppID)
		}
	}

	err = mongo.DeletePlayerWishlistApps(player.ID, toDelete)
	if err != nil {
		return err
	}

	// Add
	var appIDs []int
	var toAdd []mongo.PlayerWishlistApp
	for appID, v := range newAppMap {
		if _, ok := oldAppsMap[appID]; !ok {
			appIDs = append(appIDs, appID)
			toAdd = append(toAdd, mongo.PlayerWishlistApp{
				PlayerID: player.ID,
				AppID:    appID,
				Order:    v.Priority,
			})
		}
	}

	// Fill in data from SQL
	apps, err := mongo.GetAppsByID(appIDs, bson.M{"_id": 1, "name": 1, "icon": 1, "release_state": 1, "release_date": 1, "release_date_unix": 1, "prices": 1})
	if err != nil {
		return err
	}

	var appsMap = map[int]mongo.App{}
	for _, app := range apps {
		appsMap[app.ID] = app
	}

	for k, v := range toAdd {
		toAdd[k].AppPrices = appsMap[v.AppID].Prices.Map()
		toAdd[k].AppName = appsMap[v.AppID].Name
		toAdd[k].AppIcon = appsMap[v.AppID].Icon
		toAdd[k].AppReleaseState = appsMap[v.AppID].ReleaseState
		toAdd[k].AppReleaseDate = time.Unix(appsMap[v.AppID].ReleaseDateUnix, 0)
		toAdd[k].AppReleaseDateNice = appsMap[v.AppID].ReleaseDate
	}

	err = mongo.InsertPlayerWishlistApps(toAdd)
	if err != nil {
		return err
	}

	return nil
}

func updatePlayerComments(player *mongo.Player) error {

	resp, err := steam.GetSteam().GetComments(player.ID, 1, 0)
	if err != nil {
		return err
	}

	player.CommentsCount = resp.TotalCount

	return nil
}

func savePlayerRow(player mongo.Player) error {

	_, err := mongo.ReplaceOne(mongo.CollectionPlayers, bson.D{{"_id", player.ID}}, player)
	return err
}

func updatePlayerFriendRows(player mongo.Player) error {

	update := bson.D{
		{"avatar", player.Avatar},
		{"name", player.PersonaName},
		{"games", player.GamesCount},
		{"level", player.Level},
	}

	_, err := mongo.UpdateManySet(mongo.CollectionPlayerFriends, bson.D{{"friend_id", player.ID}}, update)
	return err
}

func savePlayerToInflux(player mongo.Player) (err error) {

	fields := map[string]interface{}{
		"level":    player.Level,
		"games":    player.GamesCount,
		"badges":   player.BadgesCount,
		"playtime": player.PlayTime,
		"friends":  player.FriendsCount,
		"comments": player.CommentsCount,
	}

	// Add ranks to map
	for k, v := range mongo.PlayerRankFieldsInflux {

		if val, ok := player.Ranks[string(k)]; ok && val > 0 {
			fields[v] = val
		}
	}

	// Save
	_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementPlayers),
		Tags: map[string]string{
			"player_id": strconv.FormatInt(player.ID, 10),
		},
		Fields:    fields,
		Time:      time.Now(),
		Precision: "m",
	})

	return err
}
