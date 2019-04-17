package web

import (
	"encoding/json"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/99designs/basicauth-go"
	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/mongo"
	"github.com/gamedb/website/queue"
	"github.com/gamedb/website/sql"
	"github.com/gamedb/website/websockets"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func adminRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.NoCache)
	r.Use(basicauth.New("Steam", map[string][]string{
		config.Config.AdminUsername: {config.Config.AdminPassword},
	}))
	r.Get("/", adminHandler)
	r.Post("/", adminHandler)
	r.Get("/{option}", adminHandler)
	r.Post("/{option}", adminHandler)
	return r
}

func adminHandler(w http.ResponseWriter, r *http.Request) {

	setCacheHeaders(w, 0)

	option := chi.URLParam(r, "option")

	switch option {
	case "refresh-all-apps":
		go adminQueueEveryApp()
	case "refresh-all-packages":
		go adminQueueEveryPackage()
	case "refresh-all-players":
		go adminQueueEveryPlayer()
	case "refresh-app-players":
		go CronCheckForPlayers()
	case "refresh-genres":
		go CronGenres()
	case "refresh-tags":
		go CronTags()
	case "refresh-developers":
		go CronDevelopers()
	case "refresh-publishers":
		go CronPublishers()
	case "refresh-donations":
		go CronDonations()
	case "refresh-ranks":
		go CronRanks()
	case "wipe-memcache":
		go adminMemcache()
	case "delete-bin-logs":
		go adminDeleteBinLogs(r)
	case "disable-consumers":
		go adminDisableConsumers()
	case "run-dev-code":
		go adminDev()
	case "queues":
		err := r.ParseForm()
		if err != nil {
			log.Err(err, r)
		}
		go adminQueues(r)
	}

	// Redirect away after action
	if option != "" {
		http.Redirect(w, r, "/admin?"+option, http.StatusFound)
		return
	}

	// Get configs for times
	configs, err := sql.GetConfigs([]string{
		sql.ConfTagsUpdated,
		sql.ConfGenresUpdated,
		sql.ConfGenresUpdated,
		sql.ConfDonationsUpdated,
		sql.ConfRanksUpdated,
		sql.ConfAddedAllApps,
		sql.ConfDevelopersUpdated,
		sql.ConfPublishersUpdated,
		sql.ConfWipeMemcache + "-" + config.Config.Environment.Get(),
		sql.ConfRunDevCode,
		sql.ConfGarbageCollection,
		sql.ConfAddedAllAppPlayers,
	})
	log.Err(err, r)

	// Template
	t := adminTemplate{}
	t.fill(w, r, "Admin", "")
	t.Configs = configs
	t.Goroutines = runtime.NumGoroutine()

	//
	gorm, err := sql.GetMySQLClient()
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "Can't connect to mysql", Error: err})
		return
	}

	gorm.Raw("show binary logs").Scan(&t.BinLogs)

	var total uint64
	for k, v := range t.BinLogs {
		total = total + v.Bytes
		t.BinLogs[k].Total = total
	}

	gorm.Raw("SELECT * FROM information_schema.processlist where command != 'sleep'").Scan(&t.Queries)

	err = returnTemplate(w, r, "admin", t)
	log.Err(err, r)
}

type adminTemplate struct {
	GlobalTemplate
	Errors     []string
	Configs    map[string]sql.Config
	Goroutines int
	Queries    []adminQuery
	BinLogs    []adminBinLog
}

type adminQuery struct {
	ID       int    `gorm:"column:ID"`
	User     string `gorm:"column:USER"`
	Host     string `gorm:"column:HOST"`
	Database string `gorm:"column:DB"`
	Command  string `gorm:"column:COMMAND"`
	Seconds  int64  `gorm:"column:TIME"`
	State    string `gorm:"column:STATE"`
	Info     string `gorm:"column:INFO"`
}

type adminBinLog struct {
	Name      string `gorm:"column:Log_name"`
	Bytes     uint64 `gorm:"column:File_size"`
	Encrypted string `gorm:"column:Encrypted"`
	Total     uint64
}

func (at adminTemplate) GetMCConfigKey() string {
	return "wipe-memcache" + "-" + config.Config.Environment.Get()
}

func adminDisableConsumers() {

}

