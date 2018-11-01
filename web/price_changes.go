package web

import (
	"net/http"
	"strconv"
	"sync"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/logging"
)

func PriceChangesHandler(w http.ResponseWriter, r *http.Request) {

	t := priceChangesTemplate{}
	t.Fill(w, r, "Price Changes")

	returnTemplate(w, r, "price_changes", t)
	return
}

type priceChangesTemplate struct {
	GlobalTemplate
}

func PriceChangesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	query.FillFromURL(r.URL.Query())

	//
	var wg sync.WaitGroup

	// Get ranks
	var priceChanges []db.ProductPrice

	wg.Add(1)
	go func() {

		client, ctx, err := db.GetDSClient()
		if err == nil {

			q := datastore.NewQuery(db.KindProductPrice).Limit(100).Order("-created_at")
			q = q.Filter("currency =", steam.CountryUS)

			q, err = query.SetOffsetDS(q)
			if err == nil {
				_, err = client.GetAll(ctx, q, &priceChanges)
			}
		}

		logging.Error(err)

		wg.Done()
	}()

	// Get total
	var total int
	wg.Add(1)
	go func() {

		total = 10000

		wg.Done()
	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(total)
	response.RecordsFiltered = strconv.Itoa(total)
	response.Draw = query.Draw

	for _, v := range priceChanges {

		response.AddRow(v.OutputForJSON())
	}

	response.output(w)
}
