package pages

import (
	"net/http"
	"time"

	"github.com/Jleagle/recaptcha-go"
	"github.com/Jleagle/session-go/session"
	"github.com/badoux/checkmail"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"golang.org/x/crypto/bcrypt"
)

func ForgotRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", forgotHandler)
	r.Post("/", forgotPostHandler)
	r.Get("/reset", forgotResetPasswordHandler)

	return r
}

func forgotHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	t := forgotTemplate{}
	t.fill(w, r, "Forgot Password", "")
	t.hideAds = true
	t.RecaptchaPublic = config.Config.RecaptchaPublic.Get()

	t.LoginEmail, err = session.Get(r, "login-email")
	log.Err(err, r)

	returnTemplate(w, r, "forgot", t)
}

type forgotTemplate struct {
	GlobalTemplate
	RecaptchaPublic string
	LoginEmail      string
}

func forgotPostHandler(w http.ResponseWriter, r *http.Request) {

	time.Sleep(time.Second)

	message, success := func() (message string, success bool) {

		// Parse form
		err := r.ParseForm()
		if err != nil {
			log.Err(err, r)
			return "An error occurred", false
		}

		email := r.PostForm.Get("email")

		// Field validation
		if email == "" {
			return "Please fill in your email address", false
		}

		err = checkmail.ValidateFormat(email)
		if err != nil {
			return "Invalid email address", false
		}

		if config.IsProd() {
			err = recaptcha.CheckFromRequest(r)
			if err != nil {
				return "Please check the captcha", false
			}
		}

		// Find user
		user, err := sql.GetUserByKey("email", email, 0)
		if err == sql.ErrRecordNotFound {
			return "Email sent", true
		} else if err != nil {
			log.Err(err, r)
			return "An error occurred", false
		}

		// Create verification code
		code, err := sql.CreateUserVerification(user.ID)
		if err != nil {
			log.Err(err, r)
			return "An error occurred", false
		}

		// Send email
		body := "You are someone else has requested a new password for Game DB.<br>This link will reset your password: " +
			config.Config.GameDBDomain.Get() + "/forgot/reset?code=" + code.Code

		_, err = helpers.SendEmail(
			mail.NewEmail(email, email),
			mail.NewEmail("Game DB", "no-reply@gamedb.online"),
			"Game DB Forgotten Password",
			body,
		)
		if err != nil {
			log.Err(err, r)
			return "An error occurred", false
		}

		// Create login event
		err = mongo.CreateUserEvent(r, user.ID, mongo.EventForgotPassword)
		if err != nil {
			log.Err(err, r)
		}

		return "Email sent", true
	}()

	//
	if success {

		err := session.SetFlash(r, helpers.SessionGood, message)
		log.Err(err, r)

		err = session.Save(w, r)
		log.Err(err, r)

		http.Redirect(w, r, "/login", http.StatusFound)

	} else {

		err := session.SetFlash(r, helpers.SessionBad, message)
		log.Err(err, r)

		err = session.Save(w, r)
		log.Err(err, r)

		http.Redirect(w, r, "/forgot", http.StatusFound)
	}
}

func forgotResetPasswordHandler(w http.ResponseWriter, r *http.Request) {

	time.Sleep(time.Second)

	message, success := func() (message string, success bool) {

		// Validate code
		code := r.URL.Query().Get("code")

		if len(code) != 10 {
			return "Invalid code (1001)", false
		}

		// Find email from code
		userID, err := sql.GetUserVerification(code)
		if err != nil {
			err = helpers.IgnoreErrors(err, sql.ErrRecordNotFound)
			log.Err(err, r)
			return "Invalid code (1002)", false
		}

		// if userVerify.Expires.Unix() < time.Now().Unix() {
		// return "This verify code has expired", false
		// }

		// Get user
		user, err := sql.GetUserByID(userID)
		if err != nil {
			err = helpers.IgnoreErrors(err, sql.ErrRecordNotFound)
			log.Err(err, r)
			return "An error occurred (1001)", false
		}

		// Create password
		passwordString := helpers.RandString(16, helpers.Letters)
		passwordBytes, err := bcrypt.GenerateFromPassword([]byte(passwordString), 14)
		if err != nil {
			log.Err(err, r)
			return "An error occurred (1002)", false
		}

		// Send email
		body := "Your new Game DB password is: " + passwordString

		_, err = helpers.SendEmail(
			mail.NewEmail(user.Email, user.Email),
			mail.NewEmail("Game DB", "no-reply@gamedb.online"),
			"Game DB Forgotten Password",
			body,
		)
		if err != nil {
			log.Err(err, r)
			return "An error occurred", false
		}

		// Set password
		err = sql.UpdateUserCol(userID, "password", string(passwordBytes))
		if err != nil {
			log.Err(err, r)
			return "An error occurred (1003)", false
		}

		//
		return "A new password has been emailed to you", true
	}()

	//
	if success {

		err := session.SetFlash(r, helpers.SessionGood, message)

		err = session.Save(w, r)
		log.Err(err, r)

		http.Redirect(w, r, "/login", http.StatusFound)

	} else {

		err := session.SetFlash(r, helpers.SessionBad, message)
		log.Err(err, r)

		err = session.Save(w, r)
		log.Err(err, r)

		http.Redirect(w, r, "/signup", http.StatusFound)
	}
}
