package datastore

import (
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
)

type Change struct {
	CreatedAt time.Time `datastore:"created_at,noindex"`
	UpdatedAt time.Time `datastore:"updated_at,noindex"` // Do not use!  (backwards compatibility)
	ChangeID  int       `datastore:"change_id"`
	Apps      []int     `datastore:"apps,noindex"`
	Packages  []int     `datastore:"packages,noindex"`
}

func (change Change) GetKey() (key *datastore.Key) {
	return datastore.NameKey(KindChange, strconv.Itoa(change.ChangeID), nil)
}

func (change Change) GetName() (name string) {

	return "Change " + strconv.Itoa(change.ChangeID)
}

func (change Change) GetTimestamp() (int64) {
	return change.CreatedAt.Unix()
}

func (change Change) GetNiceDate() (string) {
	return change.CreatedAt.Format(time.RFC822)
}

func GetLatestChanges(limit int) (changes []Change, err error) {

	client, ctx, err := getDSClient()
	if err != nil {
		return changes, err
	}

	q := datastore.NewQuery(KindChange).Order("-change_id").Limit(limit)

	client.GetAll(ctx, q, &changes)

	return changes, err
}

func GetChange(id string) (change *Change, err error) {

	client, context, err := getDSClient()
	if err != nil {
		return change, err
	}

	key := datastore.NameKey(KindChange, id, nil)

	change = new(Change)
	err = client.Get(context, key, change)
	if err != nil {
		return change, err
	}

	return change, nil
}
