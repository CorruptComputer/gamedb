package main

import (
	"net/http"
	"os"

	"github.com/Jleagle/go-helpers/logger"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func contactHandler(w http.ResponseWriter, r *http.Request) {

	template:= contactTemplate{}
	template.Fill(r)

	returnTemplate(w, r, "contact", template)
}

type contactTemplate struct {
	GlobalTemplate
}

func postContactHandler(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		logger.Error(err)
	}

	// Validation
	if r.PostForm.Get("name") == "" {
		returnErrorTemplate(w, r, 500, "Please fill in the whole form.")
		return
	}
	if r.PostForm.Get("email") == "" {
		returnErrorTemplate(w, r, 500, "Please fill in the whole form.")
		return
	}
	if r.PostForm.Get("message") == "" {
		returnErrorTemplate(w, r, 500, "Please fill in the whole form.")
		return
	}

	to := mail.NewEmail("James Eagle", "jimeagle@gmail.com")
	from := mail.NewEmail(r.PostForm.Get("name"), r.PostForm.Get("email"))

	message := mail.NewSingleEmail(from, "Steam Authority Contact Form", to, r.PostForm.Get("message"), r.PostForm.Get("message"))
	client := sendgrid.NewSendClient(os.Getenv("STEAM_SENDGRID"))

	_, err := client.Send(message)
	if err != nil {
		returnErrorTemplate(w, r, 500, err.Error())
		return
	} else {
		http.Redirect(w, r, "/contact?success", 302)
		return
	}
}
