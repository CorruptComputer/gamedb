package web

import (
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/go-chi/chi"
)

func newsRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", newsHandler)
	r.Get("/ajax", newsAjaxHandler)
	return r
}

func newsHandler(w http.ResponseWriter, r *http.Request) {

	t := newsTemplate{}
	t.Fill(w, r, "News", "All the news from all the games, all in one place.")

	err := returnTemplate(w, r, "news", t)
	log.Err(err, r)
}

type newsTemplate struct {
	GlobalTemplate
}

func newsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	query := DataTablesQuery{}
	err := query.FillFromURL(r.URL.Query())
	log.Err(err, r)

	var articles []db.News

	client, ctx, err := db.GetDSClient()
	if err != nil {

		log.Err(err, r)

	} else {

		q := datastore.NewQuery(db.KindNews).Order("-date").Limit(100)
		q, err = query.SetOffsetDS(q)
		if err != nil {

			log.Err(err, r)

		} else {

			_, err := client.GetAll(ctx, q, &articles)
			log.Err(err, r)

			for k, v := range articles {
				articles[k].Contents = helpers.BBCodeCompiler.Compile(v.Contents)
			}
		}
	}

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = "10000"
	response.RecordsFiltered = "10000"
	response.Draw = query.Draw

	for _, v := range articles {
		response.AddRow(v.OutputForJSON())
	}

	response.output(w, r)
}
