package pages

import (
	"net/http"

	"github.com/Jleagle/session-go/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
)

func LogoutRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", logoutHandler)
	return r
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {

	// Make event
	steamID, err := getUserIDFromSesion(r)
	if err != nil {
		log.Err(err, r)
	} else {
		err = mongo.CreateUserEvent(r, steamID, mongo.EventLogout)
		log.Err(err, r)
	}

	// Logout
	err = session.DeleteAll(r)
	log.Err(err, r)

	err = session.SetFlash(r, helpers.SessionGood, "You have been logged out")
	log.Err(err, r)

	err = session.Save(w, r)
	log.Err(err, r)

	//
	http.Redirect(w, r, "/", http.StatusFound)
}
