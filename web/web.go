package web

import (
	"bytes"
	"encoding/json"
	"html/template"
	"math"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/steam-go/steam"
	"github.com/derekstavis/go-qs"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
	"github.com/gamedb/website/websockets"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/jinzhu/gorm"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Called from main
func Init() {

	session.Init()

	InitChat()
	InitCommits()
}

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if viper.GetString("ENV") == string(log.EnvProd) {
			log.Info(log.ServiceGoogle, log.LogNameRequests, r.Method+" "+r.URL.Path)
		}
		next.ServeHTTP(w, r)
	})
}

func middlewareTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		r.Header.Set("start-time", strconv.FormatInt(time.Now().UnixNano(), 10))

		next.ServeHTTP(w, r)
	})
}

func middlewareCors() func(next http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins: []string{viper.GetString("DOMAIN")}, // Use this to allow specific origin hosts
		AllowedMethods: []string{"GET", "POST"},
	}).Handler
}

func Serve() error {

	r := chi.NewRouter()

	r.Use(middlewareTime)
	r.Use(middlewareCors())
	r.Use(middleware.RealIP)
	r.Use(middleware.DefaultCompress)
	r.Use(middleware.RedirectSlashes)
	r.Use(middlewareLog)

	// Pages
	r.Get("/", homeRedirectHandler)
	r.Get("/commits", commitsHandler)
	r.Get("/coop", coopHandler)
	r.Get("/discounts", discountsHandler)
	r.Get("/developers", statsDevelopersHandler)
	r.Get("/donate", donateHandler)
	r.Get("/esi/header", headerHandler)
	r.Get("/genres", statsGenresHandler)
	r.Get("/health-check", healthCheckHandler)
	r.Get("/info", infoHandler)
	r.Get("/logout", logoutHandler)
	r.Get("/news", newsHandler)
	r.Get("/news/ajax", newsAjaxHandler)
	r.Get("/publishers", statsPublishersHandler)
	r.Get("/tags", statsTagsHandler)
	r.Get("/websocket/{id:[a-z]+}", websockets.WebsocketsHandler)
	r.Mount("/admin", adminRouter())
	r.Mount("/apps", appsRouter())
	r.Mount("/bundles", bundlesRouter())
	r.Mount("/changes", changesRouter())
	r.Mount("/chat", chatRouter())
	r.Mount("/contact", contactRouter())
	r.Mount("/depots", depotsRouter())
	r.Mount("/experience", experienceRouter())
	r.Mount("/free-games", freeGamesRouter())
	r.Mount("/games", gamesRouter())
	r.Mount("/login", loginRouter())
	r.Mount("/packages", packagesRouter())
	r.Mount("/players", playersRouter())
	r.Mount("/price-changes", priceChangeRouter())
	r.Mount("/product-keys", productKeysRouter())
	r.Mount("/queues", queuesRouter())
	r.Mount("/settings", settingsRouter())
	r.Mount("/stats", statsRouter())
	r.Mount("/upcoming", upcomingRouter())

	// Files
	r.Get("/browserconfig.xml", rootFileHandler)
	r.Get("/robots.txt", rootFileHandler)
	r.Get("/sitemap.xml", siteMapHandler)
	r.Get("/site.webmanifest", rootFileHandler)

	// File server
	fileServer(r)

	// 404
	r.NotFound(error404Handler)

	return http.ListenAndServe("0.0.0.0:"+viper.GetString("PORT"), r)
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func fileServer(r chi.Router) {

	path := "/assets"

	if strings.ContainsAny(path, "{}*") {
		log.Info("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(http.Dir(filepath.Join(viper.GetString("PATH"), "assets"))))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}

func setNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	w.Header().Set("Expires", "0")                                         // Proxies.
}

func returnJSON(w http.ResponseWriter, r *http.Request, bytes []byte) (err error) {

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Language", string(session.GetCountryCode(r))) // Used for varnish hash

	_, err = w.Write(bytes)
	return err
}

func returnTemplate(w http.ResponseWriter, r *http.Request, page string, pageData interface{}) (err error) {

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Language", string(session.GetCountryCode(r))) // Used for varnish hash
	w.WriteHeader(200)

	folder := viper.GetString("PATH")
	t, err := template.New("t").Funcs(getTemplateFuncMap()).ParseFiles(
		folder+"/templates/_header.gohtml",
		folder+"/templates/_header_esi.gohtml",
		folder+"/templates/_footer.gohtml",
		folder+"/templates/_stats_header.gohtml",
		folder+"/templates/_deals_header.gohtml",
		folder+"/templates/_apps_header.gohtml",
		folder+"/templates/_flashes.gohtml",
		folder+"/templates/"+page+".gohtml",
	)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Something has gone wrong!", Error: err})
		return err
	}

	// Write a respone
	buf := &bytes.Buffer{}
	err = t.ExecuteTemplate(buf, page, pageData)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "Something has gone wrong!", Error: err})
		return err
	}

	_, err = buf.WriteTo(w)
	log.Err(err)

	return nil
}

