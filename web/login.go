package web

import (
	"errors"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/Jleagle/recaptcha-go"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/queue"
	"github.com/gamedb/website/session"
	"github.com/go-chi/chi"
	"github.com/yohcop/openid-go"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	recaptcha.SetSecret(config.Config.RecaptchaPrivate)
}

func loginRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", loginHandler)
	r.Post("/", loginPostHandler)
	r.Get("/openid", loginOpenIDHandler)
	r.Get("/callback", loginOpenIDCallbackHandler)
	return r
}

func loginHandler(w http.ResponseWriter, r *http.Request) {

	t := loginTemplate{}
	t.Fill(w, r, "Login", "Login to Game DB to set your currency and other things.")
	t.RecaptchaPublic = config.Config.RecaptchaPublic
	t.Domain = config.Config.GameDBDomain.Get()

	err := returnTemplate(w, r, "login", t)
	log.Err(err, r)
}

type loginTemplate struct {
	GlobalTemplate
	RecaptchaPublic string
	Domain          string
}

var ErrInvalidCreds = errors.New("invalid username or password")
var ErrInvalidCaptcha = errors.New("please check the captcha")

func loginPostHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	err := func() (err error) {

		// Parse form
		err = r.ParseForm()
		if err != nil {
			return err
		}

		// Save email so they don't need to keep typing it
		err = session.Write(w, r, "login-email", r.PostForm.Get("email"))
		log.Err(err, r)

		// Recaptcha
		if config.Config.IsProd() {
			err = recaptcha.CheckFromRequest(r)
			if err != nil {

				if err == recaptcha.ErrNotChecked {
					return ErrInvalidCaptcha
				}

				return err
			}
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
		log.Err(err, r)

		return nil
	}()

	// Redirect
	if err != nil {

		err2 := helpers.IgnoreErrors(err, ErrInvalidCreds, ErrInvalidCaptcha)
		log.Err(err2)

		// Stop brute forces
		time.Sleep(time.Second)

		err = session.SetGoodFlash(w, r, err.Error())
		log.Err(err, r)

		http.Redirect(w, r, "/login", 302)

	} else {

		err = session.SetGoodFlash(w, r, "Login successful")
		log.Err(err, r)

		http.Redirect(w, r, "/settings", 302)
	}
}

func loginOpenIDHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	loggedIn, err := session.IsLoggedIn(r)
	if err != nil {
		log.Err(err, r)
	}

	if loggedIn {
		http.Redirect(w, r, "/settings", 303)
		return
	}

	var url string
	var domain = config.Config.GameDBDomain.Get()
	url, err = openid.RedirectURL("https://steamcommunity.com/openid", domain+"/login/callback", domain+"/")
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

func loginOpenIDCallbackHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	// Get ID from OpenID
	openID, err := openid.Verify(config.Config.GameDBDomain.Get()+r.URL.String(), discoveryCache, nonceStore)
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

	// Queue for an update
	err = player.ShouldUpdate(r.UserAgent(), db.PlayerUpdateAuto)
	if err != nil {
		err = helpers.IgnoreErrors(err, db.ErrUpdatingPlayerTooSoon, db.ErrUpdatingPlayerInQueue, db.ErrUpdatingPlayerBot)
		log.Err(err, r)
	} else {
		err = queue.QueuePlayer(player.PlayerID)
		log.Err(err, r)
	}

	// Get user
	gorm, err := db.GetMySQLClient()
	log.Err(err, r)

	var user db.User
	gorm = gorm.First(&user, idInt)
	log.Err(gorm.Error)

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
		session.UserEmail:   user.Email,
		session.UserCountry: user.CountryCode,
	})

	if err != nil {
		return err
	}

	// Create login record
	return db.CreateEvent(r, player.PlayerID, db.EventLogin)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	id, err := getPlayerIDFromSession(r)
	err = helpers.IgnoreErrors(err, errNotLoggedIn)
	log.Err(err, r)

	err = db.CreateEvent(r, id, db.EventLogout)
	log.Err(err, r)

	err = session.Clear(w, r)
	log.Err(err, r)

	http.Redirect(w, r, "/", 303)
}
