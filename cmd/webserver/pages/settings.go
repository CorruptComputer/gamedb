package pages

import (
	"context"
	"net/http"
	"path"
	"strconv"
	"sync"

	"github.com/Jleagle/steam-go/steam"
	"github.com/badoux/checkmail"
	"github.com/gamedb/gamedb/cmd/webserver/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
	"github.com/mxpv/patreon-go"
	"github.com/yohcop/openid-go"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

func SettingsRouter() http.Handler {

	r := chi.NewRouter()
	r.Use(middlewareAuthCheck())

	r.Get("/", settingsHandler)
	r.Post("/", settingsPostHandler)
	r.Post("/delete", deletePostHandler)
	r.Get("/events.json", settingsEventsAjaxHandler)

	// r.Get("/link-steam", linkSteamHandler)
	r.Get("/steam-callback", linkSteamCallbackHandler)
	r.Get("/unlink-steam", unlinkSteamHandler)

	r.Get("/link-patreon", linkPatreonHandler)
	r.Get("/unlink-patreon", unlinkPatreonHandler)
	r.Get("/patreon-callback", linkPatreonCallbackHandler)

	return r
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	var err error

	//
	t := settingsTemplate{}
	t.fill(w, r, "Settings", "")
	t.addAssetPasswordStrength()
	t.setFlashes(w, r, false)
	t.Domain = config.Config.GameDBDomain.Get()

	// Get user
	t.User, err = getUserFromSession(r)
	log.Err(err)

	// Set Steam name to session if missing, can happen after linking
	if t.User.SteamID != 0 {

		name, err := session.Read(r, session.PlayerName)
		log.Err(err)

		if name == "" && err == nil {

			t.Player, err = mongo.GetPlayer(t.User.SteamID)
			err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
			log.Err(err)

			if t.Player.PersonaName != "" {
				err = session.Write(r, session.PlayerName, t.Player.VanintyURL)
				log.Err(err)
			}
		}
	}

	err = session.Save(w, r)
	log.Err(err)

	//
	var wg sync.WaitGroup

	// Get games
	wg.Add(1)
	go func() {

		defer wg.Done()

		if t.User.SteamID == 0 {
			return
		}

		playerApps, err := mongo.GetPlayerApps(t.User.SteamID, 0, 0, mongo.D{})
		if err != nil {
			log.Err(err, r)
			return
		}

		var appIDs []int
		for _, v := range playerApps {
			appIDs = append(appIDs, v.AppID)
		}

		t.Games = string(helpers.MarshalLog(appIDs))

	}()

	// Wait
	wg.Wait()

	// Countries
	for _, v := range helpers.GetActiveCountries() {
		t.Countries = append(t.Countries, []string{string(v), steam.Countries[v]})
	}

	// Template
	err = returnTemplate(w, r, "settings", t)
	log.Err(err, r)
}

type settingsTemplate struct {
	GlobalTemplate
	User      sql.User
	Player    mongo.Player
	Games     string
	Countries [][]string
	Domain    string
}

func deletePostHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	var err error

	redirect, good, bad := func() (redirect string, good string, bad string) {

		// Parse form
		err = r.ParseForm()
		if err != nil {
			log.Err(err)
			return "/settings", "", "There was an eror saving your information."
		}

		user, err := getUserFromSession(r)
		if err != nil {
			return "/settings", "", "There was an eror saving your information."
		}

		if r.PostForm.Get("id") == strconv.FormatInt(user.SteamID, 10) {

			err = session.Clear(r)
			log.Err(err)
			return "/", "Your account has been deleted", ""

		} else {
			return "/settings", "", "Invalid player ID."
		}
	}()

	if good != "" {
		err = session.SetGoodFlash(r, good)
		log.Err(err)
	}
	if bad != "" {
		err = session.SetBadFlash(r, bad)
		log.Err(err)
	}

	err = session.Save(w, r)
	log.Err(err)

	http.Redirect(w, r, redirect, http.StatusFound)
}

func settingsPostHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	redirect, good, bad := func() (redirect string, good string, bad string) {

		// Get user
		user, err := getUserFromSession(r)
		log.Err(err)
		if err != nil {
			return "/settings", "", "User not found"
		}

		// Parse form
		err = r.ParseForm()
		log.Err(err)
		if err != nil {
			return "/settings", "", "Could not read form data"
		}

		// Email
		email := r.PostForm.Get("email")

		if email != "" && email != user.Email {

			err = checkmail.ValidateFormat(r.PostForm.Get("email"))
			if err != nil {
				return "/settings", "", "Invalid email address"
			}

			user.Email = r.PostForm.Get("email")
		}

		// Password
		password := r.PostForm.Get("password")

		if email != user.Email {
			user.EmailVerified = false
		}

		if password != "" {

			if len(password) < 8 {
				return "/settings", "", "Password must be at least 8 characters long"
			}

			passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
			log.Err(err, r)
			if err != nil {
				return "/settings", "", "Something went wrong encrypting your password"
			}

			user.Password = string(passwordBytes)
		}

		// Country code
		code := r.PostForm.Get("country_code")

		if _, ok := steam.Countries[steam.CountryCode(code)]; ok {
			user.CountryCode = code
		} else {
			user.CountryCode = string(steam.CountryUS)
		}

		// Save hidden
		if r.PostForm.Get("hide") == "1" {
			user.HideProfile = true
		} else {
			user.HideProfile = false
		}

		// Save alerts
		if r.PostForm.Get("alerts") == "1" {
			user.ShowAlerts = true
		} else {
			user.ShowAlerts = false
		}

		// Save user
		db, err := sql.GetMySQLClient()
		log.Err(err)
		if err != nil {
			return "/settings", "", "We had trouble saving your settings"
		}

		// Have to save as a map because gorm does not save empty values otherwise
		db = db.Model(&user).Updates(map[string]interface{}{
			"email":        user.Email,
			"verified":     user.EmailVerified,
			"password":     user.Password,
			"hide_profile": user.HideProfile,
			"show_alerts":  user.ShowAlerts,
			"country_code": user.CountryCode,
		})

		log.Err(db.Error, r)
		if db.Error != nil {
			return "/settings", "", "Something went wrong saving your settings"
		}

		// Update session
		err = session.WriteMany(r, map[string]string{
			session.UserCountry:    user.CountryCode,
			session.UserEmail:      user.Email,
			session.UserShowAlerts: strconv.FormatBool(user.ShowAlerts),
		})
		if err != nil {
			log.Err(err, r)
			return "/settings", "", "Something went wrong saving your settings"
		}

		return "/settings", "Settings saved", ""
	}()

	if good != "" {
		err = session.SetGoodFlash(r, good)
		log.Err(err)
	}
	if bad != "" {
		err = session.SetBadFlash(r, bad)
		log.Err(err)
	}

	err = session.Save(w, r)
	log.Err(err)

	http.Redirect(w, r, redirect, http.StatusFound)
}

func settingsEventsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{"draw", "start"})
	if ret {
		return
	}

	user, err := getUserFromSession(r)
	if err != nil {
		log.Err(err)
		return
	}

	query := DataTablesQuery{}
	err = query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	var wg sync.WaitGroup

	// Get events
	var events []mongo.Event
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		events, err = mongo.GetEvents(user.ID, query.getOffset64())
		if err != nil {
			log.Err(err, r)
			return
		}

	}(r)

	// Get total
	var total int64
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		total, err = mongo.CountEvents(user.ID)
		log.Err(err, r)

	}(r)

	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.FormatInt(total, 10)
	response.RecordsFiltered = response.RecordsTotal
	response.Draw = query.Draw

	for _, v := range events {
		response.AddRow(v.OutputForJSON(r.RemoteAddr))
	}

	response.output(w, r)
}

