package web

import (
	"net/http"
	"time"

	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
	"github.com/gamedb/website/sql"
	"github.com/go-chi/chi"
)

func tagsRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", statsTagsHandler)
	return r
}

func statsTagsHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour*24)

	// Get config
	config, err := sql.GetConfig(sql.ConfTagsUpdated)
	log.Err(err, r)

	// Get tags
	tags, err := sql.GetAllTags()
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the tags.", Error: err})
		return
	}

	code := session.GetCountryCode(r)
	prices := map[int]string{}
	for _, v := range tags {
		price, err := v.GetMeanPrice(code)
		log.Err(err, r)
		prices[v.ID] = price
	}

	// Template
	t := statsTagsTemplate{}
	t.fill(w, r, "Tags", "")
	t.Tags = tags
	t.Date = config.Value
	t.Prices = prices

	err = returnTemplate(w, r, "tags", t)
	log.Err(err, r)
}

type statsTagsTemplate struct {
	GlobalTemplate
	Tags   []sql.Tag
	Date   string
	Prices map[int]string
}
