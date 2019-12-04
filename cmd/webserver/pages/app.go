package pages

import (
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/sql/pics"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

func appRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", appHandler)
	r.Get("/news.json", appNewsAjaxHandler)
	r.Get("/prices.json", appPricesAjaxHandler)
	r.Get("/players.json", appPlayersAjaxHandler)
	r.Get("/items.json", appItemsAjaxHandler)
	r.Get("/reviews.json", appReviewsAjaxHandler)
	r.Get("/time.json", appTimeAjaxHandler)
	r.Get("/{slug}", appHandler)
	return r
}

func appHandler(w http.ResponseWriter, r *http.Request) {

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

	if !helpers.IsValidAppID(idx) {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid App ID: " + id})
		return
	}

	// Get app
	app, err := sql.GetApp(idx, nil)
	if err != nil {

		if err == sql.ErrRecordNotFound {
			returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Sorry but we can not find this app."})
			return
		}

		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the app.", Error: err})
		return
	}

	// Template
	t := appTemplate{}
	t.setBackground(app, false, false)
	t.fill(w, r, app.GetName(), "")
	t.metaImage = app.GetMetaImage()
	t.addAssetCarousel()
	t.addAssetHighCharts()
	t.IncludeSocialJS = true
	t.App = app
	t.Description = template.HTML(app.ShortDescription)
	t.Canonical = app.GetPath()

	//
	var wg sync.WaitGroup

	// Update news, reviews etc
	wg.Add(1)
	go func() {

		defer wg.Done()

		if helpers.IsBot(r.UserAgent()) {
			return
		}

		if app.UpdatedAt.After(time.Now().Add(time.Hour * -24)) {
			return
		}

		err = queue.ProduceToSteam(queue.SteamPayload{AppIDs: []int{app.ID}, Force: false})
		if err != nil && err != queue.ErrInQueue {
			log.Err(err, r)
		} else {
			t.addToast(Toast{Title: "Update", Message: "App has been queued for an update"})
		}
	}()

	// Tags
	wg.Add(1)
	go func(app sql.App) {

		defer wg.Done()

		var err error
		t.Tags, err = app.GetTags()
		log.Err(err, r)
	}(app)

	// Categories
	wg.Add(1)
	go func(app sql.App) {

		defer wg.Done()

		var err error
		t.Categories, err = app.GetCategories()
		log.Err(err, r)
	}(app)

	// Genres
	wg.Add(1)
	go func(app sql.App) {

		defer wg.Done()

		var err error
		t.Genres, err = app.GetGenres()
		log.Err(err, r)
	}(app)

	// Bundles
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Bundles, err = app.GetBundles()
		log.Err(err, r)
	}()

	// Get packages
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Packages, err = sql.GetPackagesAppIsIn(app.ID)
		log.Err(err, r)
	}()

	// Get related apps
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Related, err = app.GetRelatedApps()
		log.Err(err, r)
	}()

	// Get demos
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Demos, err = app.GetDemos()
		log.Err(err, r)
	}()

	// Get DLC
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.DLCs, err = app.GetDLCs()
		log.Err(err, r)
	}()

	// Get Developers
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Developers, err = t.App.GetDevelopers()
		log.Err(err, r)
	}()

	// Get Publishers
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Publishers, err = t.App.GetPublishers()
		log.Err(err, r)
	}()

	// Wait
	wg.Wait()

	// Functions that get called multiple times in the template
	t.Price = app.GetPrice(helpers.GetProductCC(r))

	t.Album = t.App.GetAlbum()

	t.Achievements, err = t.App.GetAchievements()
	log.Err(err, r)

	t.NewsIDs, err = t.App.GetNewsIDs()
	log.Err(err, r)

	t.Stats, err = t.App.GetStats()
	log.Err(err, r)

	t.Prices, err = t.App.GetPrices()
	log.Err(err, r)

	t.Screenshots, err = t.App.GetScreenshots()
	log.Err(err, r)

	t.Movies, err = t.App.GetMovies()
	log.Err(err, r)

	t.Reviews, err = t.App.GetReviews()
	log.Err(err, r)

	t.SteamSpy, err = t.App.GetSteamSpy()
	log.Err(err, r)

	t.Common, err = t.App.GetCommon().Formatted(app.ID, pics.CommonKeys)
	log.Err(err, r)

	t.Extended, err = t.App.GetExtended().Formatted(app.ID, pics.ExtendedKeys)
	log.Err(err, r)

	t.Config, err = t.App.GetConfig().Formatted(app.ID, pics.ConfigKeys)
	log.Err(err, r)

	t.UFS, err = t.App.GetUFS().Formatted(app.ID, pics.UFSKeys)
	log.Err(err, r)

	//
	sort.Slice(t.Reviews.Reviews, func(i, j int) bool {
		return t.Reviews.Reviews[i].VotesGood > t.Reviews.Reviews[j].VotesGood
	})

	// Make banners
	var banners = map[string][]string{
		"primary": {},
		"warning": {},
	}

	if app.ID == 753 {
		banners["primary"] = append(banners["primary"], "This app record is for the Steam client")
	}

	if app.GetCommon().GetValue("app_retired_publisher_request") == "1" {
		banners["warning"] = append(banners["warning"], "At the request of the publisher, "+app.GetName()+" is no longer available for sale on Steam.")
	}

	t.Banners = banners

	//
	returnTemplate(w, r, "app", t)
}

