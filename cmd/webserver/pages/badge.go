package pages

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
)

func BadgeRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", badgeHandler)
	r.Get("/{slug}", badgeHandler)
	r.Get("/players.json", badgeAjaxHandler)
	return r
}

func badgeHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid badge ID"})
		return
	}

	idx, err := strconv.Atoi(id)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid badge ID"})
		return
	}

	val, ok := mongo.Badges[idx]
	if !ok {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid badge ID"})
		return
	}

	t := badgeTemplate{}
	t.fill(w, r, "Badge", "")
	t.Badge = val
	t.Foil = r.URL.Query().Get("foil")

	err = returnTemplate(w, r, "badge", t)
	log.Err(err, r)
}

type badgeTemplate struct {
	GlobalTemplate
	Badge mongo.PlayerBadge
	Foil  string
}

type badgePlayerAjax struct {
	badgeLevel int
	playerID   int64
	playerName string
	playerIcon string
	playerLink string
	rank       int
	time       string
}

func badgeAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid badge ID"})
		return
	}

	idx, err := strconv.Atoi(id)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid badge ID"})
		return
	}

	badge, ok := mongo.Badges[idx]
	if !ok {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid badge ID"})
		return
	}

	query := DataTablesQuery{}
	err = query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	var wg sync.WaitGroup

	var filter = mongo.M{}

	if badge.IsSpecial() {
		filter["app_id"] = 0
		filter["badge_id"] = idx
	} else {
		filter["app_id"] = idx
		filter["badge_id"] = 1
		filter["badge_foil"] = r.URL.Query().Get("foil") == "1"
	}

	var badgeAjaxRows []badgePlayerAjax
	wg.Add(1)
	go func() {

		defer wg.Done()

		badges, err := mongo.GetBadgePlayers(query.getOffset64(), filter)
		log.Err(err, r)

		for k, badge := range badges {

			badgeAjaxRows = append(badgeAjaxRows, badgePlayerAjax{
				badgeLevel: badge.BadgeLevel,
				playerID:   badge.PlayerID,
				playerName: badge.PlayerName,
				playerLink: badge.GetPlayerPath(),
				playerIcon: badge.GetPlayerIcon(),
				rank:       query.getOffset() + k + 1,
				time:       badge.BadgeCompletionTime.Format("2006-01-02 15:04:05"),
			})
		}
	}()

	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountDocuments(mongo.CollectionPlayerBadges, filter, 0)
		log.Err(err, r)
	}()

	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.FormatInt(count, 10)
	response.RecordsFiltered = response.RecordsTotal
	response.Draw = query.Draw

	for _, player := range badgeAjaxRows {
		response.AddRow([]interface{}{
			helpers.OrdinalComma(player.rank), // 0
			player.playerName,                 // 1
			player.playerIcon,                 // 2
			player.badgeLevel,                 // 3
			player.time,                       // 4
			player.playerLink,                 // 5
		})
	}

	response.output(w, r)
}