func returnErrorTemplate(w http.ResponseWriter, r *http.Request, data errorTemplate) {

	if data.Title == "" {
		data.Title = "Error " + strconv.Itoa(data.Code)
	}

	if data.Code == 0 {
		data.Code = 500
	}

	log.Err(data.Error)

	data.Fill(w, r, "Error", "Something has gone wrong!")

	w.WriteHeader(data.Code)

	err := returnTemplate(w, r, "error", data)
	log.Err(err)
}

type errorTemplate struct {
	GlobalTemplate
	Title   string
	Message string
	Code    int
	Error   error
}

func getTemplateFuncMap() map[string]interface{} {
	return template.FuncMap{
		//"title":  func(a string) string { return strings.Title(a) },
		//"slug":   func(a string) string { return slug.Make(a) },
		//"unix":       func(t time.Time) int64 { return t.Unix() },
		//"contains":   func(a string, b string) bool { return strings.Contains(a, b) },
		"join": func(a []string) string { return strings.Join(a, ", ") },
		"joinInt": func(a []int) string {
			var join []string
			for _, v := range a {
				join = append(join, strconv.Itoa(v))
			}
			return strings.Join(join, ", ")
		},
		"comma":  func(a int) string { return humanize.Comma(int64(a)) },
		"commaf": func(a float64) string { return humanize.Commaf(a) },
		"apps": func(a []int, appsMap map[int]db.App) template.HTML {
			var apps []string
			for _, v := range a {
				apps = append(apps, "<a href=\"/apps/"+strconv.Itoa(v)+"\">"+appsMap[v].GetName()+"</a>")
			}
			return template.HTML("Apps: " + strings.Join(apps, ", "))
		},
		"packages": func(a []int, packagesMap map[int]db.Package) template.HTML {
			var packages []string
			for _, v := range a {
				packages = append(packages, "<a href=\"/packages/"+strconv.Itoa(v)+"\">"+packagesMap[v].GetName()+"</a>")
			}
			return template.HTML("Packages: " + strings.Join(packages, ", "))
		},
		"tags": func(a []db.Tag) template.HTML {

			sort.Slice(a, func(i, j int) bool {
				return a[i].Name < a[j].Name
			})

			var tags []string
			for _, v := range a {
				tags = append(tags, "<a class=\"badge badge-success\" href=\"/apps?tags="+strconv.Itoa(v.ID)+"\">"+v.GetName()+"</a>")
			}
			return template.HTML(strings.Join(tags, " "))
		},
		"genres": func(a []steam.AppDetailsGenre) template.HTML {

			sort.Slice(a, func(i, j int) bool {
				return a[i].Description < a[j].Description
			})

			var genres []string
			for _, v := range a {
				genres = append(genres, "<a class=\"badge badge-success\" href=\"/apps?genres="+strconv.Itoa(v.ID)+"\">"+v.Description+"</a>")
			}
			return template.HTML(strings.Join(genres, " "))
		},
		"startsWith": func(a string, b string) bool { return strings.HasPrefix(a, b) },
		"endsWith":   func(a string, b string) bool { return strings.HasSuffix(a, b) },
		"max":        func(a int, b int) float64 { return math.Max(float64(a), float64(b)) },
		"json": func(v interface{}) (string, error) {
			b, err := json.Marshal(v)
			log.Err(err)
			return string(b), err
		},
	}
}

