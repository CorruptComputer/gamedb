package web

import (
	"io/ioutil"
	"net/http"
	"os"
)

func InfoHandler(w http.ResponseWriter, r *http.Request) {

	template := staticTemplate{}
	template.Fill(r, "Info")

	returnTemplate(w, r, "info", template)
}

func DonateHandler(w http.ResponseWriter, r *http.Request) {

	template := staticTemplate{}
	template.Fill(r, "Donate")

	returnTemplate(w, r, "donate", template)
}

func Error404Handler(w http.ResponseWriter, r *http.Request) {

	returnErrorTemplate(w, r, 404, "Page not found")
}

func RootFileHandler(w http.ResponseWriter, r *http.Request) {

	data, err := ioutil.ReadFile(os.Getenv("STEAM_PATH") + r.URL.Path)
	if err != nil {
		returnErrorTemplate(w, r, 404, "Unable to read file.")
		return
	}

	w.Write(data)

}

type staticTemplate struct {
	GlobalTemplate
}