type appTemplate struct {
	GlobalTemplate
	Achievements []sql.AppAchievement
	App          sql.App
	Banners      map[string][]string
	Bundles      []sql.Bundle
	Categories   []sql.Category
	Common       []pics.KeyValue
	Config       []pics.KeyValue
	Demos        []sql.App
	Related      []sql.App
	Developers   []sql.Developer
	DLCs         []sql.App
	Extended     []pics.KeyValue
	Genres       []sql.Genre
	Movies       []sql.AppVideo
	NewsIDs      []int64
	Packages     []sql.Package
	Price        sql.ProductPrice
	Prices       sql.ProductPrices
	Album        pics.AlbumMetaData
	Publishers   []sql.Publisher
	Reviews      sql.AppReviewSummary
	Screenshots  []sql.AppImage
	SteamSpy     sql.AppSteamSpy
	Stats        []sql.AppStat
	Tags         []sql.Tag
	UFS          []pics.KeyValue
}

func (t appTemplate) GetReleaseDate() string {
	nice := t.App.GetReleaseDateNice()
	state := t.App.GetReleaseState()

	if nice != "" {
		state = " (" + state + ")"
	}

	return nice + state
}

func appNewsAjaxHandler(w http.ResponseWriter, r *http.Request) {

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
	err = query.fillFromURL(r.URL.Query())
	if err != nil {
		log.Err(err, r, idx)
	}

	query.limit(r)

	//
	var wg sync.WaitGroup

	// Get events
	var articles []mongo.Article

	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		articles, err = mongo.GetArticlesByApp(idx, query.getOffset64())
		if err != nil {
			log.Err(err, r, idx)
			return
		}

		for k, v := range articles {
			articles[k].Contents = helpers.BBCodeCompiler.Compile(v.Contents)
		}
	}(r)

	// Get total
	var total int
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		app, err := sql.GetApp(idx, nil)
		if err != nil {
			log.Err(err, r, idx)
			return
		}

		newsIDs, err := app.GetNewsIDs()
		if err != nil {
			log.Err(err, r, idx)
			return
		}

		total = len(newsIDs)

	}(r)

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = int64(total)
	response.RecordsFiltered = int64(total)
	response.Draw = query.Draw
	response.limit(r)

	for _, v := range articles {
		response.AddRow(v.OutputForJSON())
	}

	response.output(w, r)
}

func appPricesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	productPricesAjaxHandler(w, r, helpers.ProductTypeApp)
}

func appItemsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Err("invalid id", r)
		return
	}

	idx, err := strconv.Atoi(id)
	if err != nil {
		log.Err(err, r)
		return
	}

	query := DataTablesQuery{}
	err = query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	query.limit(r)

	// Make filter
	var search = query.getSearchString("search")

	filter := bson.D{
		{"app_id", idx},
	}

	if len(search) > 1 {
		filter = append(filter, bson.E{Key: "$or", Value: bson.A{
			bson.M{"name": bson.M{"$regex": search, "$options": "i"}},
			bson.M{"description": bson.M{"$regex": search, "$options": "i"}},
		}})
	}

	//
	var wg sync.WaitGroup

	// Get items
	var items []mongo.AppItem
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		items, err = mongo.GetAppItems(query.getOffset64(), 100, filter, nil)
		if err != nil {
			log.Err(err)
			return
		}

	}()

	// Get total
	var total int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionAppItems, bson.D{{"app_id", idx}}, 0)
		log.Err(err, r)
	}()

	// Get filtered count
	var filtered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionAppItems, filter, 0)
		log.Err(err, r)
	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = total
	response.RecordsFiltered = filtered
	response.Draw = query.Draw
	response.limit(r)

	for _, item := range items {

		response.AddRow([]interface{}{
			item.AppID,              // 0
			item.Bundle,             // 1
			item.Commodity,          // 2
			item.DateCreated,        // 3
			item.Description,        // 4
			item.DisplayType,        // 5
			item.DropInterval,       // 6
			item.DropMaxPerWindow,   // 7
			item.Exchange,           // 8
			item.Hash,               // 9
			item.IconURL,            // 10
			item.IconURLLarge,       // 11
			item.ItemDefID,          // 12
			item.ItemQuality,        // 13
			item.Marketable,         // 14
			item.Modified,           // 15
			item.Name,               // 16
			item.Price,              // 17
			item.Promo,              // 18
			item.Quantity,           // 19
			item.Tags,               // 20
			item.Timestamp,          // 21
			item.Tradable,           // 22
			item.Type,               // 23
			item.WorkshopID,         // 24
			item.Image(36, true),    // 25
			item.Image(256, false),  // 26
			item.GetType(),          // 27
			item.Link(),             // 28
			item.ShortDescription(), // 29
		})
	}

	response.output(w, r)
}

