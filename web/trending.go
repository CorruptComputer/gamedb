package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
	"github.com/gamedb/website/sql"
	"github.com/go-chi/chi"
)

func trendingRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", trendingHandler)
	r.Get("/trending.json", trendingAjaxHandler)
	r.Get("/charts.json", trendingChartsAjaxHandler)
	return r
}

func trendingHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	// Template
	t := trendingTemplate{}
	t.fill(w, r, "Trending", "")
	t.addAssetHighCharts()

	t.Apps, err = countUpcomingApps()
	log.Err(err, r)

	err = returnTemplate(w, r, "trending", t)
	log.Err(err, r)
}

type trendingTemplate struct {
	GlobalTemplate
	Apps int
}

func trendingAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err, r)
		return
	}

	columns := map[string]string{
		"2": "player_trend",
	}

	gorm = gorm.Model(sql.App{})
	gorm = gorm.Select([]string{"id", "name", "icon", "prices", "player_trend"})
	gorm = gorm.Order(query.getOrderSQL(columns, session.GetCountryCode(r)))

	// Count before limitting
	// gorm.Count(&count)
	// log.Err(gorm.Error, r)

	gorm = gorm.Limit(50)
	gorm = gorm.Offset(query.getOffset())

	var apps []sql.App
	gorm = gorm.Find(&apps)
	log.Err(gorm.Error, r)

	var code = session.GetCountryCode(r)

	count, err := sql.CountApps()
	log.Err(err)

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(count)
	response.RecordsFiltered = strconv.Itoa(count)
	response.Draw = query.Draw

	for _, app := range apps {
		response.AddRow([]interface{}{
			app.ID,                                 // 0
			app.GetName(),                          // 1
			app.GetIcon(),                          // 2
			app.GetPath(),                          // 3
			sql.GetPriceFormatted(app, code).Final, // 5
			app.PlayerTrend,                        // 6
		})
	}

	response.output(w, r)
}

func trendingChartsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	idsString := r.URL.Query().Get("ids")
	idsSlice := strings.Split(idsString, ",")

	if len(idsSlice) == 0 {
		return
	}

	if len(idsSlice) > 50 {
		idsSlice = idsSlice[0:50]
	}

	var or []string
	for _, v := range idsSlice {
		v = strings.TrimSpace(v)
		if v != "" {
			or = append(or, `"app_id" = '`+v+`'`)
		}
	}

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.SetFrom("GameDB", "alltime", "apps")
	builder.AddWhere("time", ">", "NOW()-7d")
	builder.AddWhereRaw("(" + strings.Join(or, " OR ") + ")")
	builder.AddGroupByTime("1h")
	builder.AddGroupBy("app_id")
	builder.SetFillNone()

	resp, err := sql.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	ret := map[string]sql.HighChartsJson{}
	if len(resp.Results) > 0 {
		for _, v := range resp.Results[0].Series {
			ret[v.Tags["app_id"]] = sql.InfluxResponseToHighCharts(v)
		}
	}

	b, err := json.Marshal(ret)
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
