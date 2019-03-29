package web

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gamedb/website/log"
	"github.com/gamedb/website/mongo"
	"github.com/gamedb/website/session"
	"github.com/gamedb/website/sql"
	"github.com/go-chi/chi"
)

func priceChangeRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", priceChangesHandler)
	r.Get("/price-changes.json", priceChangesAjaxHandler)
	return r
}

func priceChangesHandler(w http.ResponseWriter, r *http.Request) {

	t := priceChangesTemplate{}
	t.fill(w, r, "Price Changes", "Pick up a bargain.")
	t.addAssetChosen()
	t.addAssetSlider()

	price, err := sql.GetMostExpensiveApp(session.GetCountryCode(r))
	log.Err(err, r)

	// Convert dollars to cents
	t.ExpensiveApp = int(math.Ceil(float64(price) / 100))

	err = returnTemplate(w, r, "price_changes", t)
	log.Err(err, r)
}

type priceChangesTemplate struct {
	GlobalTemplate
	ExpensiveApp int
}

func priceChangesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	setCacheHeaders(w, 0)

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	//
	var wg sync.WaitGroup

	// Get ranks
	var priceChanges []mongo.ProductPrice

	var code = session.GetCountryCode(r)

	var dateLimit = time.Now().AddDate(0, 0, -30)

	var filter = mongo.M{
		"currency":   string(code),
		"created_at": mongo.M{"$gt": dateLimit},
	}

	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		priceChanges, err = mongo.GetPrices(query.getOffset64(), filter)
		log.Err(err, r)
	}(r)

	// Get filtered
	var filtered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		filtered, err = mongo.CountDocuments(mongo.CollectionProductPrices, filter)
		log.Err(err, r)
	}()

	// Get total
	var total int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionProductPrices, mongo.M{
			"created_at": mongo.M{"$gt": dateLimit},
		})
		log.Err(err, r)
	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.FormatInt(total, 10)
	response.RecordsFiltered = strconv.FormatInt(filtered, 10)
	response.Draw = query.Draw

	for _, v := range priceChanges {

		response.AddRow(v.OutputForJSON())
	}

	response.output(w, r)
}
