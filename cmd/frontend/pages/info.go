package pages

import (
	"net/http"

	"github.com/go-chi/chi"
)

func InfoRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", infoHandler)
	return r
}

func infoHandler(w http.ResponseWriter, r *http.Request) {

	t := globalTemplate{}
	t.fill(w, r, "info", "Info", "Game DB Information")

	returnTemplate(w, r, t)
}
