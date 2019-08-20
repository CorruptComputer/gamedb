package pages

import (
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
)

func BundleRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", bundleHandler)
	r.Get("/prices.json", bundlePricesAjaxHandler)
	r.Get("/{slug}", bundleHandler)
	return r
}

func bundleHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid bundle ID."})
		return
	}

	idx, err := strconv.Atoi(id)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid bundle ID: " + id})
		return
	}

	// todo, validate
	// if !db.IsValidAppID(idx) {
	// 	returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid bundle ID: " + id})
	// 	return
	// }

	// Get bundle
	bundle, err := sql.GetBundle(idx, []string{})
	if err != nil {

		if err == sql.ErrRecordNotFound {
			returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Sorry but we can not find this bundle."})
			return
		}

		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the bundle.", Error: err})
		return
	}

	//
	var wg sync.WaitGroup

	// Get apps
	var apps []sql.App
	wg.Add(1)
	go func(bundle sql.Bundle) {

		defer wg.Done()

		appIDs, err := bundle.GetAppIDs()
		if err != nil {
			log.Err(err, r)
			return
		}

		apps, err = sql.GetAppsByID(appIDs, []string{})
		log.Err(err, r)

		// Queue missing apps
		if len(appIDs) != len(apps) {
			for _, v := range appIDs {
				var found = false
				for _, vv := range apps {
					if v == vv.ID {
						found = true
						break
					}
				}

				if !found {
					queue.ProduceAppsWithPICS([]int{v})
				}
			}
		}
	}(bundle)

	// Get packages
	var packages []sql.Package
	wg.Add(1)
	go func() {

		defer wg.Done()

		appIDs, err := bundle.GetPackageIDs()
		if err != nil {
			log.Err(err, r)
			return
		}

		packages, err = sql.GetPackages(appIDs, []string{})
		log.Err(err, r)

	}()

	// Wait
	wg.Wait()

	// Template
	t := bundleTemplate{}
	for _, v := range apps {
		if v.Background != "" {
			t.setBackground(v, true, true)
			break
		}
	}
	t.fill(w, r, bundle.Name, "")
	t.addAssetHighCharts()
	t.Bundle = bundle
	t.Canonical = bundle.GetPath()
	t.Apps = apps
	t.Packages = packages

	//
	err = returnTemplate(w, r, "bundle", t)
	log.Err(err, r)
}

type bundleTemplate struct {
	GlobalTemplate
	Bundle   sql.Bundle
	Apps     []sql.App
	Packages []sql.Package
}

func bundlePricesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Err("invalid id", r)
		return
	}

	idx, err := strconv.Atoi(id)
	if err != nil {
		log.Err("invalid id", r)
		return
	}

	// Get prices
	pricesResp, err := mongo.GetBundlePrices(idx)
	if err != nil {
		log.Err(err, r)
		return
	}

	// Make JSON response
	var prices [][]int64

	for _, v := range pricesResp {
		prices = append(prices, []int64{v.CreatedAt.Unix() * 1000, int64(v.Discount)})
	}

	// Add current price
	price, err := sql.GetBundle(idx, []string{"discount"})
	log.Err(err)
	if err == nil {
		prices = append(prices, []int64{time.Now().Unix() * 1000, int64(price.Discount)})
	}

	// Sort prices for Highcharts
	sort.Slice(prices, func(i, j int) bool {
		return prices[i][0] < prices[j][0]
	})

	// Return
	err = returnJSON(w, r, prices)
	log.Err(err, r)
}
