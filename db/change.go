package db

import (
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gamedb/website/helpers"
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

func (change Change) GetName() (name string) {

	return "Change " + strconv.Itoa(change.ChangeID)
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

	err = helpers.GetMemcache().GetSet(item.Key, item.Expiration, &change, func() (interface{}, error) {

		var change Change

		client, context, err := GetDSClient()
		if err != nil {
			return change, err
		}

		err = client.Get(context, datastore.NameKey(KindChange, strconv.FormatInt(id, 10), nil), &change)
		if err != nil {
			if err2, ok := err.(*datastore.ErrFieldMismatch); ok {

				removedColumns := []string{
					"updated_at",
					"apps",
					"packages",
				}

				if helpers.SliceHasString(removedColumns, err2.FieldName) {
					err = nil
				}
			}
		}

		return change, err
	})

	return change, err
}
