package sql

import (
	"errors"
	"time"

	"github.com/cenkalti/backoff/v3"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
)

const (
	apiSessionLength  = time.Second * 70 // Time to keep the key for
	apiSessionRefresh = time.Second * 60 // Heartbeat to retake the key
	apiSessionRetry   = time.Second * 10 // Retry on no keys availabile

	sqlKeyName = "api_keys"
)

type APIKey struct {
	Key     string    `gorm:"not null;column:key;PRIMARY_KEY"`
	Expires time.Time `gorm:"not null;column:expires;type:datetime"`
	Owner   string    `gorm:"not null;column:owner"`
	Notes   string    `gorm:"-"`
}

func GetAPIKey(tag string, getUnusedKey bool) (err error) {

	tag = config.Config.Environment.Get() + "-" + tag

	if config.IsLocal() {
		getUnusedKey = false
	}

	// Retrying as this call can fail
	operation := func() (err error) {

		db, err := GetMySQLClient()
		if err != nil {
			return err
		}

		// https://stackoverflow.com/questions/7698211/prevent-two-calls-to-the-same-script-from-selecting-the-same-mysql-row
		db = db.New().Raw("SELECT GET_LOCK('" + sqlKeyName + "', 10) as `lock`")
		if db.Error != nil {
			return db.Error
		}

		defer func() {
			db = db.New().Raw("SELECT RELEASE_LOCK('" + sqlKeyName + "')")
			if db.Error != nil {
				log.Err(db.Error)
			}
		}()

		db = db.New()

		// Get key
		var row = APIKey{}

		if getUnusedKey {
			db = db.Where("expires < ?", time.Now())
		}

		db = db.Order("expires ASC").First(&row)
		if db.Error == ErrRecordNotFound {
			return errors.New("waiting for API key")
		} else if db.Error != nil {
			return db.Error
		}

		// Update key
		if getUnusedKey {

			db = db.New().Table("api_keys").Where("`key` = ?", row.Key).Updates(map[string]interface{}{
				"expires": time.Now().Add(apiSessionLength),
				"owner":   tag,
			})
			if db.Error != nil {
				return db.Error
			}
		}

		config.Config.SteamAPIKey.SetDefault(row.Key)

		log.Info("Using key: " + config.GetSteamKeyTag())

		return nil
	}

	policy := backoff.NewConstantBackOff(apiSessionRetry)
	err = backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Warning(err) })
	if err != nil {
		return err
	}

	// Keep the key in use with a heartbeat
	if getUnusedKey {

		db, err := GetMySQLClient()
		if err != nil {
			return err
		}

		go func() {
			for {
				time.Sleep(apiSessionRefresh)

				// Update key
				db = db.Model(&APIKey{}).Where("`key` = ?", config.Config.SteamAPIKey.Get()).Updates(map[string]interface{}{
					"expires": time.Now().Add(apiSessionLength),
					"owner":   tag,
				})
				if db.Error != nil {
					log.Err(db.Error)
				}
			}
		}()
	}

	return err
}
