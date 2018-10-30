package queue

import (
	"strconv"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/logging"
	"github.com/streadway/amqp"
)

type RabbitMessagePackage struct {
	PICSPackageInfo RabbitMessageProduct
}

func (d RabbitMessagePackage) getQueueName() string {
	return QueuePackagesData
}

func (d RabbitMessagePackage) getRetryData() RabbitMessageDelay {
	return RabbitMessageDelay{}
}

func (d RabbitMessagePackage) process(msg amqp.Delivery) (ack bool, requeue bool, err error) {

	// Get message
	rabbitMessage := new(RabbitMessagePackage)

	err = helpers.Unmarshal(msg.Body, rabbitMessage)
	if err != nil {
		return false, false, err
	}

	message := rabbitMessage.PICSPackageInfo

	logging.Info("Consuming package: " + strconv.Itoa(message.ID))

	// Load current package
	gorm, err := db.GetMySQLClient()
	if err != nil {
		return false, true, err
	}

	pack := new(db.Package)
	gorm.First(&pack, message.ID)
	if gorm.Error != nil && !gorm.RecordNotFound() {
		return false, true, gorm.Error
	}

	var packageBeforeUpdate = pack

	// Update with new details
	pack.ID = message.ID
	pack.PICSChangeID = message.ChangeNumber
	pack.PICSName = message.KeyValues.Name
	pack.PICSRaw = string(msg.Body)

	var i int
	var i64 int64

	for _, v := range message.KeyValues.Children {

		switch v.Name {
		case "billingtype":
			i64, err = strconv.ParseInt(v.Value.(string), 10, 8)
			pack.PICSBillingType = int8(i64)
		case "licensetype":
			i64, err = strconv.ParseInt(v.Value.(string), 10, 8)
			pack.PICSLicenseType = int8(i64)
		case "status":
			i64, err = strconv.ParseInt(v.Value.(string), 10, 8)
			pack.PICSStatus = int8(i64)
		case "packageid":
			// Empty
		case "appids":

			var appIDs []int
			for _, vv := range v.Children {
				i, err = strconv.Atoi(vv.Value.(string))
				logging.Error(err)
				appIDs = append(appIDs, i)
			}
			pack.SetAppIDs(appIDs)

		case "depotids":

			var depotIDs []int
			for _, vv := range v.Children {
				i, err = strconv.Atoi(vv.Value.(string))
				logging.Error(err)
				depotIDs = append(depotIDs, i)
			}
			pack.SetDepotIDs(depotIDs)

		case "appitems":

			var appItems = map[string]string{}
			for _, vv := range v.Children {
				if len(vv.Children) == 1 {
					appItems[vv.Name] = vv.Children[0].Value.(string)
				}
			}
			pack.SetAppItems(appItems)

		case "extended":

			var extended = db.Extended{}
			for _, vv := range v.Children {
				extended[vv.Name] = vv.Value.(string)
			}
			pack.SetExtended(extended)

		default:
			logging.Info(v.Name + " field in PICS ignored (Change " + strconv.Itoa(pack.PICSChangeID) + ")")
		}

		logging.Error(err)
	}

	// Update from API
	err = pack.Update()
	if err != nil && err != steam.ErrPackageNotFound {
		return false, true, err
	}

	// Save new data
	gorm.Save(&pack)
	if gorm.Error != nil {
		return false, true, gorm.Error
	}

	// Save price changes
	var prices db.ProductPrices
	var price db.ProductPriceCache
	var kinds []db.Kind
	for code := range steam.Countries {

		var oldPrice, newPrice int

		prices, err = packageBeforeUpdate.GetPrices()
		if err == nil {
			price, err = prices.Get(code)
			if err == nil {
				oldPrice = price.Final
			} else {
				continue // Only compare if there is an old price to compare to
			}
		}

		prices, err = pack.GetPrices()
		if err == nil {
			price, err = prices.Get(code)
			if err == nil {
				newPrice = price.Final
			}
		}

		if oldPrice != newPrice {
			kinds = append(kinds, db.CreateProductPrice(pack, code, oldPrice, newPrice))
		}
	}

	err = db.BulkSaveKinds(kinds, db.KindProductPrice, true)
	if err != nil {
		return false, true, err
	}

	return true, false, nil
}