// GlobalTemplate is added to every other template
type GlobalTemplate struct {
	Title       string        // Page title
	Description template.HTML // Page description
	Path        string        // URL path
	Env         string        // Environment

	// Session
	userName           string
	userEmail          string
	userID             int
	userLevel          int
	userCountry        steam.CountryCode
	userCurrencySymbol string

	// Session
	flashesGood []interface{}
	flashesBad  []interface{}
	session     map[string]string

	//
	toasts []Toast

	//
	request *http.Request // Internal
}

func (t *GlobalTemplate) Fill(w http.ResponseWriter, r *http.Request, title string, description template.HTML) {

	var err error

	t.request = r

	t.Title = title
	t.Description = description
	t.Env = viper.GetString("ENV")
	t.Path = r.URL.Path

	// User ID
	id, err := session.Read(r, session.PlayerID)
	log.Err(err)

	if id == "" {
		t.userID = 0
	} else {
		t.userID, err = strconv.Atoi(id)
		log.Err(err)
	}

	// User name
	t.userName, err = session.Read(r, session.PlayerName)
	log.Err(err)

	// Email
	t.userEmail, err = session.Read(r, session.UserEmail)
	log.Err(err)

	// Level
	level, err := session.Read(r, session.PlayerLevel)
	log.Err(err)

	if level == "" {
		t.userLevel = 0
	} else {
		t.userLevel, err = strconv.Atoi(level)
		log.Err(err)
	}

	// Country
	var code = session.GetCountryCode(r)
	t.userCountry = code

	locale, err := helpers.GetLocaleFromCountry(code)
	log.Err(err)

	t.userCurrencySymbol = locale.CurrencySymbol

	// Flashes
	t.flashesGood, err = session.GetGoodFlashes(w, r)
	log.Err(err)

	t.flashesBad, err = session.GetBadFlashes(w, r)
	log.Err(err)

	// All session data, todo, remove this, security etc
	t.session, err = session.ReadAll(r)
	log.Err(err)
}

func (t GlobalTemplate) GetUserJSON() string {

	stringMap := map[string]interface{}{
		"userID":         strconv.Itoa(t.userID), // Too long for JS int
		"userLevel":      t.userLevel,
		"userName":       t.userName,
		"userEmail":      t.userEmail,
		"isLoggedIn":     t.isLoggedIn(),
		"isLocal":        t.isLocal(),
		"isAdmin":        t.isAdmin(),
		"showAds":        t.showAds(),
		"country":        t.userCountry,
		"currencySymbol": t.userCurrencySymbol,
		"flashesGood":    t.flashesGood,
		"flashesBad":     t.flashesBad,
		"toasts":         t.toasts,
		"session":        t.session,
	}

	b, err := json.Marshal(stringMap)
	log.Err(err)

	return string(b)
}

func (t GlobalTemplate) GetFooterText() (text string) {

	ts := time.Now()
	dayint, err := strconv.Atoi(ts.Format("2"))
	log.Err(err)

	text = "Page created on " + ts.Format("Mon") + " the " + humanize.Ordinal(dayint) + " @ " + ts.Format("15:04:05")

	// Get cashed
	if t.IsCacheHit() {
		text += " from cache"
	}

	// Get time
	startTimeString := t.request.Header.Get("start-time")
	if startTimeString == "" {
		return text
	}

	startTimeInt, err := strconv.ParseInt(startTimeString, 10, 64)
	if err != nil {
		log.Err(err)
		return text
	}

	d := time.Duration(time.Now().UnixNano() - startTimeInt)

	return text + " in " + d.String()
}

func (t GlobalTemplate) IsCacheHit() bool {
	return t.request.Header.Get("X-Cache") == "HIT"
}

func (t GlobalTemplate) IsFromVarnish() bool {
	return t.request.Header.Get("X-From-Varnish") == "true"
}

func (t GlobalTemplate) isLoggedIn() bool {
	return t.userID > 0
}

func (t GlobalTemplate) isLocal() bool {
	return t.Env == string(log.EnvLocal)
}

func (t GlobalTemplate) isAdmin() bool {
	return t.request.Header.Get("Authorization") != ""
}

