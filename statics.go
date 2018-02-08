package main

import "net/http"

func creditsHandler(w http.ResponseWriter, r *http.Request) {
	returnTemplate(w, "credits", nil)
}

func donateHandler(w http.ResponseWriter, r *http.Request) {
	returnTemplate(w, "donate", nil)
}

func faqsHandler(w http.ResponseWriter, r *http.Request) {
	returnTemplate(w, "faqs", nil)
}
