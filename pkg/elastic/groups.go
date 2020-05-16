package elastic

import (
	"encoding/json"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/olivere/elastic/v7"
)

type Group struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	URL          string `json:"url"`
	Abbreviation string `json:"abbreviation"`
	Headline     string `json:"headline"`
	Icon         string `json:"icon"`
	Members      int    `json:"members"`
	Trend        int64  `json:"trend"`
	Error        bool   `json:"error"`
}

func IndexGroup(group Group) error {
	return indexDocument(IndexGroups, group.ID, group)
}

func SearchGroups(limit int, offset int, search string, sorters []elastic.Sorter) (groups []Group, total int64, err error) {

	client, ctx, err := GetElastic()
	if err != nil {
		return groups, 0, err
	}

	searchService := client.Search().
		Index(IndexGroups).
		From(offset).
		Size(limit).
		TrackTotalHits(true)

	if search != "" {
		searchService.Query(elastic.NewBoolQuery().Must(
			elastic.NewMultiMatchQuery(search, "name^2", "abbreviation^2", "url^1.5", "headline^1").Type("best_fields"),
			elastic.NewFunctionScoreQuery().AddScoreFunc(elastic.NewFieldValueFactorFunction().Field("members").Factor(0.001)),
		))
	}

	if sorters != nil && len(sorters) > 0 {
		searchService.SortBy(sorters...)
	}

	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return groups, 0, err
	}

	for _, hit := range searchResult.Hits.Hits {

		var group Group
		err := json.Unmarshal(hit.Source, &group)
		if err != nil {
			log.Err(err)
		}

		groups = append(groups, group)
	}

	return groups, searchResult.TotalHits(), err
}

//noinspection GoUnusedExportedFunction
func DeleteAndRebuildGroupsIndex() {

	var mapping = map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "keyword",
				},
				"name": map[string]interface{}{
					"type": "text",
				},
				"url": map[string]interface{}{
					"type": "text",
				},
				"abbreviation": map[string]interface{}{
					"type": "text",
				},
				"headline": map[string]interface{}{
					"type": "text",
				},
				"icon": map[string]interface{}{
					"enabled": false,
				},
				"members": map[string]interface{}{
					"type": "integer",
				},
				"trend": map[string]interface{}{
					"type": "integer",
				},
				"error": map[string]interface{}{
					"type": "boolean",
				},
			},
		},
	}

	err := rebuildIndex(IndexGroups, mapping)
	log.Err(err)
}
