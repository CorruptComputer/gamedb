package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/tasks"
	"github.com/go-chi/chi"
)

func GenresRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", genresHandler)
	return r
}

func genresHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := tasks.TaskRegister[tasks.Genres{}.ID()].GetTaskConfig()
	if err != nil {
		err = helpers.IgnoreErrors(err, sql.ErrRecordNotFound)
		log.Err(err, r)
	}

	// Get genres
	genres, err := sql.GetAllGenres(false)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the genres.", Error: err})
		return
	}

	prices := map[int]string{}
	for _, v := range genres {
		price, err := v.GetMeanPrice(helpers.GetProductCC(r))
		log.Err(err, r)
		prices[v.ID] = price
	}

	// Template
	t := statsGenresTemplate{}
	t.fill(w, r, "Genres", "")
	t.Genres = genres
	t.Date = config.Value
	t.Prices = prices

	err = returnTemplate(w, r, "genres", t)
	log.Err(err, r)
}

type statsGenresTemplate struct {
	GlobalTemplate
	Genres []sql.Genre
	Date   string
	Prices map[int]string
}
