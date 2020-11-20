package steam

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gocolly/colly/v2"
	"go.uber.org/zap"
)

func AllowSteamCodes(err error, allowedCodes ...int) error {

	// if err == steam.ErrHTMLResponse {
	// 	log.ErrS(err, string(bytes))
	// 	time.Sleep(time.Second * 30)
	// }

	err2, ok := err.(steamapi.Error)
	if ok {
		if allowedCodes != nil && helpers.SliceHasInt(allowedCodes, err2.Code) {
			return nil
		}
	}
	return err
}

// Downgrade some Steam errors to info.
func LogSteamError(err error, interfaces ...interface{}) {

	isError := func() bool {

		// Sleeps on rate limits etc
		if val, ok := err.(steamapi.Error); ok {

			if val.Code == 429 { // Rate limit
				time.Sleep(time.Second * 30)
			} else {
				time.Sleep(time.Second * 5)
			}

			return false
		}

		steamErrors := []string{
			"Bad Gateway",
			"Client.Timeout exceeded while awaiting headers",
			"read: connection reset by peer",
			"connect: connection timed out",
			"Gateway Timeout",
			"html response",
			"i/o timeout",
			"Internal Server Error",
			"invalid character '<' looking for beginning of value",
			"net/http: request canceled (Client.Timeout exceeded while reading body)",
			"remote error: tls: internal error",
			"Service Unavailable",
			"something went wrong",
			"net/http: TLS handshake timeout",
			"unexpected end of JSON input",
			"EOF",
			"write: connection reset by peer",
			"steam: store: null response",
		}

		for _, v := range steamErrors {
			if strings.Contains(err.Error(), v) {
				return false
			}
		}

		return true
	}()

	// Prepend error
	interfaces = append([]interface{}{err}, interfaces...)

	// These don't use the log helper to fix the extra stack offset
	if isError {
		zap.S().Error(interfaces...)
	} else {
		zap.S().Info(interfaces...)
	}
}

func WithAgeCheckCookie(c *colly.Collector) {

	cookieURL, _ := url.Parse("https://store.steampowered.com")

	jar, _ := cookiejar.New(nil)

	jar.SetCookies(cookieURL, []*http.Cookie{
		{Name: "birthtime", Value: "536457601", Path: "/", Domain: "store.steampowered.com"},
		{Name: "lastagecheckage", Value: "1-January-1987", Path: "/", Domain: "store.steampowered.com"},
		{Name: "mature_content", Value: "1", Path: "/", Domain: "store.steampowered.com"},
	})

	c.SetCookieJar(jar)
}

var WithTimeout = func(seconds int) func(c *colly.Collector) {

	if seconds == 0 {
		seconds = 10
	}

	return func(c *colly.Collector) {
		c.SetRequestTimeout(time.Duration(seconds) * time.Second)
	}
}
