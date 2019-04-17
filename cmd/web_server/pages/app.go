package pages

import (
	"encoding/json"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/influxql"
	main2 "github.com/gamedb/website/cmd/consumers"
	"github.com/gamedb/website/pkg"
	"github.com/go-chi/chi"
)

func appRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", appHandler)
	r.Get("/news.json", appNewsAjaxHandler)
	r.Get("/prices.json", appPricesAjaxHandler)
	r.Get("/players.json", appPlayersAjaxHandler)
	r.Get("/reviews.json", appReviewsAjaxHandler)
	r.Get("/time.json", appTimeAjaxHandler)
	r.Get("/{slug}", appHandler)
	return r
}

func appHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour)

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

	if !pkg.IsValidAppID(idx) {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid App ID: " + id})
		return
	}

	// Get app
	app, err := pkg.GetApp(idx, []string{})
	if err != nil {

		if err == pkg.ErrRecordNotFound {
			returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Sorry but we can not find this app."})
			return
		}

		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the app.", Error: err})
		return
	}

	// Template
	t := appTemplate{}
	t.fill(w, r, app.GetName(), "")
	t.metaImage = app.GetMetaImage()
	t.addAssetCarousel()
	t.addAssetHighCharts()
	t.App = app
	t.Description = template.HTML(app.ShortDescription)

	//
	var wg sync.WaitGroup

	// Update news, reviews etc
	wg.Add(1)
	go func() {

		defer wg.Done()

		if pkg.IsBot(r.UserAgent()) {
			return
		}

		if app.UpdatedAt.Unix() > time.Now().Add(time.Hour * -24).Unix() {
			return
		}

		err = main2.ProduceApp(app.ID)
		if err != nil {
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

		gorm, err := sql.GetMySQLClient()
		if err != nil {
			log.Err(err, r)
			return
		}

		gorm = gorm.Where("JSON_CONTAINS(app_ids, '[" + strconv.Itoa(app.ID) + "]')")
		gorm = gorm.Find(&t.Bundles)

		log.Err(gorm.Error, r)
	}()

	// Get packages
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Packages, err = pkg.GetPackagesAppIsIn(app.ID)
		log.Err(err, r)
	}()

	// Get demos
	wg.Add(1)
	go func() {

		defer wg.Done()

		demoIDs, err := app.GetDemoIDs()
		if err != nil {
			log.Err(err, r)
			return
		}

		if len(demoIDs) > 0 {

			gorm, err := sql.GetMySQLClient()
			if err != nil {
				log.Err(err, r)
				return
			}

			var demos []sql.App
			gorm = gorm.Where("id IN (?)", demoIDs)
			gorm = gorm.Find(&demos)
			if gorm.Error != nil {
				log.Err(gorm.Error, r)
				return
			}

			t.Demos = demos
		}
	}()

	// Get DLC
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.DLCs, err = pkg.GetDLC(app, []string{"id", "name"})
		log.Err(err, r)
	}()

	// Wait
	wg.Wait()

	// Get price
	t.Price = pkg.GetPriceFormatted(app, pkg.GetCountryCode(r))

	// Functions that get called multiple times in the template
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
	t.Developers, err = t.App.GetDevelopers()
	log.Err(err, r)
	t.Publishers, err = t.App.GetPublishers()
	log.Err(err, r)
	t.SteamSpy, err = t.App.GetSteamSpy()
	log.Err(err, r)

	// Make banners
	banners := make(map[string][]string)
	var primary []string

	if app.ID == 753 {
		primary = append(primary, "This app record is for the Steam client")
	}

	if len(t.Achievements) > 5000 {
		primary = append(primary, "Apps are limited to 5000 achievements, but this app created more before the limit put in place")
	}

	if len(primary) > 0 {
		banners["primary"] = primary
	}

	t.Banners = banners

	//
	err = returnTemplate(w, r, "app", t)
	log.Err(err, r)
}

