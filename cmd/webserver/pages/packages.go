package pages

import (
	"html/template"
	"net/http"
	"strconv"
	"sync"

	"github.com/gamedb/gamedb/cmd/webserver/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

func PackagesRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", packagesHandler)
	r.Get("/packages.json", packagesAjaxHandler)
	r.Mount("/{id}", PackageRouter())
	return r
}

func packagesHandler(w http.ResponseWriter, r *http.Request) {

	total, err := mongo.CountDocuments(mongo.CollectionPackages, nil, 0)
	if err != nil {
		log.Err(err, r)
	}

	// Template
	t := packagesTemplate{}
	t.fill(w, r, "Packages", "The last "+template.HTML(helpers.ShortHandNumber(total))+" packages to be updated.")
	t.addAssetChosen()

	returnTemplate(w, r, "packages", t)
}

type packagesTemplate struct {
	GlobalTemplate
}

func packagesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, false)

	//
	var code = helpers.GetProductCC(r)
	var wg sync.WaitGroup
	var countLock sync.Mutex
	var filter = bson.D{}

	// Status
	status := query.GetSearchSlice("status")
	if len(status) > 0 {

		a := bson.A{}
		for _, v := range status {
			i, err := strconv.Atoi(v)
			if err == nil {
				a = append(a, i)
			}
		}

		filter = append(filter, bson.E{Key: "status", Value: bson.M{"$in": a}})
	}

	// Platforms
	platforms := query.GetSearchSlice("platform")
	if len(platforms) > 0 {

		a := bson.A{}
		for _, v := range platforms {
			a = append(a, v)
		}

		filter = append(filter, bson.E{Key: "platforms", Value: bson.M{"$in": a}})
	}

	// License
	license := query.GetSearchSlice("license")
	if len(license) > 0 {

		a := bson.A{}
		for _, v := range license {
			i, err := strconv.Atoi(v)
			if err == nil {
				a = append(a, i)
			}
		}

		filter = append(filter, bson.E{Key: "license_type", Value: bson.M{"$in": a}})
	}

	// billing
	billing := query.GetSearchSlice("billing")
	if len(billing) > 0 {

		a := bson.A{}
		for _, v := range billing {
			i, err := strconv.Atoi(v)
			if err == nil {
				a = append(a, i)
			}
		}

		filter = append(filter, bson.E{Key: "billing_type", Value: bson.M{"$in": a}})
	}

	// Get packages
	var packages []mongo.Package
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		var projection = bson.M{"id": 1, "name": 1, "apps_count": 1, "change_number_date": 1, "prices": 1, "icon": 1}
		var sortCols = map[string]string{
			"1": "prices." + string(code) + ".final",
			"2": "prices." + string(code) + ".discount_percent",
			"3": "apps_count",
			"4": "change_number_date",
		}

		packages, err = mongo.GetPackages(query.GetOffset64(), 100, query.GetOrderMongo(sortCols), filter, projection, nil)
		if err != nil {
			log.Err(err, r)
			return
		}
	}(r)

	// Get total
	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		countLock.Lock()
		count, err = mongo.CountDocuments(mongo.CollectionPackages, nil, 0)
		countLock.Unlock()
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Get filtered
	var filtered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		countLock.Lock()
		filtered, err = mongo.CountDocuments(mongo.CollectionPackages, filter, 0)
		countLock.Unlock()
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Wait
	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, count, filtered)
	for _, v := range packages {
		response.AddRow(v.OutputForJSON(code))
	}

	returnJSON(w, r, response)
}
