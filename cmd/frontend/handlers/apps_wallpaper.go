package handlers

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
)

func wallpaperRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", wallpaperHandler)
	return r
}

func wallpaperHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := wallpaperTemplate{}
	t.fill(w, r, "wallpaper", "Stats", "Some interesting Steam Store stats.")

	apps, err := mongo.GetApps(0, 112, bson.D{{"player_peak_week", -1}}, nil, nil)
	if err != nil {
		log.ErrS(err)
	}

	for _, v := range apps {
		if v.ID != 480 && v.ID != 218 {
			t.Apps = append(t.Apps, v.ID)
		}
	}

	returnTemplate(w, r, t)
}

type wallpaperTemplate struct {
	globalTemplate
	Apps []int
}
