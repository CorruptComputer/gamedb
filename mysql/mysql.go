package mysql

import (
	"errors"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	gormConnection *gorm.DB
	debug          = false

	ErrNotFound = errors.New("not found")
)

func SetDebug(val bool) {
	debug = val
	return
}

func GetDB() (conn *gorm.DB, err error) {

	if gormConnection == nil {

		db, err := gorm.Open("mysql", os.Getenv("STEAM_MYSQL_DSN"))
		db.LogMode(debug)
		if err != nil {
			return db, nil
		}

		gormConnection = db
	}

	return gormConnection, nil
}

type UpdateError struct {
	err  string
	hard bool
	log  bool
}

func (e UpdateError) Error() string {
	return e.err
}

func (e UpdateError) IsHard() bool {
	return e.hard
}

func (e UpdateError) IsSoft() bool {
	return !e.hard
}

func (e UpdateError) Log() bool {
	return e.log
}
