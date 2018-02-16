package mysql

import (
	"strconv"
	"time"
)

type Tag struct {
	ID        int        `gorm:"not null;column:id;primary_key;AUTO_INCREMENT"`
	CreatedAt *time.Time `gorm:"not null;column:created_at"`
	UpdatedAt *time.Time `gorm:"not null;column:updated_at"`
	Name      string     `gorm:"not null;column:name"`
	Games     int        `gorm:"not null;column:games"`
	Votes     int        `gorm:"not null;column:votes"`
}

func (tag Tag) GetPath() string {
	return "/apps?tag=" + strconv.Itoa(tag.ID)
}

func GetAllTags() (tags []Tag, err error) {

	db, err := getDB()
	if err != nil {
		return tags, err
	}

	db = db.Limit(1000).Order("id DESC").Find(&tags)
	if db.Error != nil {
		return tags, err
	}

	return tags, nil
}
