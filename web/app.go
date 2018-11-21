package web

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/logging"
	"github.com/gamedb/website/session"
	"github.com/go-chi/chi"
)

func AppHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid App ID."})
		return
	}

	idx, err := strconv.Atoi(id)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid App ID: " + id})
		return
	}

	if !db.IsValidAppID(idx) {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid App ID: " + id})
		return
	}

	// Get app
	app, err := db.GetApp(idx)
	if err != nil {

		if err == db.ErrCantFindApp {
			returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Sorry but we can not find this app."})
			return
		}

		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the app.", Error: err})
		return
	}

	// Redirect to correct slug
	if r.URL.Path != app.GetPath() {
		http.Redirect(w, r, app.GetPath(), 302)
		return
	}

	// Update news, reviews etc
	// todo, add to queue instead!!
	errs := app.UpdateFromRequest(r.UserAgent())
	for _, v := range errs {
		logging.Error(v)
	}

	// Template
	t := appTemplate{}
	t.Fill(w, r, app.GetName())
	t.App = app

	//
	var wg sync.WaitGroup

	// todo, dont call steam here!
	// Get achievements
	//wg.Add(1)
	//go func() {
	//
	//	achievementsResp, _, err := helpers.GetSteam().GetGlobalAchievementPercentagesForApp(app.ID)
	//	if err != nil {
	//
	//		logging.Error(err)
	//
	//	} else {
	//
	//		achievementsMap := make(map[string]float64)
	//		for _, v := range achievementsResp.GlobalAchievementPercentage {
	//			achievementsMap[v.Name] = v.Percent
	//		}
	//
	//		// Get schema
	//		schema, _, err := helpers.GetSteam().GetSchemaForGame(app.ID)
	//		if err != nil {
	//
	//			logging.Error(err)
	//
	//		} else {
	//
	//			// Make template struct
	//			for _, v := range schema.AvailableGameStats.Achievements {
	//				t.Achievements = append(t.Achievements, appAchievementTemplate{
	//					v.Icon,
	//					v.DisplayName,
	//					v.Description,
	//					achievementsMap[v.Name],
	//				})
	//			}
	//		}
	//	}
	//
	//	wg.Done()
	//}()

	// Tags
	wg.Add(1)
	go func() {

		var err error
		t.Tags, err = app.GetTags()
		logging.Error(err)

		wg.Done()
	}()

	// Get prices
	wg.Add(1)
	go func() {

		var code = session.GetCountryCode(r)

		pricesResp, err := db.GetProductPrices(app.ID, db.ProductTypeApp, code)
		if err != nil {

			logging.Error(err)

		} else {

			t.PricesCount = len(pricesResp)

			var prices [][]float64

			for _, v := range pricesResp {

				prices = append(prices, []float64{float64(v.CreatedAt.Unix()), float64(v.PriceAfter) / 100})
			}

			// Add current price
			price := app.GetPrice(code)

			prices = append(prices, []float64{float64(time.Now().Unix()), float64(price.Final) / 100})

			// Make into a JSON string
			pricesBytes, err := json.Marshal(prices)
			if err != nil {

				logging.Error(err)

			} else {

				t.Prices = string(pricesBytes)

			}
		}

		wg.Done()
	}()

	// Get packages
	wg.Add(1)
	go func() {

		var err error
		t.Packages, err = db.GetPackagesAppIsIn(app.ID)
		logging.Error(err)

		wg.Done()
	}()

	// Get DLC
	wg.Add(1)
	go func() {

		var err error
		t.DLC, err = db.GetDLC(app, []string{"id", "name"})
		logging.Error(err)

		wg.Done()
	}()

	// Get reviews
	wg.Add(1)
	go func() {

		reviewsResponse, err := app.GetReviews()
		if err != nil {

			logging.Error(err)

		} else {

			t.ReviewsCount = reviewsResponse.QuerySummary

			// Make slice of playerIDs
			var playerIDs []int64
			for _, v := range reviewsResponse.Reviews {
				playerIDs = append(playerIDs, v.Author.SteamID)
			}

			players, err := db.GetPlayersByIDs(playerIDs)
			if err != nil {

				logging.Error(err)

			} else {

				// Make map of players
				var playersMap = map[int64]db.Player{}
				for _, v := range players {
					playersMap[v.PlayerID] = v
				}

				// Make template slice
				for _, v := range reviewsResponse.Reviews {

					var player db.Player
					if val, ok := playersMap[v.Author.SteamID]; ok {
						player = val
					} else {
						player = db.Player{}
						player.PlayerID = v.Author.SteamID
						player.PersonaName = "Unknown"
					}

					// Remove extra new lines
					regex := regexp.MustCompile("[\n]{3,}") // After comma
					v.Review = regex.ReplaceAllString(v.Review, "\n\n")

					t.Reviews = append(t.Reviews, appReviewTemplate{
						Review:     v.Review,
						Player:     player,
						Date:       time.Unix(v.TimestampCreated, 0).Format(helpers.DateYear),
						VotesGood:  v.VotesUp,
						VotesFunny: v.VotesFunny,
						Vote:       v.VotedUp,
					})
				}
			}
		}

		wg.Done()
	}()

	// Wait
	wg.Wait()

	err = returnTemplate(w, r, "app", t)
	logging.Error(err)
}

