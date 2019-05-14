package pages

import (
	"errors"
	"net/http"

	"github.com/Jleagle/recaptcha-go"
	"github.com/Jleagle/session-go/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func ContactRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", contactHandler)
	r.Post("/", postContactHandler)
	return r
}

func contactHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	t := contactTemplate{}
	t.fill(w, r, "Contact", "Get in touch with Game DB.")
	t.RecaptchaPublic = config.Config.RecaptchaPublic.Get()
	t.setFlashes(w, r, true)

	var err error

	t.SessionName, err = session.Get(r, "contact-name")
	log.Err(err)

	t.SessionEmail, err = session.Get(r, "contact-email")
	log.Err(err)

	if t.SessionEmail == "" {
		t.SessionEmail, err = session.Get(r, helpers.SessionUserEmail)
		log.Err(err)
	}

	t.SessionMessage, err = session.Get(r, "contact-message")
	log.Err(err)

	err = returnTemplate(w, r, "contact", t)
	log.Err(err, r)
}

type contactTemplate struct {
	GlobalTemplate
	RecaptchaPublic string
	Messages        []string
	Success         bool
	SessionName     string
	SessionEmail    string
	SessionMessage  string
}

func postContactHandler(w http.ResponseWriter, r *http.Request) {

	err := func() (err error) {

		var ErrSomething = errors.New("something went wrong")

		// Parse form
		err = r.ParseForm()
		if err != nil {
			log.Err(err, r)
			return err
		}

		// Backup
		err = session.SetMany(r, map[string]interface{}{
			"contact-name":    r.PostForm.Get("name"),
			"contact-email":   r.PostForm.Get("email"),
			"contact-message": r.PostForm.Get("message"),
		})
		log.Err(err, r)

		// Form validation
		if r.PostForm.Get("name") == "" {
			return errors.New("Please fill in your name")
		}
		if r.PostForm.Get("email") == "" {
			return errors.New("Please fill in your email")
		}
		if r.PostForm.Get("message") == "" {
			return errors.New("Please fill in a message")
		}

		// Recaptcha
		if config.IsProd() {
			err = recaptcha.CheckFromRequest(r)
			if err != nil {

				if err == recaptcha.ErrNotChecked {
					return errors.New("please check the captcha")
				}

				log.Err(err, r)
				return ErrSomething
			}
		}

		// Send
		message := mail.NewSingleEmail(
			mail.NewEmail(r.PostForm.Get("name"), r.PostForm.Get("email")),
			"Game DB Contact Form",
			mail.NewEmail(config.Config.AdminName.Get(), config.Config.AdminEmail.Get()),
			r.PostForm.Get("message"),
			r.PostForm.Get("message"),
		)
		client := sendgrid.NewSendClient(config.Config.SendGridAPIKey.Get())

		_, err = client.Send(message)
		if err != nil {
			log.Err(err, r)
			return ErrSomething
		}

		// Remove backup
		err = session.SetMany(r, map[string]interface{}{
			"contact-name":    "",
			"contact-email":   "",
			"contact-message": "",
		})
		log.Err(err, r)

		return nil
	}()

	// Redirect
	if err != nil {
		err = session.SetFlash(r, helpers.SessionBad, err.Error())
	} else {
		err = session.SetFlash(r, helpers.SessionGood, "Message sent!")
	}

	err = session.Save(w, r)
	log.Err(err)

	log.Err(err, r)
	http.Redirect(w, r, "/contact", http.StatusFound)
}
