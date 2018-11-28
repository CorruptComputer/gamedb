package queue

import (
	"errors"
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/websockets"
	"github.com/streadway/amqp"
)

type RabbitMessagePackage struct {
	PICSPackageInfo RabbitMessageProduct
}

func (d RabbitMessagePackage) getConsumeQueue() RabbitQueue {
	return QueuePackagesData
}

func (d RabbitMessagePackage) getProduceQueue() RabbitQueue {
	return QueuePackages
}

func (d RabbitMessagePackage) getRetryData() RabbitMessageDelay {
	return RabbitMessageDelay{}
}

func (d RabbitMessagePackage) process(msg amqp.Delivery) (requeue bool, err error) {

	// Get message
	rabbitMessage := new(RabbitMessagePackage)

	err = helpers.Unmarshal(msg.Body, rabbitMessage)
	if err != nil {
		return false, err
	}

	message := rabbitMessage.PICSPackageInfo

	log.Log(log.SeverityInfo, "Consuming package: " + strconv.Itoa(message.ID))

	if !db.IsValidPackageID(message.ID) {
		return false, errors.New("invalid package ID: " + strconv.Itoa(message.ID))
	}

	// Load current package
	gorm, err := db.GetMySQLClient()
	if err != nil {
		return true, err
	}

	pack := db.Package{}
	gorm.First(&pack, message.ID)
	if gorm.Error != nil && !gorm.RecordNotFound() {
		return true, gorm.Error
	}

	if pack.PICSChangeNumber >= message.ChangeNumber {
		return false, nil
	}

	var packageBeforeUpdate = pack

	// Update with new details
	pack.ID = message.ID

	if message.ChangeNumber > pack.PICSChangeNumber {
		pack.PICSChangeNumberDate = time.Now()
	}

	pack.PICSChangeNumber = message.ChangeNumber
	pack.PICSName = message.KeyValues.Name

	for _, v := range message.KeyValues.Children {

		switch v.Name {
		case "billingtype":
			var i64 int64
			i64, err = strconv.ParseInt(v.Value.(string), 10, 8)
			pack.PICSBillingType = int8(i64)
		case "licensetype":
			var i64 int64
			i64, err = strconv.ParseInt(v.Value.(string), 10, 8)
			pack.PICSLicenseType = int8(i64)
		case "status":
			var i64 int64
			i64, err = strconv.ParseInt(v.Value.(string), 10, 8)
			pack.PICSStatus = int8(i64)
		case "packageid":
			// Empty
		case "appids":

			err = pack.SetAppIDs(helpers.StringSliceToIntSlice(v.GetChildrenAsSlice()))
			log.Log(err)

		case "depotids":

			err = pack.SetDepotIDs(helpers.StringSliceToIntSlice(v.GetChildrenAsSlice()))
			log.Log(err)

		case "appitems":

			var appItems = map[string]string{}
			for _, vv := range v.Children {
				if len(vv.Children) == 1 {
					appItems[vv.Name] = vv.Children[0].Value.(string)
				}
			}
			err = pack.SetAppItems(appItems)
			log.Log(err)

		case "extended":

			err = pack.SetExtended(v.GetExtended())
			log.Log(err)

		default:
			log.Log(log.SeverityInfo, v.Name + " field in PICS ignored (Change " + strconv.Itoa(pack.PICSChangeNumber) + ")")
		}

		log.Log(err)
	}

	// Update from API
	err = pack.Update()
	if err != nil && err != steam.ErrPackageNotFound {
		return true, err
	}

	// Save new data
	gorm = gorm.Save(&pack)
	if gorm.Error != nil {
		return true, gorm.Error
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
			} else {
				continue // Only compare if there is a new price to compare to
			}
		}

		if oldPrice != newPrice {
			kinds = append(kinds, db.CreateProductPrice(pack, code, oldPrice, newPrice))
		}
	}

	err = db.BulkSaveKinds(kinds, db.KindProductPrice, true)
	if err != nil {
		return true, err
	}

	// Send websocket
	page, err := websockets.GetPage(websockets.PagePackages)
	if err == nil && page.HasConnections() {

		page.Send(pack.OutputForJSON(steam.CountryUS))
	}

	return false, err
}
