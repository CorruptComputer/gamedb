package pages

import (
	"net/http"
	"regexp"
	"sync"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

func GroupsRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", groupsHandler)
	r.Get("/groups.json", groupsTrendingAjaxHandler)
	r.Mount("/{id}", GroupRouter())
	return r
}

func groupsHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	t := groupsTemplate{}
	t.fill(w, r, "Groups", "A database of all Steam groups")

	count, err := mongo.CountDocuments(mongo.CollectionGroups, nil, 0)
	log.Err(err, r)

	t.Count = helpers.ShortHandNumber(count)

	returnTemplate(w, r, "groups", t)
}

type groupsTemplate struct {
	GlobalTemplate
	Count string
}

func groupsTrendingAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := newDataTableQuery(r, true)

	// Filter
	var filter = bson.D{
		{Key: "type", Value: helpers.GroupTypeGroup},
	}
	var unfiltered = filter

	search := helpers.RegexNonAlphaNumericSpace.ReplaceAllString(query.getSearchString("search"), "")
	if len(search) > 0 {

		quoted := regexp.QuoteMeta(search)

		filter = append(filter, bson.E{Key: "$or", Value: bson.A{
			bson.M{"name": bson.M{"$regex": quoted, "$options": "i"}},
			bson.M{"abbreviation": bson.M{"$regex": quoted, "$options": "i"}},
			bson.M{"url": bson.M{"$regex": quoted, "$options": "i"}},
		}})
	}

	//
	var wg sync.WaitGroup

	// Get groups
	var groups []mongo.Group
	wg.Add(1)
	go func() {

		defer wg.Done()

		columns := map[string]string{
			"1": "members",
			"2": "trending",
		}

		var err error
		groups, err = mongo.GetGroups(100, query.getOffset64(), query.getOrderMongo(columns), filter, nil)
		if err != nil {
			log.Err(err, r)
			return
		}
	}()

	// Get total
	var total int64
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionGroups, unfiltered, 60*60*6)
		log.Err(err, r)
	}(r)

	var totalFiltered int64
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		totalFiltered, err = mongo.CountDocuments(mongo.CollectionGroups, filter, 60*60)
		log.Err(err, r)
	}(r)

	wg.Wait()

	response := DataTablesResponse{}
	response.RecordsTotal = total
	response.RecordsFiltered = totalFiltered
	response.Draw = query.Draw
	response.limit(r)

	for _, group := range groups {
		response.AddRow([]interface{}{
			group.ID,                           // 0
			group.GetName(),                    // 1
			group.GetPath(),                    // 2
			group.GetIcon(),                    // 3
			group.Headline,                     // 4
			group.Members,                      // 5
			group.URL,                          // 6
			group.Type,                         // 7
			group.GetURL(),                     // 8
			group.Error != "",                  // 9
			helpers.TrendValue(group.Trending), // 10
		})
	}

	response.output(w, r)
}
