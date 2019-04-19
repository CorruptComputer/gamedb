package pages

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"sync"

	"github.com/Jleagle/steam-go/steam"
	"github.com/badoux/checkmail"
	"github.com/gamedb/website/pkg/helpers"
	"github.com/gamedb/website/pkg/log"
	"github.com/gamedb/website/pkg/mongo"
	"github.com/gamedb/website/pkg/session"
	"github.com/gamedb/website/pkg/sql"
	"github.com/go-chi/chi"
	"golang.org/x/crypto/bcrypt"
)

var (
	errNotLoggedIn = errors.New("not logged in")
)

func SettingsRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", settingsHandler)
	r.Post("/", settingsPostHandler)
	r.Get("/events.json", settingsEventsAjaxHandler)
	return r
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, 0)

	player, err := getPlayer(r)
	if err != nil {
		if err == errNotLoggedIn {
			err := session.SetBadFlash(w, r, "please login")
			log.Err(err, r)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving your data.", Error: err})
		return
	}

	//
	var wg sync.WaitGroup

	// Get donations
	var donations []sql.Donation
	wg.Add(1)
	go func(player mongo.Player) {

		defer wg.Done()

		// if player.Donated > 0 {
		// 	donations, err = db.GetDonations(player.ID, 10)
		// 	log.Err(err, r)
		// }

	}(player)

	// Get games
	var games string
	wg.Add(1)
	go func(player mongo.Player) {

		defer wg.Done()

		playerApps, err := mongo.GetPlayerApps(player.ID, 0, 0, mongo.D{})
		if err != nil {
			log.Err(err, r)
			return
		}

		var appIDs []int
		for _, v := range playerApps {
			appIDs = append(appIDs, v.AppID)
		}

		games = string(helpers.MarshalLog(appIDs))

	}(player)

	// Get User
	var user sql.User
	wg.Add(1)
	go func(player mongo.Player) {

		defer wg.Done()

		user, err = getUser(r, 0)
		if err != nil {
			log.Err(err, r)
			return
		}

	}(player)

	// Wait
	wg.Wait()

	// Countries
	var countries [][]string
	for k, v := range steam.Countries {
		countries = append(countries, []string{string(k), v})
	}
	sort.Slice(countries, func(i, j int) bool {
		return countries[i][1] < countries[j][1]
	})

	// Template
	t := settingsTemplate{}
	t.fill(w, r, "Settings", "")
	t.addAssetPasswordStrength()
	t.Player = player
	t.User = user
	t.Donations = donations
	t.Games = games
	t.Countries = countries

	err = returnTemplate(w, r, "settings", t)
	log.Err(err, r)
}

type settingsTemplate struct {
	GlobalTemplate
	Player    mongo.Player
	User      sql.User
	Donations []sql.Donation
	Games     string
	Messages  []interface{}
	Countries [][]string
}

func settingsPostHandler(w http.ResponseWriter, r *http.Request) {

	setCacheHeaders(w, 0)

	// Get user
	user, err := getUser(r, 0)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an eror saving your information.", Error: err})
		return
	}

	// Parse form
	err = r.ParseForm()
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an eror saving your information.", Error: err})
		return
	}

	// Email
	email := r.PostForm.Get("email")

	err = checkmail.ValidateFormat(r.PostForm.Get("email"))
	if err != nil {
		err = session.SetBadFlash(w, r, "Password must be at least 8 characters long")
		http.Redirect(w, r, "/settings", http.StatusFound)
		return
	}

	user.Email = r.PostForm.Get("email")

	// Password
	password := r.PostForm.Get("password")

	if email != user.Email {
		user.Verified = false
	}

	if len(password) > 0 {

		if len(password) < 8 {
			err := session.SetBadFlash(w, r, "Password must be at least 8 characters long")
			log.Err(err, r)
			http.Redirect(w, r, "/settings", http.StatusFound)
			return
		}

		passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
		if err != nil {
			log.Err(err, r)
			err := session.SetBadFlash(w, r, "Something went wrong encrypting your password")
			log.Err(err, r)
			http.Redirect(w, r, "/settings", http.StatusFound)
			return
		}

		user.Password = string(passwordBytes)
	}

	// Country code
	user.CountryCode = r.PostForm.Get("country_code")

	// Save hidden
	if r.PostForm.Get("hide") == "1" {
		user.HideProfile = 1
	} else {
		user.HideProfile = 0
	}

	// Save alerts
	if r.PostForm.Get("alerts") == "1" {
		user.ShowAlerts = 1
	} else {
		user.ShowAlerts = 0
	}

	// Save user
	err = user.Save()
	log.Err(err, r)
	if err != nil {
		err = session.SetBadFlash(w, r, "Something went wrong saving settings")
		log.Err(err, r)
	} else {
		err = session.SetGoodFlash(w, r, "Settings saved")
		log.Err(err, r)
	}

	// Update session
	err = session.WriteMany(w, r, map[string]string{
		session.UserCountry: user.CountryCode,
	})
	log.Err(err, r)

	http.Redirect(w, r, "/settings", http.StatusFound)
}

func settingsEventsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{"draw", "start"})
	if ret {
		return
	}

	setCacheHeaders(w, 0)

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	// Get player ID
	playerID, err := getPlayerIDFromSession(r)
	if err != nil {
		log.Err(err, r)
		return
	}

	var wg sync.WaitGroup

	// Get events
	var events []mongo.Event
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		events, err = mongo.GetEvents(playerID, query.getOffset64())
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

		total, err = mongo.CountEvents(playerID)
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

func getPlayerIDFromSession(r *http.Request) (playerID int64, err error) {

	// Check if logged in
	loggedIn, err := session.IsLoggedIn(r)
	if err != nil {
		return playerID, errNotLoggedIn
	}

	if !loggedIn {
		return playerID, errNotLoggedIn
	}

	// Get session
	id, err := session.Read(r, session.PlayerID)
	if err != nil {
		return playerID, err
	}

	// Convert ID
	return strconv.ParseInt(id, 10, 64)
}

func getPlayer(r *http.Request) (player mongo.Player, err error) {

	playerID, err := getPlayerIDFromSession(r)
	if err != nil {
		return player, err
	}

	return mongo.GetPlayer(playerID)
}

func getUser(r *http.Request, playerID int64) (user sql.User, err error) {

	if playerID == 0 {
		playerID, err = getPlayerIDFromSession(r)
		if err != nil {
			return user, err
		}
	}

	user, err = sql.GetUser(playerID)
	if err != nil {
		return user, err
	}

	return user, nil
}