// Player counts chart
func appPlayersAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Err("invalid id", r)
		return
	}

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.AddSelect("max(twitch_viewers)", "max_twitch_viewers")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW()-7d")
	builder.AddWhere("app_id", "=", id)
	builder.AddGroupByTime("10m")
	builder.SetFillNone()

	resp, err := influx.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var hc influx.HighChartsJSON

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

		hc = influx.InfluxResponseToHighCharts(resp.Results[0].Series[0])
	}

	returnJSON(w, r, hc)
}

// Player ranks table
func appTimeAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Err("invalid id", r)
		return
	}

	idx, err := strconv.Atoi(id)
	if err != nil {
		log.Err(err, r)
		return
	}

	query := DataTablesQuery{}
	err = query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	query.limit(r)

	playerAppFilter := bson.D{{"app_id", idx}, {"app_time", bson.M{"$gt": 0}}}

	playerApps, err := mongo.GetPlayerAppsByApp(query.getOffset64(), playerAppFilter)
	if err != nil {
		log.Err(err, r)
		return
	}

	if len(playerApps) < 1 {
		return
	}

	var playerIDsMap = map[int64]int{}
	var playerIDsSlice []int64
	for _, v := range playerApps {
		playerIDsMap[v.PlayerID] = v.AppTime
		playerIDsSlice = append(playerIDsSlice, v.PlayerID)
	}

	//
	var wg sync.WaitGroup

	// Get players
	var playersAppRows []appTimeAjax
	wg.Add(1)
	go func() {

		defer wg.Done()

		players, err := mongo.GetPlayersByID(playerIDsSlice, bson.M{"_id": 1, "persona_name": 1, "avatar": 1, "country_code": 1})
		if err != nil {
			log.Err(err)
			return
		}

		for _, player := range players {

			if _, ok := playerIDsMap[player.ID]; !ok {
				continue
			}

			playersAppRows = append(playersAppRows, appTimeAjax{
				ID:      player.ID,
				Name:    player.PersonaName,
				Avatar:  player.Avatar,
				Time:    playerIDsMap[player.ID],
				Country: player.CountryCode,
			})
		}

		sort.Slice(playersAppRows, func(i, j int) bool {
			return playersAppRows[i].Time > playersAppRows[j].Time
		})

		for k := range playersAppRows {
			playersAppRows[k].Rank = query.getOffset() + k + 1
		}
	}()

	// Get total
	var total int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionPlayerApps, playerAppFilter, 0)
		log.Err(err, r)
	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = total
	response.RecordsFiltered = total
	response.Draw = query.Draw
	response.limit(r)

	for _, v := range playersAppRows {

		response.AddRow([]interface{}{
			strconv.FormatInt(v.ID, 10),          // 0
			v.Name,                               // 1
			helpers.GetTimeLong(v.Time, 3),       // 2
			helpers.GetPlayerFlagPath(v.Country), // 3
			helpers.OrdinalComma(v.Rank),         // 4
			helpers.GetPlayerAvatar(v.Avatar),    // 5
			helpers.GetPlayerPath(v.ID, v.Name),  // 6
			helpers.CountryCodeToName(v.Country), // 7
		})
	}

	response.output(w, r)
}

type appTimeAjax struct {
	ID      int64
	Name    string
	Avatar  string
	Time    int
	Rank    int
	Country string
}

// Review score over time chart
func appReviewsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Err("invalid id", r)
		return
	}

	builder := influxql.NewBuilder()
	builder.AddSelect("mean(reviews_score)", "mean_reviews_score")
	builder.AddSelect("mean(reviews_positive)", "mean_reviews_positive")
	builder.AddSelect("mean(reviews_negative)", "mean_reviews_negative")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW()-365d")
	builder.AddWhere("app_id", "=", id)
	builder.AddGroupByTime("1d")
	builder.SetFillNone()

	resp, err := influx.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var hc influx.HighChartsJSON

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

		hc = influx.InfluxResponseToHighCharts(resp.Results[0].Series[0])
	}

	returnJSON(w, r, hc)
}
