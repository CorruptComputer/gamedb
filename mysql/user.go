package mysql

import (
	"time"
)

type User struct {
	PlayerID    int64      `gorm:"not null;primary_key"`
	CreatedAt   *time.Time `gorm:"not null"`
	UpdatedAt   *time.Time `gorm:"not null"`
	Email       string     `gorm:"not null;index:email"`
	Password    string     `gorm:"not null"`
	HideProfile bool       `gorm:"not null"`
	ShowAlerts  bool       `gorm:"not null"`
}

func GetUsersByEmail(email string) (users []User, err error) {

	db, err := GetDB()
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

	db, err := GetDB()
	if err != nil {
		return user, err
	}

	db.FirstOrCreate(&user, User{PlayerID: playerID})
	if db.Error != nil {
		return user, db.Error
	}

	return user, nil
}

func (u User) SaveOrUpdateUser() (err error) {

	db, err := GetDB()
	if err != nil {
		return err
	}

	user := new(User)
	db.Assign(u).FirstOrCreate(user, User{PlayerID: u.PlayerID})
	if db.Error != nil {
		return db.Error
	}

	return nil
}
