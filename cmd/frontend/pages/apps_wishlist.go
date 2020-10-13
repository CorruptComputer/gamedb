package pages

import (
	"math"
	"net/http"
	"sync"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

func wishlistsRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", wishlistsHandler)
	r.Get("/games.json", wishlistAppsHandler)
	// r.Get("/tags.json", wishlistTagsHandler)
	return r
}

func wishlistsHandler(w http.ResponseWriter, r *http.Request) {

	t := wishlistsTemplate{}
	t.fill(w, r, "wishlists", "Wishlists", "Games on the most wishlists")

	returnTemplate(w, r, t)
}

type wishlistsTemplate struct {
	globalTemplate
}

func (t wishlistsTemplate) includes() []string {
	return []string{"includes/apps_header.gohtml"}
}

func wishlistAppsHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, true)

	var filter = bson.D{
		{"wishlist_count", bson.M{"$gt": 0}},
		{"wishlist_avg_position", bson.M{"$gt": 0}},
	}

	filter2 := filter
	search := query.GetSearchString("search")
	if search != "" {
		filter2 = append(filter2, bson.E{Key: "$text", Value: bson.M{"$search": search}})
	}

	var wg sync.WaitGroup
	var countLock sync.Mutex
	var code = session.GetProductCC(r)

	// Count
	var apps []mongo.App
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		columns := map[string]string{
			"1": "wishlist_count",
			"2": "wishlist_avg_position, wishlist_count desc",
			"3": "group_followers",
			"4": "prices." + string(code) + ".final",
			"5": "release_date_unix",
		}

		projection := bson.M{"_id": 1, "name": 1, "icon": 1, "wishlist_count": 1, "wishlist_avg_position": 1, "prices": 1, "group_followers": 1, "group_id": 1, "release_date_unix": 1, "release_state": 1}
		order := query.GetOrderMongo(columns)
		offset := query.GetOffset64()

		apps, err = mongo.GetApps(offset, 100, order, filter2, projection)
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Get filtered count
	var filtered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		countLock.Lock()
		filtered, err = mongo.CountDocuments(mongo.CollectionApps, filter2, 0)
		countLock.Unlock()
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Get count
	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		countLock.Lock()
		count, err = mongo.CountDocuments(mongo.CollectionApps, filter, 86400)
		countLock.Unlock()
		if err != nil {
			log.ErrS(err)
		}
	}()

	wg.Wait()

	//
	var response = datatable.NewDataTablesResponse(r, query, count, filtered, nil)
	for _, app := range apps {

		avgPosition := math.Round(app.WishlistAvgPosition*100) / 100

		response.AddRow([]interface{}{
			app.ID,                          // 0
			app.GetName(),                   // 1
			app.GetIcon(),                   // 2
			app.GetPath(),                   // 3
			app.WishlistCount,               // 4
			avgPosition,                     // 5
			app.GetFollowers(),              // 6
			helpers.GetAppStoreLink(app.ID), // 7
			app.ReleaseDateUnix,             // 8
			app.GetReleaseDateNice(),        // 9
			app.Prices.Get(code).GetFinal(), // 10
		})
	}

	returnJSON(w, r, response)
}
