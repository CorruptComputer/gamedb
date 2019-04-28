package pages

import (
	"net/http"
	"strconv"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
)

func DepotsRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", depotsHandler)
	r.Get("/{id}", depotHandler)
	return r
}

func depotsHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	// Template
	t := depotsTemplate{}
	t.fill(w, r, "Depots", "")

	err := returnTemplate(w, r, "depots", t)
	log.Err(err, r)
}

type depotsTemplate struct {
	GlobalTemplate
}

func depotHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid Depot ID."})
		return
	}

	idx, err := strconv.Atoi(id)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid Depot ID: " + id})
		return
	}

	// Template
	t := depotTemplate{}
	t.fill(w, r, "Depot", "")
	t.Depot = sql.Depot{}
	t.Depot.ID = idx

	err = returnTemplate(w, r, "depot", t)
	log.Err(err, r)
}

type depotTemplate struct {
	GlobalTemplate
	Depot sql.Depot
}