// todo
// For the demo, we use in-memory infinite storage nonce and discovery
// cache. In your app, do not use this as it will eat up memory and never
// free it. Use your own implementation, on a better database system.
// If you have multiple servers for example, you may need to share at least
// the nonceStore between them.
var nonceStore = openid.NewSimpleNonceStore()
var discoveryCache = openid.NewSimpleDiscoveryCache()

func linkSteamCallbackHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {
		err := session.Save(w, r)
		log.Err(err)

		http.Redirect(w, r, "/settings", http.StatusFound)
	}()

	// Get Steam ID
	openID, err := openid.Verify(config.Config.GameDBDomain.Get()+r.URL.String(), discoveryCache, nonceStore)
	if err != nil {
		log.Err(err)
		err = session.SetBadFlash(r, "We could not verify your Steam account")
		log.Err(err)
		return
	}

	// Convert to int
	ID, err := strconv.ParseInt(path.Base(openID), 10, 64)
	if err != nil {
		log.Err(err)
		err = session.SetBadFlash(r, "An error occurred (1001)")
		log.Err(err)
		return
	}

	// Update user row
	db, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err)
		err = session.SetBadFlash(r, "An error occurred (1002)")
		log.Err(err)
		return
	}

	user, err := getUserFromSession(r)
	if err != nil {
		log.Err(err)
		err = session.SetBadFlash(r, "An error occurred (1003)")
		log.Err(err)
		return
	}

	db = db.Model(&user).Update("steam_id", ID)
	if db.Error != nil {
		err = session.SetBadFlash(r, "An error occurred (1004)")
		log.Err(err)
		log.Err(db.Error)
		return
	}

	// Success flash
	err = session.SetGoodFlash(r, "Steam account linked")
	log.Err(err)

	// Create event
	err = mongo.CreateUserEvent(r, user.ID, mongo.EventLinkSteam)
	if err != nil {
		log.Err(err, r)
	}

	// Queue for an update
	player, err := mongo.GetPlayer(ID)
	if err != nil {
		err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
		log.Err(err)
	} else {
		if player.ShouldUpdate(r.UserAgent(), mongo.PlayerUpdateManual) {
			err = queue.ProducePlayer(player.ID)
			log.Err(err, r)

			// Queued flash
			err = session.SetGoodFlash(r, "Player has been queued for an update")
			log.Err(err)
		}
	}

	// Update session
	err = session.Write(r, session.PlayerID, openID)
	log.Err(err)
}

func unlinkSteamHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {

		err := session.Save(w, r)
		log.Err(err)

		http.Redirect(w, r, "/settings", http.StatusFound)
	}()

	user, err := getUserFromSession(r)
	if err != nil {
		log.Err(err)
		err = session.SetBadFlash(r, "An error occurred")
		log.Err(err)
		return
	}

	db, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err)
		err = session.SetBadFlash(r, "An error occurred")
		log.Err(err)
		return
	}

	db = db.Model(&user).Update("steam_id", 0)
	if db.Error != nil {
		log.Err(db.Error)
		err = session.SetBadFlash(r, "An error occurred")
		log.Err(err)
		return
	}

	// Clear session
	err = session.Delete(r, session.PlayerID)
	log.Err(err)

	err = session.Delete(r, session.PlayerName)
	log.Err(err)

	err = session.Delete(r, session.PlayerLevel)
	log.Err(err)

	// Create event
	err = mongo.CreateUserEvent(r, user.ID, mongo.EventUnlinkSteam)
	if err != nil {
		log.Err(err, r)
	}

	// Flash message
	err = session.SetGoodFlash(r, "Steam unlinked")
	log.Err(err)
}

var (
	patreonConfig = oauth2.Config{
		ClientID:     config.Config.PatreonClientID.Get(),
		ClientSecret: config.Config.PatreonClientSecret.Get(),
		Scopes:       []string{"identity", "identity[email]"}, // identity[email] scope is only needed as the Patreon package we are using only handles v1 API
		RedirectURL: func() string {
			if config.IsLocal() {
				return "http://localhost:" + config.Config.WebserverPort.Get() + "/settings/patreon-callback"
			}
			return "https://gamedb.online/settings/patreon-callback"
		}(),
		Endpoint: oauth2.Endpoint{
			AuthURL:  patreon.AuthorizationURL,
			TokenURL: patreon.AccessTokenURL,
		},
	}
)

