package db

import (
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gamedb/website/helpers"
	"go.mongodb.org/mongo-driver/bson"
)

type Change struct {
	CreatedAt time.Time    `datastore:"created_at,noindex"`
	ChangeID  int          `datastore:"change_id"`
	Apps      []ChangeItem `datastore:"apps,noindex"`
	Packages  []ChangeItem `datastore:"packages,noindex"`
}

type ChangeItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (change Change) GetKey() (key *datastore.Key) {
	return datastore.NameKey(KindChange, strconv.Itoa(change.ChangeID), nil)
}

func (change Change) GetMongoKey() interface{} {
	return change.ChangeID
}

func (change Change) GetName() (name string) {

	return "Change " + strconv.Itoa(change.ChangeID)
}

func (change Change) ToBSON() (ret interface{}) {

	m := bson.M{
		"_id":        change.ChangeID,
		"created_at": change.CreatedAt,
	}

	// Apps
	if len(change.Apps) > 0 {
		var apps bson.A
		for _, v := range change.Apps {
			apps = append(apps, bson.M{
				"id":   v.ID,
				"name": v.Name,
			})
		}
		m["apps"] = apps
	}

	// Packages
	if len(change.Packages) > 0 {
		var packages bson.A
		for _, v := range change.Packages {
			packages = append(packages, bson.M{
				"id":   v.ID,
				"name": v.Name,
			})
		}
		m["packages"] = packages
	}

	return m
}

func (change Change) GetTimestamp() int64 {
	return change.CreatedAt.Unix()
}

func (change Change) GetNiceDate() string {
	return change.CreatedAt.Format(helpers.DateYearTime)
}

func (change Change) GetPath() string {
	return "/changes/" + strconv.Itoa(change.ChangeID)
}

func (change Change) GetAppIDs() (ids []int) {
	for _, v := range change.Apps {
		ids = append(ids, v.ID)
	}
	return ids
}

func (change Change) GetPackageIDs() (ids []int) {
	for _, v := range change.Packages {
		ids = append(ids, v.ID)
	}
	return ids
}

func (change Change) OutputForJSON() (output []interface{}) {

	return []interface{}{
		change.ChangeID,
		change.CreatedAt.Unix(),
		change.CreatedAt.Format(helpers.DateYearTime),
		change.Apps,
		change.Packages,
		change.GetPath(),
	}
}

func GetChange(id int64) (change Change, err error) {

	var item = helpers.MemcacheChangeRow(id)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &change, func() (interface{}, error) {

		var change Change

		// Try MySQL
		db, err := GetMySQLClient()
		if err != nil {
			return change, err
		}

		var buffer DatastoreBuffer
		db.Where("kind = ?", KindChange).Where("key_name = ?", id).First(&buffer)
		if db.Error != nil {
			return change, db.Error
		}

		if buffer.Kind != "" {
			return buffer.ToChange()
		}

		// Try Datastore
		client, context, err := GetDSClient()
		if err != nil {
			return change, err
		}

		err = client.Get(context, datastore.NameKey(KindChange, strconv.FormatInt(id, 10), nil), &change)
		err = handleDSSingleError(err, OldChangeFields)

		return change, err
	})

	return change, err
}
