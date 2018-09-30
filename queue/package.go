package queue

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/steam-authority/steam-authority/db"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/streadway/amqp"
)

type RabbitMessagePackage struct {
	RabbitMessageProduct
}

func (d RabbitMessagePackage) getQueueName() string {
	return QueuePackagesData
}

func (d RabbitMessagePackage) getRetryData() RabbitMessageDelay {
	return RabbitMessageDelay{}
}

func (d RabbitMessagePackage) process(msg amqp.Delivery) (ack bool, requeue bool, err error) {

	// Get message
	message := new(RabbitMessagePackage)

	err = json.Unmarshal(msg.Body, message)
	if err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(msg.Body))
		}

		return false, false, err
	}

	// Create mysql row data
	pack := new(db.Package)

	pack.ID = message.ID
	pack.PICSChangeID = message.ChangeNumber
	pack.PICSName = message.KeyValues.Name
	pack.PICSRaw = string(msg.Body)

	for _, v := range message.KeyValues.Children {

		var err error
		var i int
		var i64 int64

		if v.Value != nil {

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
			default:
				logger.Info(v.Name + " field in PICS ignored (Change " + strconv.Itoa(pack.PICSChangeID) + ")")
			}

		} else {

			switch v.Name {
			case "appids":

				var appIDs []int
				for _, vv := range v.Children {
					i, err = strconv.Atoi(vv.Value.(string))
					appIDs = append(appIDs, i)
				}
				pack.SetAppIDs(appIDs)

			case "depotids":

				var depotIDs []int
				for _, vv := range v.Children {
					i, err = strconv.Atoi(vv.Value.(string))
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
				logger.Info(v.Name + " field in PICS ignored (Change " + strconv.Itoa(pack.PICSChangeID) + ")")
			}

		}

		if err != nil {
			logger.Error(err)
		}
	}

	// Update package
	gorm, err := db.GetMySQLClient()
	if err != nil {
		return false, true, err
	}

	gorm.Attrs(db.GetDefaultPackageJSON()).FirstOrCreate(pack, db.Package{ID: pack.ID})
	if gorm.Error != nil {
		logger.Error(gorm.Error)
	}

	return false, true, err

	//if message.PICSChangeID != 0 {
	//	pack.PICSChangeID = message.PICSChangeID
	//}

	priceBeforeFill := pack.PriceFinal

	errs := pack.Update()
	if len(errs) > 0 {
		// Nack on hard fails
		for _, err = range errs {
			if err, ok := err.(db.UpdateError); ok {
				if err.IsHard() {
					return false, false, err
				}
			}
		}
		// Retry on all other errors
		for _, err = range errs {
			logger.Error(err)
			return false, true, err
		}
	}
	//if v.Error() == steam.ErrInvalidJson || v == steam.ErrNullResponse || strings.HasSuffix(v.Error(), "connect: connection refused") {
	//	return false, true
	//}

	gorm.Save(pack)
	if gorm.Error != nil {
		logger.Error(gorm.Error)
	}

	// Save price change
	price := new(db.PriceChange)
	price.Change = pack.PriceFinal - priceBeforeFill

	if price.Change != 0 {

		price.CreatedAt = time.Now()
		price.PackageID = pack.ID
		price.Name = pack.GetName()
		price.PriceInitial = pack.PriceInitial
		price.PriceFinal = pack.PriceFinal
		price.Discount = pack.PriceDiscount
		price.Currency = "usd"
		price.Icon = pack.GetDefaultAvatar()
		price.ReleaseDateNice = pack.GetReleaseDateNice()
		price.ReleaseDateUnix = pack.GetReleaseDateUnix()

		prices, err := db.GetPackagePrices(pack.ID, 1)
		if err != nil {
			logger.Error(err)
		}

		if len(prices) == 0 {
			price.First = true
		}

		_, err = db.SaveKind(price.GetKey(), price)
		if err != nil {
			logger.Error(err)
		}
	}

	return true, false, nil
}
