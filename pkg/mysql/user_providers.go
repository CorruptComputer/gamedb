package mysql

import (
	"strconv"
	"time"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/oauth"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
)

type UserProvider struct {
	UserID    int                `gorm:"not null;column:user_id;primary_key"`
	Provider  oauth.ProviderEnum `gorm:"not null;column:provider;primary_key"`
	CreatedAt time.Time          `gorm:"not null;column:created_at"`
	UpdatedAt time.Time          `gorm:"not null;column:updated_at"`
	DeletedAt *time.Time         `gorm:"not null;column:deleted_at"`
	Token     string             `gorm:"not null;column:token"`
	ID        string             `gorm:"not null;column:id"`
	Email     string             `gorm:"not null;column:email"`
	Username  string             `gorm:"not null;column:username"`
	Avatar    string             `gorm:"not null;column:avatar"`
}

func UpdateUserProvider(userID int, provider oauth.ProviderEnum, resp oauth.User) error {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	user := UserProvider{}
	user.UserID = userID
	user.Provider = provider
	user.Token = resp.Token
	user.ID = resp.ID
	user.Email = resp.Email
	user.Username = resp.Username
	user.Avatar = resp.Avatar

	db = db.Unscoped().Save(&user)
	return db.Error
}

func DeleteUserProvider(providerEnum oauth.ProviderEnum, userID int) (err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	db = db.Where("user_id = ?", userID)
	db = db.Where("provider = ?", providerEnum)
	db = db.Delete(&UserProvider{})

	return db.Error
}

func CheckExistingUserProvider(provider oauth.ProviderEnum, id string, userID int) (used bool, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return used, err
	}

	db = db.Where("provider = ?", provider)
	db = db.Where("id = ?", id)
	db = db.Where("user_id != ?", userID)
	db = db.First(&UserProvider{})

	return db.Error != ErrRecordNotFound, helpers.IgnoreErrors(db.Error, ErrRecordNotFound)
}

func GetUserProviders(userID int) (providers []UserProvider, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return providers, err
	}

	db = db.Where("user_id = ?", userID)
	db = db.Find(&providers)

	return providers, db.Error
}

func GetUserProviderByProviderID(provider oauth.ProviderEnum, providerID string) (userProvider UserProvider, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return userProvider, err
	}

	db = db.Where("provider = ?", provider)
	db = db.Where("id = ?", providerID)
	db = db.Find(&userProvider)

	return userProvider, db.Error
}

func GetUserProviderByUserID(enum oauth.ProviderEnum, userID int) (userProvider UserProvider, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return userProvider, err
	}

	db = db.Where("provider = ?", enum)
	db = db.Where("user_id = ?", userID)
	db = db.Find(&userProvider)

	return userProvider, db.Error
}

func GetUserSteamID(userID int) int64 {

	provider, err := GetUserProviderByUserID(oauth.ProviderSteam, userID)
	if err != nil {
		err = helpers.IgnoreErrors(err, ErrRecordNotFound)
		if err != nil {
			log.ErrS(err)
		}
		return 0
	}

	i, err := strconv.ParseInt(provider.ID, 10, 64)
	if err != nil {
		log.ErrS(err)
		return 0
	}

	return i
}
