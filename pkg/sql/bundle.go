package sql

import (
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
)

type Bundle struct {
	ID              int       `gorm:"not null;column:id"`
	CreatedAt       time.Time `gorm:"not null;column:created_at;type:datetime"`
	UpdatedAt       time.Time `gorm:"not null;column:updated_at;type:datetime"`
	Name            string    `gorm:"not null;column:name"`
	Discount        int       `gorm:"not null;column:discount"`
	HighestDiscount int       `gorm:"not null;column:highest_discount"`
	LowestDiscount  int       `gorm:"not null;column:lowest_discount"`
	AppIDs          string    `gorm:"not null;column:app_ids"`
	PackageIDs      string    `gorm:"not null;column:package_ids"`
	Image           string    `gorm:"not null;column:image"`
}

func (bundle *Bundle) BeforeSave(scope *gorm.Scope) error {

	if bundle.AppIDs == "" {
		bundle.AppIDs = "[]"
	}
	if bundle.PackageIDs == "" {
		bundle.PackageIDs = "[]"
	}

	return nil
}

func (bundle *Bundle) SetDiscount(discount int) {

	bundle.Discount = discount

	if discount < bundle.HighestDiscount {
		bundle.HighestDiscount = discount
	}

	if discount > bundle.LowestDiscount || bundle.LowestDiscount == 0 {
		bundle.LowestDiscount = discount
	}
}

func (bundle Bundle) GetPath() string {
	return "/bundles/" + strconv.Itoa(bundle.ID) + "/" + slug.Make(bundle.GetName())
}

func (bundle Bundle) GetName() string {
	if bundle.Name != "" {
		return bundle.Name
	}
	return "Bundle " + strconv.Itoa(bundle.ID)
}

func (bundle Bundle) GetStoreLink() string {
	name := config.Config.GameDBShortName.Get()
	return "https://store.steampowered.com/bundle/" + strconv.Itoa(bundle.ID) + "?utm_source=" + name + "&utm_medium=link&utm_campaign=" + name
}

func (bundle Bundle) GetUpdatedNice() string {
	return bundle.UpdatedAt.Format(helpers.DateYearTime)
}

func (bundle Bundle) GetAppIDs() (ids []int, err error) {

	err = helpers.Unmarshal([]byte(bundle.AppIDs), &ids)
	return ids, err
}

func (bundle Bundle) AppsCount() int {

	apps, err := bundle.GetAppIDs()
	log.Err(err)
	return len(apps)
}

func (bundle Bundle) GetPackageIDs() (ids []int, err error) {

	err = helpers.Unmarshal([]byte(bundle.PackageIDs), &ids)
	return ids, err
}

func (bundle Bundle) PackagesCount() int {

	packages, err := bundle.GetPackageIDs()
	log.Err(err)
	return len(packages)
}

func (bundle Bundle) OutputForJSON() (output []interface{}) {

	return []interface{}{
		bundle.ID,        // 0
		bundle.GetName(), // 1
		bundle.GetPath(), // 2
		strconv.FormatInt(bundle.UpdatedAt.Unix(), 10), // 3
		bundle.Discount,                           // 4
		bundle.AppsCount(),                        // 5
		bundle.PackagesCount(),                    // 6
		bundle.HighestDiscount == bundle.Discount, // 7 Is best discount
	}
}

func (bundle Bundle) Save() error {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	db = db.Save(&bundle)
	return db.Error
}

func GetBundle(id int, columns []string) (bundle Bundle, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return bundle, err
	}

	db = db.First(&bundle, id)
	if db.Error != nil {
		return bundle, db.Error
	}

	if columns != nil && len(columns) > 0 {
		db = db.Select(columns)
		if db.Error != nil {
			return bundle, db.Error
		}
	}

	if bundle.ID == 0 {
		return bundle, ErrRecordNotFound
	}

	return bundle, nil
}

func CountBundles() (count int, err error) {

	var item = helpers.MemcacheBundlesCount

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		var count int

		db, err := GetMySQLClient()
		if err != nil {
			return count, err
		}

		db.Model(&Bundle{}).Count(&count)

		return count, db.Error
	})

	return count, err
}
