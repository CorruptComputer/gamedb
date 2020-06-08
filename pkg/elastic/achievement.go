package elastic

import (
	"encoding/json"
	"strconv"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/olivere/elastic/v7"
)

type Achievement struct {
	ID          string  `json:"id"` // Achievement key
	Name        string  `json:"name"`
	Icon        string  `json:"icon"`
	Description string  `json:"description"`
	Hidden      bool    `json:"hidden"`
	Completed   float64 `json:"completed"`
	AppID       int     `json:"app_id"`
	AppName     string  `json:"app_name"`
	Score       float64 `json:"-"` // Not stored, just used on frontend
}

func (achievement Achievement) GetKey() string {
	return strconv.Itoa(achievement.AppID) + "-" + achievement.ID
}

func IndexAchievement(achievement Achievement) error {
	return indexDocument(IndexAchievements, achievement.GetKey(), achievement)
}

func IndexAchievementBulk(achievements map[string]Achievement) error {

	i := map[string]interface{}{}
	for k, v := range achievements {
		i[k] = v
	}

	return indexDocuments(IndexAchievements, i)
}

func SearchAppAchievements(offset int, search string, sorters []elastic.Sorter) (achievements []Achievement, total int64, err error) {

	client, ctx, err := GetElastic()
	if err != nil {
		return achievements, 0, err
	}

	searchService := client.Search().
		Index(IndexAchievements).
		From(offset).
		Size(100).
		TrackTotalHits(true).
		Highlight(elastic.NewHighlight().Field("name").Field("description").Field("app_name").PreTags("<mark>").PostTags("</mark>")).
		SortBy(sorters...)

	if search != "" {
		searchService.Query(elastic.NewBoolQuery().
			Must(
				elastic.NewBoolQuery().MinimumNumberShouldMatch(1).Should(
					elastic.NewMatchQuery("name", search).Fuzziness("1").Boost(3),
					elastic.NewMatchQuery("description", search).Fuzziness("1").Boost(2),
					elastic.NewMatchQuery("app_name", search).Fuzziness("1").Boost(1),
				),
			).
			Should(
				elastic.NewMatchPhraseQuery("name", search).Boost(1), // Exact match
			),
		)
	}

	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return achievements, 0, err
	}

	for _, hit := range searchResult.Hits.Hits {

		var achievement Achievement
		err := json.Unmarshal(hit.Source, &achievement)
		if err != nil {
			log.Err(err)
			continue
		}

		if hit.Score != nil {
			achievement.Score = *hit.Score
		}

		if val, ok := hit.Highlight["name"]; ok {
			if len(val) > 0 {
				achievement.Name = val[0]
			}
		}

		if val, ok := hit.Highlight["description"]; ok {
			if len(val) > 0 {
				achievement.Description = val[0]
			}
		}

		if val, ok := hit.Highlight["app_name"]; ok {
			if len(val) > 0 {
				achievement.AppName = val[0]
			}
		}

		achievements = append(achievements, achievement)
	}

	return achievements, searchResult.TotalHits(), err
}

//noinspection GoUnusedExportedFunction
func DeleteAndRebuildAchievementsIndex() {

	var mapping = map[string]interface{}{
		"settings": settings,
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id":          fieldTypeKeyword,
				"name":        fieldTypeText,
				"icon":        fieldTypeDisabled,
				"description": fieldTypeText,
				"hidden":      fieldTypeBool,
				"completed":   fieldTypeHalfFloat,
				"app_id":      fieldTypeInteger,
				"app_name":    fieldTypeText,
			},
		},
	}

	rebuildIndex(IndexAchievements, mapping)
}
