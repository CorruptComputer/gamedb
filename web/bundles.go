package web

import (
	"html/template"
	"net/http"
	"strconv"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
	"github.com/go-chi/chi"
)

func bundlesRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", bundlesHandler)
	r.Get("/ajax", bundlesAjaxHandler)
	//r.Get("/{id}", bundleHandler)
	//r.Get("/{id}/{slug}", vHandler)
	return r
}

func bundlesHandler(w http.ResponseWriter, r *http.Request) {

	total, err := db.CountPackages()
	log.Err(err)

	// Template
	t := bundlesTemplate{}
	t.Fill(w, r, "Packages", "The last "+template.HTML(humanize.Comma(int64(total)))+" packages to be updated.")

	err = returnTemplate(w, r, "bundles", t)
	log.Err(err)
}

type bundlesTemplate struct {
	GlobalTemplate
}

func bundlesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	query := DataTablesQuery{}
	err := query.FillFromURL(r.URL.Query())
	log.Err(err)

	//
	var code = session.GetCountryCode(r)
	var wg sync.WaitGroup

	// Get apps
	var packages []db.Package

	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		gorm, err := db.GetMySQLClient()
		if err != nil {

			log.Err(err)
			return
		}

		gorm = gorm.Model(&db.Package{})
		gorm = gorm.Select([]string{"id", "name", "apps_count", "change_number_date", "prices", "coming_soon", "icon"})

		gorm = query.SetOrderOffsetGorm(gorm, code, map[string]string{
			"0": "name",
			"2": "apps_count",
			"3": "price",
			"4": "change_number_date",
		})

		gorm = gorm.Limit(100)
		gorm = gorm.Find(&packages)

		log.Err(gorm.Error)

	}(r)

	// Get total
	var count int
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = db.CountPackages()
		log.Err(err)

	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(count)
	response.RecordsFiltered = strconv.Itoa(count)
	response.Draw = query.Draw

	for _, v := range packages {
		response.AddRow(v.OutputForJSON(code))
	}

	response.output(w, r)
}
