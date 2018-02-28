package steam

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/kr/pretty"
)

func GetNewsForApp(id string) (articles []GetNewsForAppArticle, err error) {

	options := url.Values{}
	options.Set("appid", id)
	options.Set("count", "20")

	bytes, err := get("ISteamNews/GetNewsForApp/v2/", options)
	if err != nil {
		return articles, err
	}

	// Unmarshal JSON
	var resp *GetNewsForAppBody
	if err := json.Unmarshal(bytes, &resp); err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			pretty.Print(string(bytes))
		}
		return articles, err
	}

	return resp.App.Items, nil
}

type GetNewsForAppBody struct {
	App struct {
		Appid int                    `json:"appid"`
		Items []GetNewsForAppArticle `json:"newsitems"`
		Count int                    `json:"count"`
	} `json:"appnews"`
}

type GetNewsForAppArticle struct {
	Gid           string `json:"gid"`
	Title         string `json:"title"`
	URL           string `json:"url"`
	IsExternalURL bool   `json:"is_external_url"`
	Author        string `json:"author"`
	Contents      string `json:"contents"`
	Feedlabel     string `json:"feedlabel"`
	Date          int    `json:"date"`
	Feedname      string `json:"feedname"`
	FeedType      int    `json:"feed_type"`
	Appid         int    `json:"appid"`
}
