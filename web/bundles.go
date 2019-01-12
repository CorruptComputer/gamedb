package web

import (
	"html/template"
	"net/http"
	"strconv"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
	"github.com/go-chi/chi"
)

func bundlesRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", bundlesHandler)
	r.Get("/ajax", bundlesAjaxHandler)
	r.Get("/{id}", bundleHandler)
	r.Get("/{id}/{slug}", bundleHandler)
	return r
}

func bundlesHandler(w http.ResponseWriter, r *http.Request) {

	total, err := db.CountBundles()
	log.Err(err, r)

	// Template
	t := bundlesTemplate{}
	t.Fill(w, r, "Bundles", "The last "+template.HTML(humanize.Comma(int64(total)))+" bundles to be updated.")

	err = returnTemplate(w, r, "bundles", t)
	log.Err(err, r)
}

type bundlesTemplate struct {
	GlobalTemplate
}

func bundlesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	query := DataTablesQuery{}
	err := query.FillFromURL(r.URL.Query())
	log.Err(err, r)

	//
	var wg sync.WaitGroup

	// Get apps
	var bundles []db.Bundle

	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		gorm, err := db.GetMySQLClient()
		if err != nil {

			log.Err(err, r)
			return
		}

		gorm = gorm.Model(&db.Bundle{})
		gorm = gorm.Select([]string{"id", "name", "updated_at", "discount", "app_ids", "package_ids"})

		gorm = query.SetOrderOffsetGorm(gorm, "", map[string]string{
			"0": "name",
			"1": "discount",
			"2": "JSON_LENGTH(app_ids)",
			"3": "JSON_LENGTH(package_ids)",
			"4": "updated_at",
		})

		gorm = gorm.Limit(100)
		gorm = gorm.Find(&bundles)

		log.Err(gorm.Error)

	}(r)

	// Get total
	var count int
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = db.CountBundles()
		log.Err(err, r)

	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(count)
	response.RecordsFiltered = strconv.Itoa(count)
	response.Draw = query.Draw

	for _, v := range bundles {
		response.AddRow(v.OutputForJSON())
	}

	response.output(w, r)
}