type appTemplate struct {
	GlobalTemplate
	Achievements []sql.AppAchievement
	App          sql.App
	Banners      map[string][]string
	Bundles      []pkg.Bundle
	Demos        []sql.App
	Developers   []sql.Developer
	DLCs         []sql.App
	Genres       []sql.Genre
	Movies       []sql.AppVideo
	NewsIDs      []int64
	Packages     []pkg.Package
	Price        pkg.ProductPriceFormattedStruct
	Prices       sql.ProductPrices
	Publishers   []pkg.Publisher
	Reviews      sql.AppReviewSummary
	Screenshots  []sql.AppImage
	SteamSpy     sql.AppSteamSpy
	Stats        []sql.AppStat
	Tags         []pkg.Tag
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

	ret := setAllowedQueries(w, r, []string{"draw", "start", "order[0][dir]", "order[0][column]"})
	if ret {
		return
	}

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

	//
	var wg sync.WaitGroup

	// Get events
	var articles []pkg.Article

	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		articles, err = pkg.GetArticlesByApp(idx, query.getOffset64())
		if err != nil {
			log.Err(err, r, idx)
			return
		}

		for k, v := range articles {
			articles[k].Contents = pkg.BBCodeCompiler.Compile(v.Contents)
		}

	}(r)

	// Get total
	var total int
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		app, err := pkg.GetApp(idx, []string{})
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
	response.RecordsTotal = strconv.Itoa(total)
	response.RecordsFiltered = strconv.Itoa(total)
	response.Draw = query.Draw

	for _, v := range articles {
		response.AddRow(v.OutputForJSON())
	}

	response.output(w, r)
}

func appPricesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{"code"})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour*3)

	productPricesAjaxHandler(w, r, helpers.ProductTypeApp)
}

// Player counts chart
func appPlayersAjaxHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Err("invalid id", r)
		return
	}

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.AddSelect("max(twitch_viewers)", "max_twitch_viewers")
	builder.SetFrom("GameDB", "alltime", "apps")
	builder.AddWhere("time", ">", "NOW()-7d")
	builder.AddWhere("app_id", "=", id)
	builder.AddGroupByTime("10m")
	builder.SetFillNone()

	resp, err := pkg.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var hc pkg.HighChartsJson

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

		hc = pkg.InfluxResponseToHighCharts(resp.Results[0].Series[0])
	}

	b, err := json.Marshal(hc)
	if err != nil {
		log.Err(err, r)
		return
	}

	err = returnJSON(w, r, b)
	if err != nil {
		log.Err(err, r)
		return
	}
}

// Player ranks table
func appTimeAjaxHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{"draw", "start"})
	if ret {
		return
	}

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

	playerAppFilter := pkg.M{"app_id": idx, "app_time": pkg.M{"$gt": 0}}

	playerApps, err := pkg.GetPlayerAppsByApp(idx, query.getOffset64(), playerAppFilter)
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

		var err error
		players, err := pkg.GetPlayersByID(playerIDsSlice, pkg.M{"_id": 1, "persona_name": 1, "avatar": 1, "country_code": 1})
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
		total, err = pkg.CountDocuments(pkg.CollectionPlayerApps, playerAppFilter)
		log.Err(err, r)
	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.FormatInt(total, 10)
	response.RecordsFiltered = response.RecordsTotal
	response.Draw = query.Draw

	for _, v := range playersAppRows {

		response.AddRow([]interface{}{
			strconv.FormatInt(v.ID, 10),      // 0
			v.Name,                           // 1
			pkg.GetTimeLong(v.Time, 3),       // 2
			pkg.GetPlayerFlagPath(v.Country), // 3
			pkg.OrdinalComma(v.Rank),         // 4
			pkg.GetPlayerAvatar(v.Avatar),    // 5
			pkg.GetPlayerPath(v.ID, v.Name),  // 6
			pkg.CountryCodeToName(v.Country), // 7
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

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Err("invalid id", r)
		return
	}

	builder := influxql.NewBuilder()
	builder.AddSelect("mean(reviews_score)", "mean_reviews_score")
	builder.AddSelect("mean(reviews_positive)", "mean_reviews_positive")
	builder.AddSelect("mean(reviews_negative)", "mean_reviews_negative")
	builder.SetFrom("GameDB", "alltime", "apps")
	builder.AddWhere("time", ">", "NOW()-365d")
	builder.AddWhere("app_id", "=", id)
	builder.AddGroupByTime("1d")
	builder.SetFillNone()

	resp, err := pkg.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var hc pkg.HighChartsJson

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

		hc = pkg.InfluxResponseToHighCharts(resp.Results[0].Series[0])
	}

	b, err := json.Marshal(hc)
	if err != nil {
		log.Err(err, r)
		return
	}

	err = returnJSON(w, r, b)
	if err != nil {
		log.Err(err, r)
		return
	}
}
