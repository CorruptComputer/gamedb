package helpers

import (
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/log"
)

var steamClient *steam.Steam

func GetSteam() *steam.Steam {

	if steamClient == nil {

		steamClient = &steam.Steam{}
		steamClient.SetKey(config.Config.SteamAPIKey)
		steamClient.SetUserAgent("http://gamedb.online")
		steamClient.SetAPIRateLimit(time.Millisecond*1000, 10)
		steamClient.SetStoreRateLimit(time.Millisecond*1600, 10)
		steamClient.SetLogger(steamLogger{})
	}

	return steamClient
}

type steamLogger struct {
}

func (l steamLogger) Write(i steam.Log) {
	if config.Config.IsLocal() {
		// log.Info(i.String(), log.LogNameSteam)
	}
}

func HandleSteamStoreErr(err error, bytes []byte, allowedCodes []int) error {

	if err == steam.ErrHTMLResponse {
		log.Err(err, string(bytes))
	}

	err2, ok := err.(steam.Error)
	if ok {
		if allowedCodes != nil && SliceHasInt(allowedCodes, err2.Code) {
			return nil
		}
	}
	return err
}
