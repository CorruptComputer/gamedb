package web

import (
	"net/http"
)

func StatsHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	template := statsTemplate{}
	template.Fill(r, "Stats")

	returnTemplate(w, r, "stats", template)
}

type statsTemplate struct {
	GlobalTemplate
}
