package main

import (
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {

	http.Redirect(w, r, "/players", 302)
	return

	template := homeTemplate{}
	returnTemplate(w, "home", template)
}

type homeTemplate struct {
}
