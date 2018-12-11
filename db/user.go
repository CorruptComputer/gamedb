package db

import (
	"time"
)

type User struct {
	PlayerID    int64     `gorm:"not null;primary_key"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
	Email       string    `gorm:"not null;index:email"`
	Password    string    `gorm:"not null"`
	HideProfile int8      `gorm:"not null"`
	ShowAlerts  int8      `gorm:"not null"`
	CountryCode string    `gorm:"not null"`
}

func (u User) Save() error {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	db = db.Assign(u).FirstOrCreate(&u)
	if db.Error != nil {
		return db.Error
	}

	return nil
}

func GetUsersByEmail(email string) (users []User, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return users, err
	}

	db = db.Limit(100).Where("email = (?)", email).Order("created_at ASC").Find(&users)
	if db.Error != nil {
		return users, db.Error
	}

	return users, nil
}

func GetUser(playerID int64) (user User, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return user, err
	}

	db.FirstOrCreate(&user, User{PlayerID: playerID})
	if db.Error != nil {
		return user, db.Error
	}

	return user, nil
}