func (t GlobalTemplate) showAds() bool {
	return !t.isLocal()
}

func (t *GlobalTemplate) addToast(toast Toast) {
	t.toasts = append(t.toasts, toast)
}

// DataTablesAjaxResponse
type DataTablesAjaxResponse struct {
	Draw            string          `json:"draw"`
	RecordsTotal    string          `json:"recordsTotal"`
	RecordsFiltered string          `json:"recordsFiltered"`
	Data            [][]interface{} `json:"data"`
}

func (t *DataTablesAjaxResponse) AddRow(row []interface{}) {
	t.Data = append(t.Data, row)
}

func (t DataTablesAjaxResponse) output(w http.ResponseWriter, r *http.Request) {

	if len(t.Data) == 0 {
		t.Data = make([][]interface{}, 0)
	}

	bytesx, err := json.Marshal(t)
	log.Err(err)

	err = returnJSON(w, r, bytesx)
	log.Err(err)
}

// DataTablesQuery
type DataTablesQuery struct {
	Draw   string
	Order  map[string]map[string]interface{}
	Start  string
	Search map[string]interface{}
	Time   string `mapstructure:"_"`
}

func (q *DataTablesQuery) FillFromURL(url url.Values) (err error) {

	// Convert string into map
	queryMap, err := qs.Unmarshal(url.Encode())
	if err != nil {
		return err
	}

	// Convert map into struct
	err = mapstructure.Decode(queryMap, q)
	if err != nil {
		return err
	}

	return nil
}

func (q DataTablesQuery) GetSearchString(k string) (search string) {

	if val, ok := q.Search[k]; ok {
		if ok && val != "" {
			return val.(string)
		}
	}

	return ""
}

func (q DataTablesQuery) GetSearchSlice(k string) (search []string) {

	if val, ok := q.Search[k]; ok {
		if val != "" {
			for _, v := range val.([]interface{}) {
				search = append(search, v.(string))
			}
		}
	}

	return search
}

func (q DataTablesQuery) GetOrderSQL(columns map[string]string, code steam.CountryCode) (order string) {

	var ret []string

	for _, v := range q.Order {

		if col, ok := v["column"].(string); ok {
			if ok {

				if dir, ok := v["dir"].(string); ok {
					if ok {

						if col, ok := columns[col]; ok {
							if ok {

								if col == "price" {
									col = "JSON_EXTRACT(prices, \"$." + string(code) + ".final\")"
								}

								if dir == "asc" || dir == "desc" {
									ret = append(ret, col+" "+dir)
								}
							}
						}
					}
				}
			}
		}
	}

	return strings.Join(ret, ", ")
}

func (q DataTablesQuery) GetOrderDS(columns map[string]string, signed bool) (order string) {

	for _, v := range q.Order {

		if col, ok := v["column"].(string); ok {
			if ok {

				if dir, ok := v["dir"].(string); ok {
					if ok {

						if col, ok := columns[col]; ok {
							if ok {

								if dir == "desc" && signed {
									col = "-" + col
								}
								return col
							}
						}
					}
				}
			}
		}
	}

	return ""
}

func (q DataTablesQuery) SetOrderOffsetGorm(db *gorm.DB, code steam.CountryCode, columns map[string]string) *gorm.DB {

	db = db.Order(q.GetOrderSQL(columns, code))
	db = db.Offset(q.Start)

	return db
}

func (q DataTablesQuery) SetOrderOffsetDS(qu *datastore.Query, columns map[string]string) (*datastore.Query, error) {

	qu, err := q.SetOffsetDS(qu)
	if err != nil {
		return qu, err
	}

	order := q.GetOrderDS(columns, true)
	if order != "" {
		qu = qu.Order(order)
	}

	return qu, nil
}

func (q DataTablesQuery) SetOffsetDS(qu *datastore.Query) (*datastore.Query, error) {

	i, err := strconv.Atoi(q.Start)
	if err != nil {
		return qu, err
	}

	qu = qu.Offset(i)

	return qu, nil
}

// Toasts
type Toast struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	Link    string `json:"link"`
	Theme   string `json:"theme"`
	Timeout int    `json:"timeout"`
}
