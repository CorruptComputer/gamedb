package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
)

func DonateRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", donateHandler)
	return r
}

func donateHandler(w http.ResponseWriter, r *http.Request) {

	t := donateTemplate{}
	t.fill(w, r, "Donate", "Databases take up a tonne of memory and space. Help pay for the server costs or just buy me a beer.")
	t.Pages = []int{
		sql.UserLevelLimit0,
		sql.UserLevelLimit1,
		sql.UserLevelLimit2,
		sql.UserLevelLimit3,
	}

	err := returnTemplate(w, r, "donate", t)
	log.Err(err, r)
}

type donateTemplate struct {
	GlobalTemplate
	Pages []int
}
