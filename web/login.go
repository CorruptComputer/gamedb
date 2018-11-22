package web

import (
	"errors"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/Jleagle/recaptcha-go"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/logging"
	"github.com/gamedb/website/session"
	"github.com/go-chi/chi"
	"github.com/spf13/viper"
	"github.com/yohcop/openid-go"
	"golang.org/x/crypto/bcrypt"
)

func loginRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", LoginHandler)
	r.Post("/", LoginPostHandler)
	r.Get("/openid", LoginOpenIDHandler)
	r.Get("/callback", LoginCallbackHandler)
	return r
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	t := loginTemplate{}
	t.Fill(w, r, "Login")
	t.RecaptchaPublic = viper.GetString("RECAPTCHA_PUBLIC")
	t.Domain = viper.GetString("DOMAIN")

	err := returnTemplate(w, r, "login", t)
	logging.Error(err)
}

type loginTemplate struct {
	GlobalTemplate
	RecaptchaPublic string
	Domain          string
}

var ErrInvalidCreds = errors.New("invalid username or password")
var ErrInvalidCaptcha = errors.New("please check the captcha")

func LoginPostHandler(w http.ResponseWriter, r *http.Request) {

	err := func() (err error) {

		// Parse form
		if err := r.ParseForm(); err != nil {
			return err
		}

		// Save email so they don't need to keep typing it
		err = session.Write(w, r, "login-email", r.PostForm.Get("email"))
		logging.Error(err)

		// Recaptcha
		err = recaptcha.CheckFromRequest(r)
		if err != nil {

			if err == recaptcha.ErrNotChecked {
				return ErrInvalidCaptcha
			}

			return err
		}

		// Field validation
		email := r.PostForm.Get("email")
		password := r.PostForm.Get("password")

		if email == "" || password == "" {
			return ErrInvalidCreds
		}

		// Get users that match the email
		users, err := db.GetUsersByEmail(email)
		if err != nil {
			return err
		}

		if len(users) == 0 {
			return ErrInvalidCreds
		}

		// Check password matches
		var user db.User
		var success bool
		for _, v := range users {

			err = bcrypt.CompareHashAndPassword([]byte(v.Password), []byte(password))
			if err == nil {
				success = true
				user = v
				break
			}
		}

		if !success {
			return ErrInvalidCreds
		}

		// Get player from user
		player, err := db.GetPlayer(user.PlayerID)
		if err != nil {
			return errors.New("no corresponding player")
		}

		// Log user in
		err = login(w, r, player, user)
		if err != nil {
			return err
		}

		// Remove form prefill on success
		err = session.Write(w, r, "login-email", "")
		logging.Error(err)

		return nil
	}()

	// Redirect
	if err != nil {
		time.Sleep(time.Second) // Stop brute forces

		if err != ErrInvalidCreds && err != ErrInvalidCaptcha {
			logging.Error(err)
		}

		err = session.SetGoodFlash(w, r, err.Error())
		logging.Error(err)
		http.Redirect(w, r, "/login", 302)
	} else {
		err = session.SetGoodFlash(w, r, "Login successful")
		logging.Error(err)
		http.Redirect(w, r, "/settings", 302)
	}
}

func LoginOpenIDHandler(w http.ResponseWriter, r *http.Request) {

	loggedIn, err := session.IsLoggedIn(r)
	if err != nil {
		logging.Error(err)
	}

	if loggedIn {
		http.Redirect(w, r, "/settings", 303)
		return
	}

	var url string
	url, err = openid.RedirectURL("https://steamcommunity.com/openid", viper.GetString("DOMAIN")+"/login/callback", viper.GetString("DOMAIN")+"/")
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "Something went wrong sending you to Steam.", Error: err})
		return
	}

	http.Redirect(w, r, url, 303)
}

// todo
// For the demo, we use in-memory infinite storage nonce and discovery
// cache. In your app, do not use this as it will eat up memory and never
// free it. Use your own implementation, on a better database system.
// If you have multiple servers for example, you may need to share at least
// the nonceStore between them.
var nonceStore = openid.NewSimpleNonceStore()
var discoveryCache = openid.NewSimpleDiscoveryCache()

func LoginCallbackHandler(w http.ResponseWriter, r *http.Request) {

	// Get ID from OpenID
	openID, err := openid.Verify(viper.GetString("DOMAIN")+r.URL.String(), discoveryCache, nonceStore)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "We could not verify your Steam account.", Error: err})
		return
	}

	// Convert to int
	idInt, err := strconv.ParseInt(path.Base(openID), 10, 64)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "We could not verify your Steam account.", Error: err})
		return
	}

	// Check if we have the player
	player, err := db.GetPlayer(idInt)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "We could not verify your Steam account.", Error: err})
		return
	}

	// Get player if they're new
	if player.PersonaName == "" {
		err = queuePlayer(r, player, db.PlayerUpdateAuto)
		logging.Error(err)
	}

	// Get user
	gorm, err := db.GetMySQLClient()
	if err != nil {
		logging.Error(err)
	}

	var user db.User
	gorm = gorm.First(&user, idInt)
	if gorm.Error != nil {
		logging.Error(gorm.Error)
	}

	err = login(w, r, player, user)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an error logging you in.", Error: err})
		return
	}

	http.Redirect(w, r, "/settings", 302)
}

func login(w http.ResponseWriter, r *http.Request, player db.Player, user db.User) (err error) {

	// Save session
	err = session.WriteMany(w, r, map[string]string{
		session.PlayerID:    strconv.FormatInt(player.PlayerID, 10),
		session.PlayerName:  player.PersonaName,
		session.PlayerLevel: strconv.Itoa(player.Level),
		session.UserCountry: user.CountryCode,
	})
	if err != nil {
		return err
	}

	// Create login record
	err = db.CreateEvent(r, player.PlayerID, db.EventLogin)
	if err != nil {
		return err
	}

	return nil
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {

	id, err := getPlayerIDFromSession(r)
	err = helpers.IgnoreErrors(err, errNotLoggedIn)
	logging.Error(err)

	err = db.CreateEvent(r, id, db.EventLogout)
	logging.Error(err)

	err = session.Clear(w, r)
	logging.Error(err)

	http.Redirect(w, r, "/", 303)
}
