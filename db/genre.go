package db

import (
	"encoding/json"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/memcache"
)

type Genre struct {
	ID        int        `gorm:"not null;primary_key;AUTO_INCREMENT"`
	CreatedAt *time.Time `gorm:"not null"`
	UpdatedAt *time.Time `gorm:"not null"`
	DeletedAt *time.Time `gorm:""`
	Name      string     `gorm:"not null;index:name"`
	Apps      int        `gorm:"not null"`
	MeanPrice string     `gorm:"not null"`
	MeanScore string     `gorm:"not null"`
}

func (g Genre) GetPath() string {
	return "/games#genres=" + g.Name
}

func (g Genre) GetName() string {
	return g.Name
}

func (g Genre) GetMeanPrice(code steam.CountryCode) (string, error) {
	return helpers.GetMeanPrice(code, g.MeanPrice)
}

func (g Genre) GetMeanScore(code steam.CountryCode) (string, error) {
	return helpers.GetMeanScore(code, g.MeanScore)
}

func GetAllGenres() (genres []Genre, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return genres, err
	}

	db.Find(&genres)
	if db.Error != nil {
		return genres, db.Error
	}

	return genres, nil
}

func GetGenresForSelect() (genres []Genre, err error) {

	s, err := memcache.GetSetString(memcache.GenreKeyNames, func() (s string, err error) {

		db, err := GetMySQLClient()
		if err != nil {
			return s, err
		}

		var genres []Genre
		db = db.Select([]string{"id", "name"}).Order("name ASC").Find(&genres)
		if db.Error != nil {
			return s, db.Error
		}

		bytes, err := json.Marshal(genres)
		return string(bytes), err
	})

	if err != nil {
		return genres, err
	}

	err = helpers.Unmarshal([]byte(s), &genres)
	return genres, err
}

func DeleteGenres(ids []int) (err error) {

	if len(ids) == 0 {
		return nil
	}

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	db.Where("id IN (?)", ids).Delete(Genre{})

	return db.Error
}
