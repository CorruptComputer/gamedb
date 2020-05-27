package pages

import (
	"net/http"
	"regexp"
	"sync"

	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	elasticHelpers "github.com/gamedb/gamedb/pkg/elastic"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/bson"
)

func GroupsRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", groupsHandler)
	r.Get("/groups.json", groupsAjaxHandler)
	r.Mount("/{id}", GroupRouter())
	return r
}

func groupsHandler(w http.ResponseWriter, r *http.Request) {

	t := groupsTemplate{}
	t.fill(w, r, "Groups", "All the groups on Steam")
	t.addAssetChosen()

	returnTemplate(w, r, "groups", t)
}

type groupsTemplate struct {
	GlobalTemplate
}

func groupsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	var query = datatable.NewDataTableQuery(r, true)

	var wg sync.WaitGroup

	// Get groups
	var groups []elasticHelpers.Group
	var filtered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		var sorters = query.GetOrderElastic(map[string]string{
			"2": "members",
			"3": "trend",
		})

		var musts []elastic.Query
		var filters []elastic.Query
		var shoulds []elastic.Query

		var search = query.GetSearchString("search")
		if search != "" {

			musts = append(musts, elastic.NewBoolQuery().MinimumNumberShouldMatch(1).Should(
				elastic.NewMatchQuery("name", search).Fuzziness("1").Boost(3),
				elastic.NewMatchQuery("abbreviation", search).Fuzziness("1").Boost(2),
				elastic.NewMatchQuery("url", search).Fuzziness("1").Boost(2),
			))

			shoulds = append(shoulds,
				// Boost if more group members
				elastic.NewFunctionScoreQuery().AddScoreFunc(elastic.NewFieldValueFactorFunction().Modifier("sqrt").Field("members").Factor(0.003)),
				// Boost if exact match
				elastic.NewMatchPhraseQuery("name", search).Boost(5),
			)
		}

		var hideerrored = query.GetSearchString("filter")
		if hideerrored == "1" {
			filters = append(filters, elastic.NewTermQuery("error", false))
		} else if hideerrored == "0" {
			filters = append(filters, elastic.NewTermQuery("error", true))
		}

		var elasticQuery = elastic.NewBoolQuery().Must(musts...).Filter(filters...).Should(shoulds...)

		groups, filtered, err = elasticHelpers.SearchGroups(100, query.GetOffset(), elasticQuery, sorters)
		if err != nil {
			log.Err(err, r)
			return
		}
	}()

	// Get total
	var total int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionGroups, bson.D{{Key: "type", Value: helpers.GroupTypeGroup}}, 60*60*6)
		log.Err(err, r)
	}()

	wg.Wait()

	var removeWhiteSpace = regexp.MustCompile(`[\s\p{Braille}]{2,}`)

	var response = datatable.NewDataTablesResponse(r, query, total, filtered)
	for k, group := range groups {

		var path = helpers.GetGroupPath(group.ID, group.Name)
		var link = helpers.GetGroupLink(helpers.GroupTypeGroup, group.URL)
		var headline = removeWhiteSpace.ReplaceAllString(group.Headline, " ")
		headline = helpers.TruncateString(headline, 100, "…")

		response.AddRow([]interface{}{
			group.ID,                         // 0
			group.Name,                       // 1
			link,                             // 2
			group.Abbreviation,               // 3
			headline,                         // 4
			helpers.GetGroupIcon(group.Icon), // 5
			group.Members,                    // 6
			helpers.TrendValue(group.Trend),  // 7
			group.Error,                      // 8
			query.GetOffset() + k + 1,        // 9
			path,                             // 10
			group.Score,                      // 11
		})
	}

	returnJSON(w, r, response)
}
