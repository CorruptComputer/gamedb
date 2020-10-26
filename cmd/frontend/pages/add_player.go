package pages

import (
	"net/http"
	"path"
	"strconv"

	"github.com/Jleagle/steam-go/steamid"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func playerAddHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {

		message := func() string {

			// Recaptcha
			// if config.IsProd() {
			// 	err = recaptcha.CheckFromRequest(r)
			// 	if err != nil {
			//
			// 		if err == recaptcha.ErrNotChecked {
			// 			return "Please check the captcha"
			// 		}
			//
			// 		return err.Error()
			// 	}
			// }

			// Parse form
			err := r.ParseForm()
			if err != nil {
				return err.Error()
			}

			search := r.PostFormValue("search")
			if search == "" {
				return "Please enter a search term"
			}

			search = path.Base(search)

			// Check if search term is a Steam ID
			id, err := steamid.ParsePlayerID(search)
			if err == nil && id > 0 {
				http.Redirect(w, r, "/players/"+strconv.FormatUint(uint64(id), 10), http.StatusFound)
				return ""
			}

			// Search Mongo
			player, _, err := mongo.SearchPlayer(search, bson.M{"_id": 1})
			if err == nil {
				http.Redirect(w, r, "/players/"+strconv.FormatInt(player.ID, 10), http.StatusFound)
				return ""
			}

			// This gets checked in mongo.SearchPlayer()
			// Check Steam API
			// resp, b, err := steam.GetSteam().ResolveVanityURL(search, steamapi.VanityURLProfile)
			// err = steam.AllowSteamCodes(err, b, nil)
			// if err == nil && resp.Success > 0 && resp.SteamID > 0 {
			//
			// 	http.Redirect(w, r, "/players/"+strconv.FormatInt(int64(resp.SteamID), 10), http.StatusFound)
			// 	return ""
			// }

			return "Player " + search + " not found on Steam"
		}()

		if message != "" {

			session.SetFlash(r, session.SessionBad, message)
			session.Save(w, r)

			http.Redirect(w, r, "/players/add", http.StatusFound)
			return
		}
	}

	t := addPlayerTemplate{}
	t.fill(w, r, "players_add", "Add Player", "Start tracking your stats in Game DB.")
	t.RecaptchaPublic = config.C.RecaptchaPublic
	t.Default = r.URL.Query().Get("search")

	//
	returnTemplate(w, r, t)
}

type addPlayerTemplate struct {
	globalTemplate
	RecaptchaPublic string
	Default         string
}
