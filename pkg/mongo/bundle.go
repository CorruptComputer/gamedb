package mongo

import (
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"go.mongodb.org/mongo-driver/bson"
)

const BundleTypeCompleteTheSet = "cts"
const BundleTypePurchaseTogether = "pt"

type Bundle struct {
	Apps            []int                      `bson:"apps"`
	CreatedAt       time.Time                  `bson:"created_at"`
	Discount        int                        `bson:"discount"`
	DiscountHighest int                        `bson:"discount_highest"`
	DiscountLowest  int                        `bson:"discount_lowest"`
	DiscountSale    int                        `bson:"discount_sale"`
	Giftable        bool                       `bson:"giftable"`
	Icon            string                     `bson:"icon"`
	ID              int                        `bson:"_id"`
	Image           string                     `bson:"image"`
	Name            string                     `bson:"name"`
	OnSale          bool                       `bson:"on_sale"`
	Packages        []int                      `bson:"packages"`
	Prices          map[steamapi.ProductCC]int `bson:"prices"`
	PricesSale      map[steamapi.ProductCC]int `bson:"prices_sale"`
	Type            string                     `bson:"type"`
	UpdatedAt       time.Time                  `bson:"updated_at"`
}

func (bundle Bundle) BSON() bson.D {

	// Dates
	bundle.UpdatedAt = time.Now()

	if bundle.CreatedAt.IsZero() || bundle.CreatedAt.Unix() == 0 {
		bundle.CreatedAt = time.Now()
	}

	// Discount always set, discountSale NOT always set
	if bundle.DiscountSale == 0 {
		bundle.DiscountSale = bundle.Discount
	}

	// priceSale is always set, price is NOT always set
	for _, v := range i18n.GetProdCCs(true) {

		if bundle.Prices[v.ProductCode] == 0 {
			bundle.Prices[v.ProductCode] = bundle.PricesSale[v.ProductCode]
		}
	}

	// Set highest and lowest
	if bundle.DiscountSale > bundle.DiscountHighest {
		bundle.DiscountHighest = bundle.DiscountSale
	}

	if bundle.Discount < bundle.DiscountLowest {
		bundle.DiscountLowest = bundle.Discount
	}

	//
	bundle.OnSale = bundle.DiscountSale > bundle.Discount

	return bson.D{
		{"_id", bundle.ID},
		{"apps", bundle.Apps},
		{"created_at", bundle.CreatedAt},
		{"discount", bundle.Discount},
		{"discount_highest", bundle.DiscountHighest},
		{"discount_lowest", bundle.DiscountLowest},
		{"discount_sale", bundle.DiscountSale},
		{"giftable", bundle.Giftable},
		{"icon", bundle.Icon},
		{"image", bundle.Image},
		{"name", bundle.Name},
		{"on_sale", bundle.OnSale},
		{"packages", bundle.Packages},
		{"prices", bundle.Prices},
		{"prices_sale", bundle.PricesSale},
		{"type", bundle.Type},
		{"updated_at", bundle.UpdatedAt},
	}
}

func (bundle Bundle) OutputForJSON() (output []interface{}) {
	return helpers.OutputBundleForJSON(bundle)
}

func (bundle Bundle) GetName() string {
	return bundle.Name
}

func (bundle Bundle) GetPath() string {
	return helpers.GetBundlePath(bundle.ID, bundle.Name)
}

func (bundle Bundle) GetStoreLink() string {
	return helpers.GetBundleStoreLink(bundle.ID)
}

func (bundle Bundle) GetID() int {
	return bundle.ID
}

func (bundle Bundle) GetUpdated() time.Time {
	return bundle.UpdatedAt
}

func (bundle Bundle) GetDiscount() int {
	return bundle.Discount
}

func (bundle Bundle) GetDiscountSale() int {
	return bundle.DiscountSale
}

func (bundle Bundle) GetDiscountHighest() int {
	return bundle.DiscountHighest
}

func (bundle Bundle) GetPrices() map[steamapi.ProductCC]int {
	return bundle.Prices
}

func (bundle Bundle) GetPricesFormatted() map[steamapi.ProductCC]string {
	return helpers.GetBundlePricesFormatted(bundle.Prices)
}

func (bundle Bundle) GetPricesSaleFormatted() map[steamapi.ProductCC]string {
	return helpers.GetBundlePricesFormatted(bundle.PricesSale)
}

func (bundle Bundle) GetScore() float64 {
	return 0
}

func (bundle Bundle) GetType() string {
	return bundle.Type
}

func (bundle Bundle) GetApps() int {
	return len(bundle.Apps)
}

func (bundle Bundle) GetPackages() int {
	return len(bundle.Packages)
}

func (bundle Bundle) IsGiftable() bool {
	return bundle.Giftable
}

func (bundle Bundle) GetUpdatedNice() string {
	return bundle.UpdatedAt.Format(helpers.DateYearTime)
}

func (bundle Bundle) GetCreatedNice() string {
	return bundle.CreatedAt.Format(helpers.DateYearTime)
}

func BatchBundles(filter bson.D, projection bson.M, callback func(bundles []Bundle)) (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		bundles, err := GetBundles(offset, limit, bson.D{{"_id", 1}}, filter, projection)
		if err != nil {
			return err
		}

		callback(bundles)

		if int64(len(bundles)) != limit {
			break
		}

		offset += limit
	}

	return nil
}

func GetBundle(id int) (bundle Bundle, err error) {

	item := memcache.ItemBundle(id)
	err = memcache.Client().GetSet(item.Key, item.Expiration, &bundle, func() (interface{}, error) {

		err = FindOne(CollectionBundles, bson.D{{"_id", id}}, nil, nil, &bundle)
		return bundle, err
	})

	return bundle, err
}

func GetBundlesByID(ids []int, projection bson.M) (bundles []Bundle, err error) {

	if len(ids) < 1 {
		return bundles, nil
	}

	a := bson.A{}
	for _, v := range ids {
		a = append(a, v)
	}

	filter := bson.D{{"_id", bson.M{"$in": a}}}

	return GetBundles(0, 0, nil, filter, projection)
}

func GetBundles(offset int64, limit int64, sort bson.D, filter bson.D, projection bson.M) (bundles []Bundle, err error) {

	cur, ctx, err := find(CollectionBundles, offset, limit, filter, sort, projection, nil)
	if err != nil {
		return bundles, err
	}

	defer closeCursor(cur, ctx)

	for cur.Next(ctx) {

		var bundle Bundle
		err := cur.Decode(&bundle)
		if err != nil {
			log.ErrS(err, bundle.ID)
		} else {
			bundles = append(bundles, bundle)
		}
	}

	return bundles, cur.Err()
}
