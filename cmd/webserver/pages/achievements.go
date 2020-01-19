package pages

import (
	"net/http"
	"sync"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

func AchievementsRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", achievementsHandler)
	r.Get("/achievements.json", achievementsAjaxHandler)
	return r
}

func achievementsHandler(w http.ResponseWriter, r *http.Request) {

	t := GlobalTemplate{}
	t.fill(w, r, "Achievements", "")

	returnTemplate(w, r, "achievements", t)
}

func achievementsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	if err != nil {
		log.Err(err, r)
	}

	query.limit(r)

	var wg sync.WaitGroup
	var count int64
	var filtered int64
	var apps []mongo.App
	var filter = bson.D{{"achievements_count", bson.M{"$gt": 0}}}
	var countLock sync.Mutex

	wg.Add(1)
	go func() {

		defer wg.Done()

		var filter2 = filter

		var search = query.getSearchString("search")
		if search != "" {
			filter2 = append(filter2, bson.E{Key: "$text", Value: bson.M{"$search": search}})
		}

		var columns = map[string]string{
			"1": "achievements_count",
			"2": "achievements_average_completion",
		}

		var projection = bson.M{"id": 1, "name": 1, "icon": 1, "achievements_5": 1, "achievements_count": 1, "achievements_average_completion": 1, "prices": 1}
		var sort = query.getOrderMongo(columns)

		var err error
		apps, err = mongo.GetApps(query.getOffset64(), 100, sort, filter2, projection, nil)
		if err != nil {
			log.Err(err)
		}

		countLock.Lock()
		filtered, err = mongo.CountDocuments(mongo.CollectionApps, filter2, 0)
		countLock.Unlock()
		if err != nil {
			log.Err(err, r)
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		countLock.Lock()
		count, err = mongo.CountDocuments(mongo.CollectionApps, filter, 60*60*24)
		countLock.Unlock()
		if err != nil {
			log.Err(err, r)
		}
	}()

	wg.Wait()

	//
	response := DataTablesAjaxResponse{}
	response.RecordsTotal = count
	response.RecordsFiltered = filtered
	response.Draw = query.Draw
	response.limit(r)

	var code = helpers.GetProductCC(r)

	for _, app := range apps {
		response.AddRow([]interface{}{
			app.ID,                            // 0
			app.GetName(),                     // 1
			app.GetIcon(),                     // 2
			app.GetPath() + "#achievements",   // 3
			app.Prices.Get(code).GetFinal(),   // 4
			app.AchievementsCount,             // 5
			app.AchievementsAverageCompletion, // 6
			app.Achievements5,                 // 7
		})
	}

	response.output(w, r)
}
