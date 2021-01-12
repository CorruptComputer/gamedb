package main

import (
	"context"
	"regexp"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/cmd/backend/helpers"
	"github.com/gamedb/gamedb/pkg/backend/generated"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type StatsServer struct {
	generated.UnimplementedStatsServiceServer
}

func (s StatsServer) List(ctx context.Context, request *generated.StatsRequest) (response *generated.StatsResponse, err error) {

	offset := request.GetPagination().GetOffset()
	limit := request.GetPagination().GetLimit()

	filter := bson.D{{"type", request.GetType()}}
	filter2 := filter

	if len(request.GetSearch()) > 0 {
		quoted := regexp.QuoteMeta(request.GetSearch())
		filter2 = append(filter2, bson.E{Key: "$or", Value: bson.A{
			bson.M{"name": bson.M{"$regex": quoted, "$options": "i"}},
		}})
	}

	stats, err := mongo.GetStats(offset, limit, filter2, helpers.MakeMongoOrder(request.GetPagination()))
	if err != nil {
		return nil, err
	}

	total, err := mongo.CountDocuments(mongo.CollectionStats, filter, 0)
	if err != nil {
		return nil, err
	}

	filtered, err := mongo.CountDocuments(mongo.CollectionStats, filter2, 0)
	if err != nil {
		return nil, err
	}

	response = &generated.StatsResponse{}
	response.Pagination = helpers.MakePaginationResponse(request.GetPagination(), total, filtered)

	for _, stat := range stats {

		s := &generated.StatResponse{
			Id:            int32(stat.ID),
			Name:          stat.Name,
			Apps:          int32(stat.Apps),
			AppsPercent:   stat.AppsPercnt,
			MeanScore:     stat.MeanScore,
			MeanPlayers:   float32(stat.MeanPlayers),
			MedianScore:   stat.MedianScore,
			MedianPlayers: int32(stat.MedianPlayers),
			// MeanPrice:     stat.MeanPrice,
			// MedianPrice:   stat.MedianPrice,
		}

		if val, ok := stat.MeanPrice[steamapi.ProductCC(request.GetCurrency())]; ok {
			s.MeanPrice = val
		}

		if val, ok := stat.MedianPrice[steamapi.ProductCC(request.GetCurrency())]; ok {
			s.MedianPrice = int32(val)
		}

		response.Stats = append(response.Stats, s)
	}

	return response, err
}
