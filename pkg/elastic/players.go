package elastic

import (
	"encoding/json"
	"strconv"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/olivere/elastic/v7"
)

type Player struct {
	ID                int64    `json:"id"`
	PersonaName       string   `json:"name"`
	PersonaNameRecent []string `json:"recent"`
	VanityURL         string   `json:"url"`
}

func IndexPlayer(player Player) error {
	return indexDocument(IndexPlayers, strconv.FormatInt(player.ID, 10), player)
}

func SearchPlayers(limit int, offset int, search string, sorters []elastic.Sorter) (players []Player, total int64, err error) {

	client, ctx, err := GetElastic()
	if err != nil {
		return players, 0, err
	}

	searchService := client.Search().
		Index(IndexPlayers).
		From(offset).
		Size(limit).
		TrackTotalHits(true)

	if search != "" {

		var filters []elastic.Query
		var musts []elastic.Query

		musts = append(musts, elastic.NewMatchQuery("name", search))

		// https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl-function-score-query.html#function-field-value-factor
		musts = append(musts, elastic.NewFunctionScoreQuery().AddScoreFunc(
			elastic.NewFieldValueFactorFunction().Field("players").Modifier("log1p")))

		searchService.Query(elastic.NewBoolQuery().Must(musts...).Filter(filters...))
	}

	if sorters != nil && len(sorters) > 0 {
		searchService.SortBy(sorters...)
	}

	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return players, 0, err
	}

	for _, hit := range searchResult.Hits.Hits {

		var player Player
		err := json.Unmarshal(hit.Source, &player)
		if err != nil {
			log.Err(err)
		}

		players = append(players, player)
	}

	return players, searchResult.TotalHits(), err
}

//noinspection GoUnusedExportedFunction
func DeleteAndRebuildPlayersIndex() {

	var mapping = map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "long",
				},
				"name": map[string]interface{}{
					"type": "text",
				},
				"recent": map[string]interface{}{
					"type": "text",
				},
				"url": map[string]interface{}{
					"type": "text",
				},
			},
		},
	}

	err := rebuildIndex(IndexPlayers, mapping)
	log.Err(err)
}