type appTemplate struct {
	GlobalTemplate
	App          db.App
	Packages     []db.Package
	DLC          []db.App
	Prices       string
	PricesCount  int
	Achievements []appAchievementTemplate
	Schema       steam.SchemaForGame
	Tags         []db.Tag
	Reviews      []appReviewTemplate
	ReviewsCount steam.ReviewsSummaryResponse
}

type appAchievementTemplate struct {
	Icon        string
	Name        string
	Description string
	Completed   float64
}

func (a appAchievementTemplate) GetCompleted() float64 {
	return helpers.DollarsFloat(a.Completed)
}

type appReviewTemplate struct {
	Review     string
	Player     db.Player
	Date       string
	VotesGood  int
	VotesFunny int
	Vote       bool
}

func AppNewsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid App ID."})
		return
	}

	idx, err := strconv.Atoi(id)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid App ID: " + id})
		return
	}

	query := DataTablesQuery{}
	query.FillFromURL(r.URL.Query())

	//
	var wg sync.WaitGroup

	// Get events
	var articles []db.News

	wg.Add(1)
	go func(r *http.Request) {

		client, ctx, err := db.GetDSClient()
		if err != nil {

			logging.Error(err)

		} else {

			q := datastore.NewQuery(db.KindNews).Filter("app_id =", idx).Limit(100)
			q, err = query.SetOrderOffsetDS(q, map[string]string{})
			q = q.Order("-date")
			if err != nil {

				logging.Error(err)

			} else {

				_, err := client.GetAll(ctx, q, &articles)
				logging.Error(err)

				// todo, use a different bbcode library that works for app 418460 & 218620
				// todo, add http to links here instead of JS
				//var regex = regexp.MustCompile(`href="(?!http)(.*)"`)
				//var conv bbConvert.HTMLConverter
				//conv.ImplementDefaults()
				// Fix broken links
				//v.Contents = regex.ReplaceAllString(v.Contents, `$1http://$2`)
				// Convert BBCdoe to HTML
				//v.Contents = conv.Convert(v.Contents)

			}
		}

		wg.Done()
	}(r)

	// Get total
	var total int
	wg.Add(1)
	go func(r *http.Request) {

		var err error
		app, err := db.GetApp(idx)
		if err != nil {
			logging.Error(err)
			return
		}

		newsIDs, err := app.GetNewsIDs()
		if err != nil {
			logging.Error(err)
			return
		}

		total = len(newsIDs)

		wg.Done()
	}(r)

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(total)
	response.RecordsFiltered = strconv.Itoa(total)
	response.Draw = query.Draw

	for _, v := range articles {
		response.AddRow(v.OutputForJSON(r))
	}

	response.output(w)
}
