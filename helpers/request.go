package helpers

import "regexp"

func IsBot(userAgent string) bool {

	if userAgent == "" {
		return false
	}

	r, _ := regexp.Compile("/bot|crawl|slurp|spider|google|msn|bing|yahoo|jeeves|facebook/i")
	if r.MatchString(userAgent) {
		return true
	}

	return false
}
