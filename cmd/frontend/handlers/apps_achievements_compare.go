package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
)

func appCompareAchievementsRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", appCompareAchievementsHandler)
	r.Get("/{ids}", appCompareAchievementsHandler)
	return r
}

const maxAppAchievementPlayersToCompare = 10

func appCompareAchievementsHandler(w http.ResponseWriter, r *http.Request) {

	// Get app
	appID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid App ID"})
		return
	}

	app, err := mongo.GetApp(appID)
	if err == mongo.ErrNoDocuments {
		returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "App Not Found"})
		return
	} else if err != nil {
		log.ErrS(err)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "Something went wrong fetching this app"})
		return
	}

	// Get achievements
	achievements, err := mongo.GetAppAchievements(0, 0, bson.D{{Key: "app_id", Value: app.ID}}, bson.D{{Key: "completed", Value: -1}})
	if err != nil {
		log.ErrS(err)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "Something went wrong (1001)"})
		return
	}

	// Get players
	var ids = helpers.UniqueString(helpers.RegexInts.FindAllString(chi.URLParam(r, "ids"), -1))
	if len(ids) > maxAppAchievementPlayersToCompare {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Too many players"})
		return
	}

	var players []compareAppAchievementsPlayerTemplate
	var playerIDs []int64
	for _, v := range ids {

		playerID, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			continue
		}

		playerID, err = helpers.IsValidPlayerID(playerID)
		if err != nil {
			continue
		}

		player, err := mongo.GetPlayer(playerID)
		if err != nil {
			err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
			if err != nil {
				log.ErrS(err)
			}
			continue
		}

		playerApp, err := mongo.GetPlayerAppByKey(playerID, app.ID)
		err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
		if err != nil {
			log.ErrS(err)
		}

		players = append(players, compareAppAchievementsPlayerTemplate{
			Player:    player,
			PlayerApp: playerApp,
		})

		playerIDs = append(playerIDs, player.ID)
	}

	// Get player app achievements
	var playerAchievements = map[int64]map[string]mongo.PlayerAchievement{}

	playerAchs, err := mongo.GetPlayerAchievementsByPlayersAndApp(playerIDs, app.ID)
	if err != nil {
		log.ErrS(err)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "Something went wrong (1002)"})
		return
	}

	for _, playerAch := range playerAchs {

		if _, ok := playerAchievements[playerAch.PlayerID]; !ok {
			playerAchievements[playerAch.PlayerID] = map[string]mongo.PlayerAchievement{}
		}

		playerAchievements[playerAch.PlayerID][playerAch.AchievementID] = playerAch
	}

	//
	t := compareAppAchievementsTemplate{}
	t.setBackground(app, false, true)
	t.fill(w, r, "apps_achievements_compare", "Compare Player Achievements", "Compare Player Achievements")
	t.addAssetChosen()
	t.App = app
	t.Achievements = achievements
	t.Players = players
	t.PlayerAchievements = playerAchievements
	t.PlayerIDs = playerIDs

	var sortCol int8 = 1
	for _, v := range players {
		if v.PlayerApp.AppTime > 0 {
			sortCol++
		}
	}

	t.SortCol = sortCol

	returnTemplate(w, r, t)
}

type compareAppAchievementsTemplate struct {
	globalTemplate
	App                mongo.App
	Achievements       []mongo.AppAchievement
	Players            []compareAppAchievementsPlayerTemplate
	PlayerAchievements map[int64]map[string]mongo.PlayerAchievement
	PlayerIDs          []int64
	SortCol            int8
}

func (t compareAppAchievementsTemplate) GetCell(playerID int64, achKey string) mongo.PlayerAchievement {
	return t.PlayerAchievements[playerID][achKey]
}

func (t compareAppAchievementsTemplate) GetRemoveLink(playerID int64) string {

	var ids []string
	for _, v := range t.PlayerIDs {
		if v != playerID {
			ids = append(ids, strconv.FormatInt(v, 10))
		}
	}

	return strings.Join(ids, ",")
}

type compareAppAchievementsPlayerTemplate struct {
	Player    mongo.Player
	PlayerApp mongo.PlayerApp
}