func adminQueueEveryApp() {

	var last = 0
	var keepGoing = true
	var apps steam.AppList
	var err error
	var count int

	for keepGoing {

		apps, b, err := helpers.GetSteam().GetAppList(1000, last, 0, "")
		err = helpers.HandleSteamStoreErr(err, b, nil)
		if err != nil {
			log.Err(err)
			return
		}

		count = count + len(apps.Apps)

		for _, v := range apps.Apps {
			err = queue.ProduceApp(v.AppID)
			if err != nil {
				log.Err(err, strconv.Itoa(v.AppID))
				continue
			}
			last = v.AppID
		}

		keepGoing = apps.HaveMoreResults
	}

	log.Info("Found " + strconv.Itoa(count) + " apps")

	//
	err = sql.SetConfig(sql.ConfAddedAllApps, strconv.FormatInt(time.Now().Unix(), 10))
	log.Err(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	log.Err(err)

	if err == nil {
		page.Send(adminWebsocket{sql.ConfAddedAllApps + " complete"})
	}

	log.Info(strconv.Itoa(len(apps.Apps)) + " apps added to rabbit")
}

func adminQueueEveryPackage() {

	apps, err := sql.GetAppsWithColumnDepth("packages", 2, []string{"packages"})
	if err != nil {
		log.Err(err)
		return
	}

	packageIDs := map[int]bool{}
	for _, v := range apps {

		packages, err := v.GetPackages()
		if err != nil {
			log.Err(err)
			return
		}

		for _, vv := range packages {
			packageIDs[vv] = true
		}
	}

	for k := range packageIDs {

		err = queue.ProducePackage(k)
		if err != nil {
			log.Err(err)
			return
		}
	}

	//
	err = sql.SetConfig(sql.ConfAddedAllPackages, strconv.FormatInt(time.Now().Unix(), 10))
	log.Err(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	log.Err(err)

	if err == nil {
		page.Send(adminWebsocket{sql.ConfAddedAllPackages + " complete"})
	}

	log.Info(strconv.Itoa(len(packageIDs)) + " packages added to rabbit")
}

func adminQueueEveryPlayer() {

	cronLogInfo("Queueing every player")

	players, err := mongo.GetPlayers(0, 0, mongo.D{{"_id", 1}}, nil, mongo.M{"_id": 1})
	if err != nil {
		log.Err(err)
		return
	}

	for _, player := range players {

		err = queue.ProducePlayer(player.ID)
		if err != nil {
			log.Err(err)
			return
		}
	}

	//
	err = sql.SetConfig(sql.ConfAddedAllPlayers, strconv.FormatInt(time.Now().Unix(), 10))
	log.Err(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	log.Err(err)

	if err == nil {
		page.Send(adminWebsocket{sql.ConfAddedAllPlayers + " complete"})
	}

	log.Info(strconv.Itoa(len(players)) + " players added to rabbit")
}

func CronDonations() {

	// donations, err := db.GetDonations(0, 0)
	// if err != nil {
	// 	cronLogErr(err)
	// 	return
	// }
	//
	// // map[player]total
	// counts := make(map[int64]int)
	//
	// for _, v := range donations {
	//
	// 	if _, ok := counts[v.PlayerID]; ok {
	// 		counts[v.PlayerID] = counts[v.PlayerID] + v.AmountUSD
	// 	} else {
	// 		counts[v.PlayerID] = v.AmountUSD
	// 	}
	// }
	//
	// for k, v := range counts {
	// 	player, err := mongo.GetPlayer(k)
	// 	if err != nil {
	// 		cronLogErr(err)
	// 		continue
	// 	}
	//
	// 	player.Donated = v
	// 	err = db.SaveKind(player.GetKey(), player)
	// 	cronLogErr(err)
	// }

	//
	err := sql.SetConfig(sql.ConfDonationsUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	log.Err(err)

	if err == nil {
		page.Send(adminWebsocket{sql.ConfDonationsUpdated + " complete"})
	}

	// cronLogInfo("Updated " + strconv.Itoa(len(counts)) + " player donation counts")
}

func adminQueues(r *http.Request) {

	if val := r.PostForm.Get("player-id"); val != "" {

		vals := strings.Split(val, ",")

		for _, val := range vals {

			val = strings.TrimSpace(val)

			playerID, err := strconv.ParseInt(val, 10, 64)
			log.Err(err, r)
			if err == nil {

				err = queue.ProducePlayer(playerID)
				log.Err(err, r)
			}
		}
	}

	if val := r.PostForm.Get("app-id"); val != "" {

		vals := strings.Split(val, ",")

		for _, val := range vals {

			val = strings.TrimSpace(val)

			appID, err := strconv.Atoi(val)
			if err == nil {

				err = queue.ProduceApp(appID)
				log.Err(err, r)
			}
		}
	}

	if val := r.PostForm.Get("package-id"); val != "" {

		vals := strings.Split(val, ",")

		for _, val := range vals {

			val = strings.TrimSpace(val)

			packageID, err := strconv.Atoi(val)
			if err == nil {

				err = queue.ProducePackage(packageID)
				log.Err(err, r)
			}
		}
	}

	if val := r.PostForm.Get("bundle-id"); val != "" {

		vals := strings.Split(val, ",")

		for _, val := range vals {

			val = strings.TrimSpace(val)

			bundleID, err := strconv.Atoi(val)
			if err == nil {

				err = queue.ProduceBundle(bundleID, 0)
				log.Err(err, r)
			}
		}
	}

	if val := r.PostForm.Get("apps-ts"); val != "" {

		log.Info("Queueing apps")

		ts, err := strconv.ParseInt(val, 10, 64)
		log.Err(err, r)
		if err == nil {

			apps, b, err := helpers.GetSteam().GetAppList(100000, 0, ts, "")
			err = helpers.HandleSteamStoreErr(err, b, nil)
			log.Err(err, r)
			if err == nil {

				log.Info("Found " + strconv.Itoa(len(apps.Apps)) + " apps")

				for _, v := range apps.Apps {
					err = queue.ProduceApp(v.AppID)
					log.Err(err, r)
				}
			}
		}
	}
}

func CronGenres() {

	cronLogInfo("Genres updating")

	// Get current genres, to delete old ones
	currentGenres, err := sql.GetAllGenres(true)
	if err != nil {
		cronLogErr(err)
		return
	}

	genresToDelete := map[int]bool{}
	for _, v := range currentGenres {
		genresToDelete[v.ID] = true
	}

	var genreNameMap = map[int]string{}
	for _, v := range currentGenres {
		genreNameMap[v.ID] = strings.TrimSpace(v.GetName())
	}

	// Get apps from mysql
	appsWithGenres, err := sql.GetAppsWithColumnDepth("genres", 2, []string{"genres", "prices", "reviews_score"})
	cronLogErr(err)

	cronLogInfo("Found " + strconv.Itoa(len(appsWithGenres)) + " apps with genres")

	newGenres := make(map[int]*statsRow)
	for _, app := range appsWithGenres {

		appGenreIDs, err := app.GetGenreIDs()
		if err != nil {
			cronLogErr(err)
			continue
		}

		if len(appGenreIDs) == 0 {
			// appGenreIDs = []db.AppGenre{{ID: 0, Name: ""}}
		}

		// For each genre in an app
		for _, genreID := range appGenreIDs {

			delete(genresToDelete, genreID)

			var genreName string
			if val, ok := genreNameMap[genreID]; ok {
				genreName = val
			} else {
				// genreName = "Unknown"
				continue
			}

			if _, ok := newGenres[genreID]; ok {
				newGenres[genreID].count++
				newGenres[genreID].totalScore += app.ReviewsScore
			} else {
				newGenres[genreID] = &statsRow{
					name:       genreName,
					count:      1,
					totalScore: app.ReviewsScore,
					totalPrice: map[steam.CountryCode]int{},
				}
			}

			for code := range steam.Countries {
				price, err := app.GetPrice(code)
				if err != nil {
					// cronLogErr(err, r)
					continue
				}
				newGenres[genreID].totalPrice[code] += price.Final
			}
		}
	}

	var limit int
	var wg sync.WaitGroup

	// Delete old genres
	limit++
	wg.Add(1)
	go func() {

		defer func() {
			limit--
			wg.Done()
		}()

		var genresToDeleteSlice []int
		for genreID := range genresToDelete {
			genresToDeleteSlice = append(genresToDeleteSlice, genreID)
		}

		err := sql.DeleteGenres(genresToDeleteSlice)
		cronLogErr(err)

	}()

	wg.Wait()

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		cronLogErr(err)
		return
	}

	// Update current genres
	var count = 1
	for k, v := range newGenres {

		if limit >= 2 {
			wg.Wait()
		}

		adminStatsLogger("genre", count, len(newGenres), v.name)

		limit++
		wg.Add(1)
		go func(genreID int, v *statsRow) {

			defer func() {
				limit--
				wg.Done()
			}()

			var genre sql.Genre

			gorm = gorm.Unscoped().FirstOrInit(&genre, sql.Genre{ID: genreID})
			cronLogErr(gorm.Error)

			genre.Name = v.name
			genre.Apps = v.count
			genre.MeanPrice = v.getMeanPrice()
			genre.MeanScore = v.getMeanScore()
			genre.DeletedAt = nil

			gorm = gorm.Unscoped().Save(&genre)
			cronLogErr(gorm.Error)

		}(k, v)

		count++
	}
	wg.Wait()

	//
	err = sql.SetConfig(sql.ConfGenresUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	//
	page, err := websockets.GetPage(websockets.PageAdmin)
	cronLogErr(err)

	if err == nil {
		page.Send(adminWebsocket{sql.ConfGenresUpdated + " complete"})
	}

	//
	err = helpers.GetMemcache().Delete(helpers.MemcacheGenreKeyNames.Key)
	err = helpers.IgnoreErrors(err, helpers.ErrCacheMiss)
	cronLogErr(err)

	//
	cronLogInfo("Genres updated")
}

func CronPublishers() {

	cronLogInfo("Publishers updating")

	// Get current publishers, to delete old ones
	allPublishers, err := sql.GetAllPublishers()
	if err != nil {
		cronLogErr(err)
		return
	}

	publishersToDelete := map[int]bool{}
	for _, publisherRow := range allPublishers {
		publishersToDelete[publisherRow.ID] = true
	}

	var publisherNameMap = map[int]string{}
	for _, v := range allPublishers {
		publisherNameMap[v.ID] = strings.TrimSpace(v.GetName())
	}

	// Get apps from mysql
	appsWithPublishers, err := sql.GetAppsWithColumnDepth("publishers", 2, []string{"publishers", "prices", "reviews_score"})
	cronLogErr(err)

	cronLogInfo("Found " + strconv.Itoa(len(appsWithPublishers)) + " apps with publishers")

	newPublishers := make(map[int]*statsRow)
	for _, app := range appsWithPublishers {

		appPublishers, err := app.GetPublisherIDs()
		if err != nil {
			cronLogErr(err)
			continue
		}

		if len(appPublishers) == 0 {
			// appPublishers = []string{""}
		}

		// For each publisher in an app
		for _, appPublisherID := range appPublishers {

			delete(publishersToDelete, appPublisherID)

			var publisherName string
			if val, ok := publisherNameMap[appPublisherID]; ok {
				publisherName = val
			} else {
				// publisherName = "Unknown"
				continue
			}

			if _, ok := newPublishers[appPublisherID]; ok {
				newPublishers[appPublisherID].count++
				newPublishers[appPublisherID].totalScore += app.ReviewsScore
			} else {
				newPublishers[appPublisherID] = &statsRow{
					name:       publisherName,
					count:      1,
					totalPrice: map[steam.CountryCode]int{},
					totalScore: app.ReviewsScore,
				}
			}

			for code := range steam.Countries {
				price, err := app.GetPrice(code)
				if err != nil {
					continue
				}
				newPublishers[appPublisherID].totalPrice[code] += price.Final
			}
		}
	}

	var limit int
	var wg sync.WaitGroup

	// Delete old publishers
	limit++
	wg.Add(1)
	go func() {

		defer func() {
			limit--
			wg.Done()
		}()

		var pubsToDeleteSlice []int
		for publisherID := range publishersToDelete {
			pubsToDeleteSlice = append(pubsToDeleteSlice, publisherID)
		}

		err := sql.DeletePublishers(pubsToDeleteSlice)
		cronLogErr(err)

	}()

	wg.Wait()

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		cronLogErr(err)
		return
	}

	// Update current publishers
	var count = 1
	for k, v := range newPublishers {

		if limit >= 2 {
			wg.Wait()
		}

		adminStatsLogger("publisher", count, len(newPublishers), v.name)

		limit++
		wg.Add(1)
		go func(publisherID int, v *statsRow) {

			defer func() {
				limit--
				wg.Done()
			}()

			var publisher sql.Publisher

			gorm = gorm.Unscoped().FirstOrInit(&publisher, sql.Publisher{ID: publisherID})
			cronLogErr(gorm.Error)

			publisher.Name = v.name
			publisher.Apps = v.count
			publisher.MeanPrice = v.getMeanPrice()
			publisher.MeanScore = v.getMeanScore()
			publisher.DeletedAt = nil

			gorm = gorm.Unscoped().Save(&publisher)
			cronLogErr(gorm.Error)

		}(k, v)

		count++
	}

	wg.Wait()

	//
	err = sql.SetConfig(sql.ConfPublishersUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	//
	page, err := websockets.GetPage(websockets.PageAdmin)
	cronLogErr(err)

	if err == nil {
		page.Send(adminWebsocket{sql.ConfPublishersUpdated + " complete"})
	}

	//
	err = helpers.GetMemcache().Delete(helpers.MemcachePublisherKeyNames.Key)
	cronLogErr(err)

	//
	cronLogInfo("Publishers updated")
}

func CronDevelopers() {

	cronLogInfo("Developers updating")

	// Get current developers, to delete old ones
	allDevelopers, err := sql.GetAllDevelopers([]string{"id", "name"})
	if err != nil {
		cronLogErr(err)
		return
	}

	developersToDelete := map[int]bool{}
	for _, v := range allDevelopers {
		developersToDelete[v.ID] = true
	}

	var developersNameMap = map[int]string{}
	for _, v := range allDevelopers {
		developersNameMap[v.ID] = strings.TrimSpace(v.GetName())
	}

	// Get apps from mysql
	appsWithDevelopers, err := sql.GetAppsWithColumnDepth("developers", 2, []string{"developers", "prices", "reviews_score"})
	cronLogErr(err)

	cronLogInfo("Found " + strconv.Itoa(len(appsWithDevelopers)) + " apps with developers")

	newDevelopers := make(map[int]*statsRow)
	for _, app := range appsWithDevelopers {

		appDevelopers, err := app.GetDeveloperIDs()
		if err != nil {
			cronLogErr(err)
			continue
		}

		if len(appDevelopers) == 0 {
			// appDevelopers = []string{""}
		}

		// For each developer in an app
		for _, appDeveloperID := range appDevelopers {

			delete(developersToDelete, appDeveloperID)

			var developersName string
			if val, ok := developersNameMap[appDeveloperID]; ok {
				developersName = val
			} else {
				continue
			}

			if _, ok := newDevelopers[appDeveloperID]; ok {
				newDevelopers[appDeveloperID].count++
				newDevelopers[appDeveloperID].totalScore += app.ReviewsScore
			} else {
				newDevelopers[appDeveloperID] = &statsRow{
					name:       developersName,
					count:      1,
					totalPrice: map[steam.CountryCode]int{},
					totalScore: app.ReviewsScore,
				}
			}

			for code := range steam.Countries {
				price, err := app.GetPrice(code)
				if err != nil {
					// cronLogErr(err, r)
					continue
				}
				newDevelopers[appDeveloperID].totalPrice[code] += price.Final
			}
		}
	}

	var limit int
	var wg sync.WaitGroup

	// Delete old developers
	limit++
	wg.Add(1)
	go func() {

		defer func() {
			limit--
			wg.Done()
		}()

		var devsToDeleteSlice []int
		for k := range developersToDelete {
			devsToDeleteSlice = append(devsToDeleteSlice, k)
		}

		err := sql.DeleteDevelopers(devsToDeleteSlice)
		cronLogErr(err)

	}()

	wg.Wait()

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		cronLogErr(err)
		return
	}

	// Update current developers
	var count = 1
	for k, v := range newDevelopers {

		if limit >= 2 {
			wg.Wait()
		}

		adminStatsLogger("developer", count, len(newDevelopers), v.name)

		limit++
		wg.Add(1)
		go func(developerInt int, v *statsRow) {

			defer func() {
				limit--
				wg.Done()
			}()

			var developer sql.Developer

			gorm = gorm.Unscoped().FirstOrInit(&developer, sql.Developer{ID: developerInt})
			cronLogErr(gorm.Error)

			developer.Name = v.name
			developer.Apps = v.count
			developer.MeanPrice = v.getMeanPrice()
			developer.MeanScore = v.getMeanScore()
			developer.DeletedAt = nil

			gorm = gorm.Unscoped().Save(&developer)
			cronLogErr(gorm.Error)

		}(k, v)

		count++
	}
	wg.Wait()

	//
	err = sql.SetConfig(sql.ConfDevelopersUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	//
	page, err := websockets.GetPage(websockets.PageAdmin)
	cronLogErr(err)

	if err == nil {
		page.Send(adminWebsocket{sql.ConfDevelopersUpdated + " complete"})
	}

	//
	err = helpers.GetMemcache().Delete(helpers.MemcacheDeveloperKeyNames.Key)
	cronLogErr(err)

	//
	cronLogInfo("Developers updated")
}

func CronTags() {

	// Get current tags, to delete old ones
	tags, err := sql.GetAllTags()
	if err != nil {
		cronLogErr(err)
		return
	}

	tagsToDelete := map[int]int{}
	for _, tag := range tags {
		tagsToDelete[tag.ID] = tag.ID
	}

	// Get tags from Steam
	tagsResp, b, err := helpers.GetSteam().GetTags()
	err = helpers.HandleSteamStoreErr(err, b, nil)
	if err != nil {
		cronLogErr(err)
		return
	}

	steamTagMap := tagsResp.GetMap()

	appsWithTags, err := sql.GetAppsWithColumnDepth("tags", 2, []string{"tags", "prices", "reviews_score"})
	cronLogErr(err)

	cronLogInfo("Found " + strconv.Itoa(len(appsWithTags)) + " apps with tags")

	newTags := make(map[int]*statsRow)
	for _, app := range appsWithTags {

		appTags, err := app.GetTagIDs()
		if err != nil {
			cronLogErr(err)
			continue
		}

		if len(appTags) == 0 {
			// appTags = []int{}
		}

		// For each tag in an app
		for _, tagID := range appTags {

			delete(tagsToDelete, tagID)

			if _, ok := newTags[tagID]; ok {
				newTags[tagID].count++
				newTags[tagID].totalScore += app.ReviewsScore
			} else {
				newTags[tagID] = &statsRow{
					name:       strings.TrimSpace(steamTagMap[tagID]),
					count:      1,
					totalPrice: map[steam.CountryCode]int{},
					totalScore: app.ReviewsScore,
				}
			}

			for code := range steam.Countries {
				price, err := app.GetPrice(code)
				if err != nil {
					// cronLogErr(err, r)
					continue
				}
				newTags[tagID].totalPrice[code] += price.Final
			}
		}
	}

	var limit int
	var wg sync.WaitGroup

	// Delete old tags
	limit++
	wg.Add(1)
	go func() {

		defer func() {
			limit--
			wg.Done()
		}()

		var tagsToDeleteSlice []int
		for _, v := range tagsToDelete {
			tagsToDeleteSlice = append(tagsToDeleteSlice, v)
		}

		err := sql.DeleteTags(tagsToDeleteSlice)
		cronLogErr(err)

	}()

	wg.Wait()

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		cronLogErr(err)
		return
	}

	// Update current tags
	var count = 1
	for k, v := range newTags {

		if limit >= 2 {
			wg.Wait()
		}

		adminStatsLogger("tag", count, len(newTags), v.name)

		limit++
		wg.Add(1)
		go func(tagID int, v *statsRow) {

			defer func() {
				limit--
				wg.Done()
			}()

			var tag sql.Tag

			gorm = gorm.Unscoped().FirstOrInit(&tag, sql.Tag{ID: tagID})
			cronLogErr(gorm.Error)

			tag.Name = v.name
			tag.Apps = v.count
			tag.MeanPrice = v.getMeanPrice()
			tag.MeanScore = v.getMeanScore()
			tag.DeletedAt = nil

			gorm = gorm.Unscoped().Save(&tag)
			cronLogErr(gorm.Error)

		}(k, v)

		count++
	}
	wg.Wait()

	//
	err = sql.SetConfig(sql.ConfTagsUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	//
	page, err := websockets.GetPage(websockets.PageAdmin)
	cronLogErr(err)

	if err == nil {
		page.Send(adminWebsocket{sql.ConfTagsUpdated + " complete"})
	}

	//
	err = helpers.GetMemcache().Delete(helpers.MemcacheTagKeyNames.Key)
	cronLogErr(err)

	//
	cronLogInfo("Tags updated")
}

func adminStatsLogger(tableName string, count int, total int, rowName string) {

	log.Info("Updating " + tableName + " - " + strconv.Itoa(count) + " / " + strconv.Itoa(total) + ": " + rowName)
}

func CronRanks() {

	cronLogInfo("Ranks updated started")

	cronLogInfo("Level")
	err := mongo.RankPlayers("level", "level_rank")
	log.Warning(err)

	cronLogInfo("Games")
	err = mongo.RankPlayers("games_count", "games_rank")
	log.Warning(err)

	cronLogInfo("Badges")
	err = mongo.RankPlayers("badges_count", "badges_rank")
	log.Warning(err)

	cronLogInfo("Time")
	err = mongo.RankPlayers("play_time", "play_time_rank")
	log.Warning(err)

	cronLogInfo("Friends")
	err = mongo.RankPlayers("friends_count", "friends_rank")
	log.Warning(err)

	//
	err = sql.SetConfig(sql.ConfRanksUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	page, err := websockets.GetPage(websockets.PageAdmin)

	if err == nil {
		page.Send(adminWebsocket{sql.ConfRanksUpdated + " complete"})
	}

	cronLogInfo("Ranks updated")
}

func CronCheckForPlayers() {

	log.Info("Queueing apps for player checks")

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Critical(err)
		return
	}

	gorm = gorm.Select([]string{"id"})
	gorm = gorm.Order("id ASC")
	gorm = gorm.Model(&[]sql.App{})

	var appIDs []int
	gorm = gorm.Pluck("id", &appIDs)
	if gorm.Error != nil {
		log.Critical(gorm.Error)
	}

	appIDs = append(appIDs, 0) // Steam client

	// Chunk appIDs
	var chunks [][]int
	for i := 0; i < len(appIDs); i += 10 {
		end := i + 10

		if end > len(appIDs) {
			end = len(appIDs)
		}

		chunks = append(chunks, appIDs[i:end])
	}

	for _, chunk := range chunks {

		err = queue.ProduceAppPlayers(chunk)
		log.Err(err)
	}
}

func adminMemcache() {

	err := helpers.GetMemcache().DeleteAll()
	log.Err(err)

	err = sql.SetConfig(sql.ConfWipeMemcache+"-"+config.Config.Environment.Get(), strconv.FormatInt(time.Now().Unix(), 10))
	log.Err(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	log.Err(err)

	if err == nil {
		page.Send(adminWebsocket{sql.ConfWipeMemcache + "-" + config.Config.Environment.Get() + " complete"})
	}

	log.Info("Memcache wiped")
}

func adminDeleteBinLogs(r *http.Request) {

	name := r.URL.Query().Get("name")
	if name != "" {

		gorm, err := sql.GetMySQLClient()
		if err != nil {
			log.Err(err)
			return
		}

		gorm.Exec("PURGE BINARY LOGS TO '" + name + "'")
	}
}

func adminDev() {

	var err error

	log.Info("Started dev code")

	//
	err = sql.SetConfig(sql.ConfRunDevCode, strconv.FormatInt(time.Now().Unix(), 10))
	log.Err(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	log.Err(err)
	if err == nil {
		page.Send(adminWebsocket{sql.ConfRunDevCode + " complete"})
	}

	log.Info("Dev code run")
}

type statsRow struct {
	name       string
	count      int
	totalPrice map[steam.CountryCode]int
	totalScore float64
}

func (t statsRow) getMeanPrice() string {

	means := map[steam.CountryCode]float64{}

	for code, total := range t.totalPrice {
		means[code] = float64(total) / float64(t.count)
	}

	bytes, err := json.Marshal(means)
	log.Err(err)

	return string(bytes)
}

func (t statsRow) getMeanScore() float64 {
	return t.totalScore / float64(t.count)
}

type adminWebsocket struct {
	Message string `json:"message"`
}

func cronLogErr(interfaces ...interface{}) {
	log.Err(append(interfaces, log.LogNameCron, log.LogNameGameDB)...)
}

func cronLogInfo(interfaces ...interface{}) {
	log.Info(append(interfaces, log.LogNameCron, log.LogNameGameDB)...)
}
