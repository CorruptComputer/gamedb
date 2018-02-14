package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/99designs/basicauth-go"
	"github.com/Jleagle/go-helpers/logger"
	"github.com/go-chi/chi"
	"github.com/steam-authority/steam-authority/pics"
	"github.com/steam-authority/steam-authority/queue"
	"github.com/steam-authority/steam-authority/websockets"
)

func main() {

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", os.Getenv("STEAM_GOOGLE_APPLICATION_CREDENTIALS"))

	arguments := os.Args[1:]

	if len(arguments) > 0 {

		switch arguments[0] {
		case "consumers":
			queue.RunConsumers()
		case "pics":
			pics.RunPICS()
		default:
			fmt.Println("No such CLI command")
		}

		os.Exit(0)
	}

	logger.SetRollbarKey(os.Getenv("STEAM_ROLLBAR_PRIVATE"))

	r := chi.NewRouter()

	// Apps
	r.Get("/apps", appsHandler)
	r.Get("/apps/{id}", appHandler)
	r.Get("/apps/{id}/{slug}", appHandler)

	// Packages
	r.Get("/packages", packagesHandler)
	r.Get("/packages/{id}", packageHandler)

	// Players
	r.Post("/players", playerIDHandler)
	r.Get("/players", playersHandler)
	r.Get("/players/{id:[a-z]+}", playersHandler)
	r.Get("/players/{id:[0-9]+}", playerHandler)
	r.Get("/players/{id:[0-9]+}/{slug}", playerHandler)

	// Changes
	r.Get("/changes", changesHandler)
	r.Get("/changes/{id}", changeHandler)

	// Experience
	r.Get("/experience", experienceHandler)
	r.Get("/experience/{id}", experienceHandler)

	// Contact
	r.Get("/contact", contactHandler)
	r.Post("/contact", postContactHandler)

	// Static pages
	r.Get("/donate", donateHandler)
	r.Get("/faqs", faqsHandler)
	r.Get("/credits", creditsHandler)

	// Chat
	r.Get("/chat", chatHandler)
	r.Get("/chat/{id}", chatHandler)

	// Other
	r.Get("/", homeHandler)
	r.Get("/websocket", websockets.Handler)
	r.Get("/changelog", changelogHandler)
	r.Get("/tags", tagsHandler)
	r.Get("/news", newsHandler)

	// Admin
	r.Mount("/admin", adminRouter())

	workDir, _ := os.Getwd()
	filesDir := filepath.Join(workDir, "assets")
	fileServer(r, "/assets", http.Dir(filesDir))

	http.ListenAndServe(":8085", r)
}

func adminRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(basicauth.New("Steam", map[string][]string{
		os.Getenv("STEAM_AUTH_USER"): {os.Getenv("STEAM_AUTH_PASS")},
	}))
	r.Get("/rerank", adminReRankHandler)
	r.Get("/fill-apps", adminUpdateAllAppsHandler)
	return r
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func fileServer(r chi.Router, path string, root http.FileSystem) {

	if strings.ContainsAny(path, "{}*") {
		logger.Info("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