func linkPatreonHandler(w http.ResponseWriter, r *http.Request) {

	state := helpers.RandString(5, helpers.Numbers)

	err := session.Write(r, "patreon-oauth-state", state)
	log.Err(err)

	err = session.Save(w, r)
	log.Err(err)

	url := patreonConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

func linkPatreonCallbackHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {

		err := session.Save(w, r)
		log.Err(err)

		http.Redirect(w, r, "/settings", http.StatusFound)
	}()

	realState, err := session.Read(r, "patreon-oauth-state")
	if err != nil {
		log.Err(err)
		err = session.SetBadFlash(r, "An error occurred (1001)")
		log.Err(err)
		return
	}

	err = r.ParseForm()
	if err != nil {
		log.Err(err)
		err = session.SetBadFlash(r, "An error occurred (1002)")
		log.Err(err)
		return
	}

	state := r.Form.Get("state")
	if state != realState {
		err = session.SetBadFlash(r, "Invalid state")
		log.Err(err)
		return
	}

	code := r.Form.Get("code")
	if code == "" {
		err = session.SetBadFlash(r, "Invalid code")
		log.Err(err)
		return
	}

	token, err := patreonConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Err(err)
		err = session.SetBadFlash(r, "Invalid token")
		log.Err(err)
		return
	}

	db, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err)
		err = session.SetBadFlash(r, "An error occurred (1003)")
		log.Err(err)
		return
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token.AccessToken})
	tc := oauth2.NewClient(context.TODO(), ts)

	patreonUser, err := patreon.NewClient(tc).FetchUser()
	if err != nil {
		log.Err(err)
		err = session.SetBadFlash(r, "An error occurred (1004)")
		log.Err(err)
		return
	}

	user, err := getUserFromSession(r)
	if err != nil {
		log.Err(err)
		err = session.SetBadFlash(r, "An error occurred (1005)")
		log.Err(err)
		return
	}

	idx, err := strconv.Atoi(patreonUser.Data.ID)
	if err != nil {
		log.Err(err)
		err = session.SetBadFlash(r, "An error occurred (1005)")
		log.Err(err)
		return
	}

	db = db.Model(&user).Update("patreon_id", idx)
	if db.Error != nil {
		err = session.SetBadFlash(r, "An error occurred (1006)")
		log.Err(err)
		log.Err(db.Error)
		return
	}

	// Flash message
	err = session.SetGoodFlash(r, "Patreon account linked")
	log.Err(err)

	// Create event
	err = mongo.CreateUserEvent(r, user.ID, mongo.EventLinkPatreon)
	if err != nil {
		log.Err(err, r)
	}
}

func unlinkPatreonHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {
		err := session.Save(w, r)
		log.Err(err)

		http.Redirect(w, r, "/settings", http.StatusFound)
	}()

	user, err := getUserFromSession(r)
	if err != nil {
		log.Err(err)
		err = session.SetBadFlash(r, "An error occurred (1001)")
		log.Err(err)
		return
	}

	db, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err)
		err = session.SetBadFlash(r, "An error occurred (1002)")
		log.Err(err)
		return
	}

	db = db.Model(&user).Update("patreon_id", 0)
	if db.Error != nil {
		log.Err(err)
		log.Err(db.Error)

		err = session.SetBadFlash(r, "An error occurred (1003)")
		log.Err(err)
		return
	}

	// Flash message
	err = session.SetGoodFlash(r, "Patreon unlinked")
	log.Err(err)

	// Create event
	err = mongo.CreateUserEvent(r, user.ID, mongo.EventUnlinkPatreon)
	if err != nil {
		log.Err(err, r)
	}
}
