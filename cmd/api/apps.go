package main

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/pkg/backend"
	generatedBackend "github.com/gamedb/gamedb/pkg/backend/generated"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.uber.org/zap"
)

func (s Server) GetGames(w http.ResponseWriter, r *http.Request, params generated.GetGamesParams) {

	var limit int64 = 10
	if params.Limit != nil && *params.Limit >= 1 && *params.Limit <= 1000 {
		limit = int64(*params.Limit)
	}

	var offset int64 = 0
	if params.Offset != nil {
		offset = int64(*params.Offset)
	}

	payload := &generatedBackend.ListAppsRequest{
		Pagination: &generatedBackend.PaginationRequest{
			Offset: offset,
			Limit:  limit,
		},
	}

	if params.Ids != nil {
		payload.Ids = *params.Ids
	}

	if params.Tags != nil {
		payload.Tags = *params.Tags
	}

	if params.Genres != nil {
		payload.Genres = *params.Genres
	}

	if params.Categories != nil {
		payload.Categories = *params.Categories
	}

	if params.Developers != nil {
		payload.Developers = *params.Developers
	}

	if params.Publishers != nil {
		payload.Publishers = *params.Publishers
	}

	if params.Platforms != nil {
		payload.Platforms = *params.Platforms
	}

	conn, ctx, err := backend.GetClient()
	if err != nil {
		log.ErrS(err)
		returnResponse(w, r, http.StatusInternalServerError, generated.GamesResponse{Error: err.Error()})
		return
	}

	resp, err := generatedBackend.NewAppsServiceClient(conn).List(ctx, payload)
	if err != nil {
		log.ErrS(err)
		returnResponse(w, r, http.StatusInternalServerError, generated.GamesResponse{Error: err.Error()})
		return
	}

	// Get stats
	var tagIDs []int
	var genreIDs []int
	var publisherIDs []int
	var developerIDs []int
	var categoryIDs []int

	var mapTagIDs = map[int]mongo.Stat{}
	var mapGenreIDs = map[int]mongo.Stat{}
	var mapPublisherIDs = map[int]mongo.Stat{}
	var mapDeveloperIDs = map[int]mongo.Stat{}
	var mapCategoryIDs = map[int]mongo.Stat{}

	for _, v := range resp.Apps {
		tagIDs = append(tagIDs, helpers.Int32sToInts(v.GetTags())...)
		genreIDs = append(genreIDs, helpers.Int32sToInts(v.GetGenres())...)
		publisherIDs = append(publisherIDs, helpers.Int32sToInts(v.GetPublishers())...)
		developerIDs = append(developerIDs, helpers.Int32sToInts(v.GetDevelopers())...)
		categoryIDs = append(categoryIDs, helpers.Int32sToInts(v.GetCategories())...)
	}

	tags, err := mongo.GetStatsByID(mongo.StatsTypeTags, tagIDs)
	if err != nil {
		log.Err("finding tags", zap.Error(err))
	} else {
		for _, v := range tags {
			mapTagIDs[v.ID] = v
		}
	}

	categories, err := mongo.GetStatsByID(mongo.StatsTypeCategories, categoryIDs)
	if err != nil {
		log.Err("finding categories", zap.Error(err))
	} else {
		for _, v := range categories {
			mapCategoryIDs[v.ID] = v
		}
	}

	developers, err := mongo.GetStatsByID(mongo.StatsTypeDevelopers, developerIDs)
	if err != nil {
		log.Err("finding developers", zap.Error(err))
	} else {
		for _, v := range developers {
			mapDeveloperIDs[v.ID] = v
		}
	}

	genres, err := mongo.GetStatsByID(mongo.StatsTypeGenres, genreIDs)
	if err != nil {
		log.Err("finding genres", zap.Error(err))
	} else {
		for _, v := range genres {
			mapGenreIDs[v.ID] = v
		}
	}

	publishers, err := mongo.GetStatsByID(mongo.StatsTypePublishers, publisherIDs)
	if err != nil {
		log.Err("finding publishers", zap.Error(err))
	} else {
		for _, v := range publishers {
			mapPublisherIDs[v.ID] = v
		}
	}

	result := generated.GamesResponse{}
	result.Pagination.Fill(offset, limit, resp.Pagination.GetTotal())

	for _, app := range resp.Apps {

		newApp := generated.GameSchema{
			Id:              int(app.GetId()),
			Name:            app.GetName(),
			Icon:            app.GetIcon(),
			MetacriticScore: app.GetMetaScore(),
			PlayersMax:      int(app.GetPlayersMax()),
			PlayersWeekMax:  int(app.GetPlayersWeekMax()),
			ReleaseDate:     app.GetReleaseDateUnix().GetSeconds(),
			ReviewsNegative: int(app.GetReviewsNegative()),
			ReviewsPositive: int(app.GetReviewsPositive()),
			ReviewsScore:    float64(app.GetReviewsScore()),
			// PlayersWeekAvg:  float64(app.GetPlayersWeekAvg()),

			// Fix nulls in JSON
			Prices: generated.GameSchema_Prices{
				AdditionalProperties: map[string]generated.ProductPriceSchema{},
			},
			Tags:       []generated.StatSchema{},
			Categories: []generated.StatSchema{},
			Genres:     []generated.StatSchema{},
			Developers: []generated.StatSchema{},
			Publishers: []generated.StatSchema{},
		}

		for k, price := range app.GetPrices() {
			newApp.Prices.AdditionalProperties[k] = generated.ProductPriceSchema{
				Currency:        price.GetCurrency(),
				DiscountPercent: price.GetDiscountPercent(),
				Final:           price.GetFinal(),
				Free:            price.GetFree(),
				Individual:      price.GetIndividual(),
				Initial:         price.GetInitial(),
			}
		}

		for _, v := range app.GetTags() {
			stat := mapTagIDs[int(v)]
			newApp.Tags = append(newApp.Tags, generated.StatSchema{Id: stat.ID, Name: stat.Name})
		}

		for _, v := range app.GetCategories() {
			stat := mapCategoryIDs[int(v)]
			newApp.Categories = append(newApp.Categories, generated.StatSchema{Id: stat.ID, Name: stat.Name})
		}

		for _, v := range app.GetGenres() {
			stat := mapGenreIDs[int(v)]
			newApp.Genres = append(newApp.Genres, generated.StatSchema{Id: stat.ID, Name: stat.Name})
		}

		for _, v := range app.GetDevelopers() {
			stat := mapDeveloperIDs[int(v)]
			newApp.Developers = append(newApp.Developers, generated.StatSchema{Id: stat.ID, Name: stat.Name})
		}

		for _, v := range app.GetPublishers() {
			stat := mapPublisherIDs[int(v)]
			newApp.Publishers = append(newApp.Publishers, generated.StatSchema{Id: stat.ID, Name: stat.Name})
		}

		result.Games = append(result.Games, newApp)
	}

	returnResponse(w, r, http.StatusOK, result)
}
