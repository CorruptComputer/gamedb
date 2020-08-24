package mongo

import (
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type BundlePrice struct {
	CreatedAt time.Time `bson:"created_at"`
	BundleID  int       `bson:"bundle_id"`
	Discount  int       `bson:"price"`
}

func (price BundlePrice) BSON() bson.D {

	return bson.D{
		{"_id", price.GetKey()},
		{"created_at", price.CreatedAt},
		{"bundle_id", price.BundleID},
		{"price", price.Discount},
	}
}

func (price BundlePrice) GetKey() string {
	return strconv.Itoa(price.BundleID) + "-" + price.CreatedAt.Format(time.RFC3339)
}

func GetBundlePrices(bundleID int) (prices []BundlePrice, err error) {

	var sort = bson.D{{"created_at", 1}}
	var filter = bson.D{{"bundle_id", bundleID}}

	cur, ctx, err := Find(CollectionBundlePrices, 0, 0, sort, filter, nil, nil)
	if err != nil {
		return prices, err
	}

	defer func(cur *mongo.Cursor) {
		err = cur.Close(ctx)
		if err != nil {
			log.ErrS(err)
		}
	}(cur)

	for cur.Next(ctx) {

		var price BundlePrice
		err := cur.Decode(&price)
		if err != nil {
			log.ErrS(err, price.GetKey())
		} else {
			prices = append(prices, price)
		}
	}

	return prices, cur.Err()
}
