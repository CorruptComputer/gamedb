package elastic

import (
	"encoding/json"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/olivere/elastic/v7"
)

type Group struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	NameMarked   string  `json:"name_marked"`
	URL          string  `json:"url"`
	Abbreviation string  `json:"abbreviation"`
	Headline     string  `json:"headline"`
	Icon         string  `json:"icon"`
	Members      int     `json:"members"`
	Trend        int64   `json:"trend"`
	Error        bool    `json:"error"`
	Score        float64 `json:"-"`
}

func (group Group) GetAbbr() string {
	return helpers.GetGroupAbbreviation(group.Abbreviation)
}

func (group Group) GetHeadline() string {
	group.Headline = helpers.RegexFilterEmptyCharacters.ReplaceAllString(group.Headline, " ")
	group.Headline = helpers.TruncateString(group.Headline, 100, "…")
	return group.Headline
}

func (group Group) GetName() string {
	return helpers.GetGroupName(group.ID, group.Name)
}

func (group Group) GetNameMarked() string {
	return helpers.GetGroupName(group.ID, group.NameMarked)
}

func (group Group) GetPath() string {
	return helpers.GetGroupPath(group.ID, group.GetName())
}

func (group Group) GetIcon() string {
	return helpers.GetGroupIcon(group.Icon)
}

func (group Group) GetTrend() string {
	return helpers.TrendValue(group.Trend)
}

func (group Group) GetGroupLink() string {
	return helpers.GetGroupLink(helpers.GroupTypeGroup, group.URL)
}

func (group Group) GetGameLink() string {
	return helpers.GetGroupLink(helpers.GroupTypeGame, group.URL)
}

func IndexGroup(group Group) error {
	return indexDocument(IndexGroups, group.ID, group)
}

func SearchGroups(offset int, limit int, sorters []elastic.Sorter, search string, errors string) (groups []Group, aggregations map[string]map[string]int64, total int64, err error) {

	client, ctx, err := GetElastic()
	if err != nil {
		return groups, aggregations, 0, err
	}

	var query = elastic.NewBoolQuery()
	if search != "" {

		query.Must(elastic.NewBoolQuery().MinimumNumberShouldMatch(1).Should(
			elastic.NewTermQuery("id", search).Boost(5),
			elastic.NewMatchQuery("name", search).Boost(1),
			// elastic.NewMatchQuery("headline", search),
			// elastic.NewMatchQuery("abbreviation", search).Boost(2),
			// elastic.NewMatchQuery("url", search).Boost(1),
		))

		query.Should(
			elastic.NewFunctionScoreQuery().
				AddScoreFunc(elastic.NewFieldValueFactorFunction().Modifier("sqrt").Field("members").Factor(0.001)),
		)
	}

	if errors == "0" || errors == "1" {
		query.Filter(elastic.NewTermQuery("error", errors == "0"))
	}

	searchService := client.Search().
		Index(IndexGroups).
		From(offset).
		Size(limit).
		TrackTotalHits(true).
		Highlight(elastic.NewHighlight().Field("name").PreTags("<mark>").PostTags("</mark>")).
		SortBy(sorters...).
		Query(query).
		SearchType("dfs_query_then_fetch"). // Improves acuracy with multiple shards
		Aggregation("error", elastic.NewTermsAggregation().Field("error").Size(10).OrderByCountDesc())

	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return groups, aggregations, 0, err
	}

	aggregations = make(map[string]map[string]int64, len(searchResult.Aggregations))
	for k := range searchResult.Aggregations {
		a, ok := searchResult.Aggregations.Terms(k)
		if ok {
			aggregations[k] = make(map[string]int64, len(a.Buckets))
			for _, v := range a.Buckets {
				aggregations[k][*v.KeyAsString] = v.DocCount
			}
		}
	}

	for _, hit := range searchResult.Hits.Hits {

		var group Group
		err := json.Unmarshal(hit.Source, &group)
		if err != nil {
			log.Err(err)
			continue
		}

		if hit.Score != nil {
			group.Score = *hit.Score
		}

		group.NameMarked = group.Name
		if val, ok := hit.Highlight["name"]; ok {
			if len(val) > 0 {
				group.NameMarked = val[0]
			}
		}

		groups = append(groups, group)
	}

	return groups, aggregations, searchResult.TotalHits(), nil
}

//noinspection GoUnusedExportedFunction
func DeleteAndRebuildGroupsIndex() {

	var mapping = map[string]interface{}{
		"settings": settings,
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id":           fieldTypeKeyword,
				"name":         fieldTypeText,
				"url":          fieldTypeText,
				"abbreviation": fieldTypeText,
				"headline":     fieldTypeText,
				"icon":         fieldTypeDisabled,
				"members":      fieldTypeInteger,
				"trend":        fieldTypeInteger,
				"error":        fieldTypeBool,
			},
		},
	}

	rebuildIndex(IndexGroups, mapping)
}
