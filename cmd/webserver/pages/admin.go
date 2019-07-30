package pages

import (
	"encoding/json"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/crons"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/go-chi/chi"
)

func AdminRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(middlewareAuthCheck())
	r.Use(middlewareAdminCheck())

	r.Get("/", adminHandler)
	r.Post("/", adminHandler)
	r.Get("/{option}", adminHandler)
	r.Post("/{option}", adminHandler)
	return r
}

func adminHandler(w http.ResponseWriter, r *http.Request) {

	option := chi.URLParam(r, "option")

	switch option {
	case "run-cron":
		go adminRunCron(r)
	case "refresh-all-apps":
		go adminQueueEveryApp()
	case "refresh-all-packages":
		go adminQueueEveryPackage()
	case "refresh-all-players":
		go adminQueueEveryPlayer()
	case "wipe-memcache":
		go adminClearMemcache()
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
	configKeys := []sql.ConfigType{
		sql.ConfAddedAllApps,
		sql.ConfWipeMemcache + "-" + sql.ConfigType(config.Config.Environment.Get()),
		sql.ConfRunDevCode,
		sql.ConfGarbageCollection,
		sql.ConfAddedAllAppPlayers,
		sql.ConfAddedAllPackages,
	}

	for _, v := range crons.CronRegister {
		configKeys = append(configKeys, v.Config())
	}

	configs, err := sql.GetConfigs(configKeys)
	log.Err(err, r)

	// Template
	t := adminTemplate{}
	t.fill(w, r, "Admin", "")
	t.Configs = configs
	t.Goroutines = runtime.NumGoroutine()
	t.Websockets = websockets.Pages
	t.Crons = crons.CronRegister

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

	gorm = gorm.Raw("SELECT * FROM information_schema.processlist where command != 'sleep'").Scan(&t.Queries)
	log.Err(gorm.Error, r)

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
	Websockets map[websockets.WebsocketPage]*websockets.Page
	Crons      map[crons.CronEnum]crons.CronInterface
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

func adminRunCron(r *http.Request) {

	c := r.URL.Query().Get("cron")

	cron := crons.CronRegister[crons.CronEnum(c)]
	cron.Work()

}

func adminQueueEveryApp() {

	var last = 0
	var keepGoing = true
	var apps steam.AppList
	var err error
	var count int

	for keepGoing {

		apps, b, err := helpers.GetSteam().GetAppList(1000, last, 0, "")
		err = helpers.AllowSteamCodes(err, b, nil)
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

	page := websockets.GetPage(websockets.PageAdmin)
	page.Send(websockets.AdminPayload{Message: string(sql.ConfAddedAllApps) + " complete"})

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

	page := websockets.GetPage(websockets.PageAdmin)
	page.Send(websockets.AdminPayload{Message: string(sql.ConfAddedAllPackages) + " complete"})

	log.Info(strconv.Itoa(len(packageIDs)) + " packages added to rabbit")
}

func adminQueueEveryPlayer() {

	log.Info("Queueing every player")

	players, err := mongo.GetPlayers(0, 0, mongo.D{{"_id", 1}}, nil, mongo.M{"_id": 1}, nil)
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

	page := websockets.GetPage(websockets.PageAdmin)
	page.Send(websockets.AdminPayload{Message: string(sql.ConfAddedAllPlayers) + " complete"})

	log.Info(strconv.Itoa(len(players)) + " players added to rabbit")
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

	if val := r.PostForm.Get("group-id"); val != "" {

		vals := strings.Split(val, ",")

		for _, val := range vals {

			err := queue.ProduceGroup([]string{val})
			log.Err(err, r)
		}
	}

	if val := r.PostForm.Get("apps-ts"); val != "" {

		log.Info("Queueing apps")

		ts, err := strconv.ParseInt(val, 10, 64)
		log.Err(err, r)
		if err == nil {

			apps, b, err := helpers.GetSteam().GetAppList(100000, 0, ts, "")
			err = helpers.AllowSteamCodes(err, b, nil)
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

func adminClearMemcache() {

	err := helpers.GetMemcache().DeleteAll()
	log.Err(err)

	err = sql.SetConfig(sql.ConfWipeMemcache+"-"+sql.ConfigType(config.Config.Environment.Get()), strconv.FormatInt(time.Now().Unix(), 10))
	log.Err(err)

	page := websockets.GetPage(websockets.PageAdmin)
	page.Send(websockets.AdminPayload{Message: string(sql.ConfWipeMemcache) + "-" + config.Config.Environment.Get() + " complete"})

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

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err)
		return
	}

	gorm = gorm.Select([]string{"id"})
	gorm = gorm.Where("id = ?", 10093)
	gorm = gorm.Order("id asc")

	// Get rows
	var bundles []sql.Bundle
	gorm = gorm.Find(&bundles)
	if gorm.Error != nil {
		log.Err(gorm.Error)
		return
	}

	for _, bundle := range bundles {

		builder := influxql.NewBuilder()
		builder.AddSelect("discount", "")
		builder.SetFrom(helpers.InfluxGameDB, helpers.InfluxRetentionPolicyAllTime.String(), helpers.InfluxMeasurementApps.String())
		builder.AddWhere("bundle_id", "=", bundle.ID)

		resp, err := helpers.InfluxQuery(builder.String())
		if err != nil {
			log.Err(err)
			continue
		}

		if len(resp.Results) == 0 || len(resp.Results[0].Series) == 0 || len(resp.Results[0].Series[0].Values) == 0 {
			log.Info(bundle.ID, "skipped")
			continue
		}

		type price struct {
			Time    time.Time `json:"time"`
			Percent int       `json:"percent"`
		}

		var prices []price
		var priceDocuments []mongo.Document

		// Convert influx response to slice
		for _, influxRow := range resp.Results[0].Series[0].Values {

			t, err := time.Parse(time.RFC3339, influxRow[0].(string))
			if err != nil {
				log.Err(err)
				continue
			}

			i, err := strconv.Atoi(influxRow[1].(json.Number).String())
			if err != nil {
				log.Err(err)
				continue
			}

			prices = append(prices, price{Time: t, Percent: i})
		}

		if len(prices) == 0 {
			log.Info(bundle.ID, "no prices")
			continue
		}

		// Sort prices, oldest first
		sort.Slice(prices, func(i, j int) bool {
			return prices[i].Time.Unix() < prices[j].Time.Unix()
		})

		// Save to mongo
		var last = 1 // A value that will never match the first price
		for _, v := range prices {

			bundle.SetDiscount(v.Percent)

			if v.Percent != last {

				document := mongo.BundlePrice{
					CreatedAt: v.Time,
					BundleID:  bundle.ID,
					Discount:  v.Percent,
				}

				priceDocuments = append(priceDocuments, document)
			}
		}

		_, err = mongo.InsertDocuments(mongo.CollectionBundlePrices, priceDocuments)

		err = bundle.Save()
		if err != nil {
			log.Err(err)
			continue
		}
	}

	//
	err = sql.SetConfig(sql.ConfRunDevCode, strconv.FormatInt(time.Now().Unix(), 10))
	log.Err(err)

	page := websockets.GetPage(websockets.PageAdmin)
	page.Send(websockets.AdminPayload{Message: string(sql.ConfRunDevCode) + " complete"})

	log.Info("Dev code run")
}
