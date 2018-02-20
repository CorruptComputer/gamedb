package mysql

import (
	"os"

	"github.com/Jleagle/go-helpers/logger"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var gormConnection *gorm.DB

func init() {

	var err error
	gormConnection, err = gorm.Open("mysql", os.Getenv("STEAM_SQL_DSN")+"?parseTime=true")
	gormConnection.LogMode(true)
	if err != nil {
		logger.Error(err)
	}

}

func getDB() (conn *gorm.DB, err error) {

	if gormConnection == nil {

		db, err := gorm.Open("mysql", os.Getenv("STEAM_SQL_DSN")+"?parseTime=true")
		if err != nil {
			logger.Error(err)
			return db, nil
		}

		gormConnection = db
	}

	return gormConnection, nil
}
