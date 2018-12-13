package web

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
	"github.com/go-chi/chi"
)

func productKeysRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", productKeysHandler)
	r.Get("/ajax", productKeysAjaxHandler)
	return r
}

func productKeysHandler(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query()

	// Template
	t := productKeysTemplate{}
	t.Fill(w, r, "Product Keys", "Search extended and common product keys")
	t.Type = q.Get("type")
	t.Key = q.Get("key")
	t.Value = q.Get("value")

	if t.Type != "app" && t.Type != "package" {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid Type."})
		return
	}

	err := returnTemplate(w, r, "product_keys", t)
	log.Log(err)
}

type productKeysTemplate struct {
	GlobalTemplate
	Key   string
	Value string
	Type  string
}

func productKeysAjaxHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	query := DataTablesQuery{}
	err := query.FillFromURL(r.URL.Query())
	log.Log(err)

	//
	var code = session.GetCountryCode(r)
	var wg sync.WaitGroup
	var productType = query.GetSearchString("type")

	// Get products
	var products []extendedRow
	var recordsFiltered int
	wg.Add(1)
	go func() {

		defer wg.Done()

		gorm, err := db.GetMySQLClient()
		if err != nil {
			log.Log(err)
			return
		}

		if productType == "app" {
			gorm = gorm.Table("apps")
		} else if productType == "package" {
			gorm = gorm.Table("packages")
		} else {
			log.Log("no product type")
			return
		}

		// Search
		key := query.GetSearchString("key")
		if key == "" {
			return
		}
		value := query.GetSearchString("value")

		gorm = gorm.Select([]string{"id", "name", "icon", "extended->>'$." + key + "' as value"})

		if value == "" {
			gorm = gorm.Where("extended->>'$." + key + "' != ''")
		} else {
			gorm = gorm.Where("extended->>'$."+key+"' = ?", value)
		}

		// Count
		gorm = gorm.Count(&recordsFiltered)
		log.Log(gorm.Error)

		// Order, offset, limit
		gorm = gorm.Limit(100)
		gorm = query.SetOrderOffsetGorm(gorm, code, map[string]string{})
		gorm = gorm.Order("change_number_date desc")

		// Get rows
		gorm = gorm.Find(&products)
		log.Log(gorm.Error)
	}()

	// Get total
	var count int
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = db.CountApps()
		log.Log(err)

	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(count)
	response.RecordsFiltered = strconv.Itoa(recordsFiltered)
	response.Draw = query.Draw

	for _, v := range products {
		response.AddRow([]interface{}{
			v.ID,
			v.Name,
			v.GetIcon(),
			v.GetPath(productType),
			v.Value,
		})
	}

	response.output(w, r)
}

type extendedRow struct {
	ID    int
	Name  string
	Icon  string
	Value string
}

func (e extendedRow) GetIcon() string {
	return db.GetAppIcon(e.ID, e.Icon)
}

func (e extendedRow) GetPath(productType string) string {
	if productType == "app" {
		return db.GetAppPath(e.ID, e.Name)
	}
	return db.GetPackagePath(e.ID, e.Name)
}
