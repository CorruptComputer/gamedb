package web

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/queue"
	"github.com/gamedb/website/session"
	"github.com/go-chi/chi"
)

func packageHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	idx, err := strconv.Atoi(id)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Invalid package ID.", Error: err})
		return
	}

	// Get package
	pack, err := db.GetPackage(idx, []string{})
	if err != nil {

		if err == db.ErrRecordNotFound {
			returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Sorry but we can not find this package."})
			return
		}

		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the package.", Error: err})
		return
	}

	// Redirect to correct slug
	if r.URL.Path != pack.GetPath() {
		http.Redirect(w, r, pack.GetPath(), 302)
		return
	}

	//
	var wg sync.WaitGroup

	var apps = map[int]db.App{}
	wg.Add(1)
	go func() {

		defer wg.Done()

		// Get apps
		appIDs, err := pack.GetAppIDs()
		log.Err(err, r)

		for _, v := range appIDs {
			apps[v] = db.App{ID: v}
		}

		appRows, err := db.GetAppsByID(appIDs, []string{"id", "name", "icon", "type", "platforms", "dlc"})
		log.Err(err, r)

		for _, v := range appRows {
			apps[v.ID] = v
		}

	}()

	// Make banners
	banners := make(map[string][]string)
	var primary []string

	// if pack.GetExtended() == "prerelease" {
	// 	primary = append(primary, "This package is intended for developers and publishers only.")
	// }

	if len(primary) > 0 {
		banners["primary"] = primary
	}

	// Wait
	wg.Wait()

	// Template
	t := packageTemplate{}
	t.Fill(w, r, pack.GetName(), "")
	t.Package = pack
	t.Apps = apps

	// Update news, reviews etc
	func() {

		if helpers.IsBot(r.UserAgent()) {
			return
		}

		if pack.UpdatedAt.Unix() > time.Now().Add(time.Hour * -24).Unix() {
			return
		}

		err = queue.QueuePackage([]int{pack.ID})
		if err != nil {
			log.Err(err, r)
		} else {
			t.addToast(Toast{Title: "Update", Message: "Package has been queued for an update"})
		}
	}()

	// Get price
	t.Price = db.GetPriceFormatted(pack, session.GetCountryCode(r))

	err = returnTemplate(w, r, "package", t)
	log.Err(err, r)
}

type packageTemplate struct {
	GlobalTemplate
	Package db.Package
	Apps    map[int]db.App
	Banners map[string][]string
	Price   db.ProductPriceFormattedStruct
}

func packagePricesAjaxHandler(w http.ResponseWriter, r *http.Request) {
	productPricesAjaxHandler(w, r, db.ProductTypePackage)
}
