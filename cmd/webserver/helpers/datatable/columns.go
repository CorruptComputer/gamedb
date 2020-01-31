package datatable

import (
	"go.mongodb.org/mongo-driver/bson"
)

type Columns struct {
	defaultCol string
	columns    map[string]Column
}

type Column struct {
	sortAsc     bool
	sortDesc    bool
	sortDefault bool
	sortAppend  bson.D
	filters     bson.D
}
