package web

import (
	"net/http"
)

func DiscountsHandler(w http.ResponseWriter, r *http.Request) {

	t := discountsTemplate{}
	t.Fill(w, r, "Discounts")

	returnTemplate(w, r, "discounts", t)
	return
}

type discountsTemplate struct {
	GlobalTemplate
}
