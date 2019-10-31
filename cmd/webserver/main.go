package main

import (
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/Jleagle/recaptcha-go"
	. "github.com/gamedb/gamedb/cmd/webserver/middleware"
	"github.com/gamedb/gamedb/cmd/webserver/pages"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

//noinspection GoUnusedGlobalVariable
var version string

func main() {

	config.SetVersion(version)
	log.Initialise([]log.LogName{log.LogNameWebserver})

	rand.Seed(time.Now().Unix())

	//
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		log.Critical("GOOGLE_APPLICATION_CREDENTIALS not found")
		os.Exit(1)
	}

	log.Info("Starting PubSub")
	go websockets.ListenToPubSub()
	go helpers.ListenToPubSubMemcache()

	// Setup Recaptcha
	recaptcha.SetSecret(config.Config.RecaptchaPrivate.Get())

	helpers.InitSession()

	// Routes
	// Don't add compression headers here, it stops js from being able to read the response content-length
	r := chi.NewRouter()
	r.Use(MiddlewareCors())
	r.Use(middleware.RedirectSlashes)
	r.Use(MiddlewareRealIP)
	r.Use(middleware.DefaultCompress)

	// Pages
	r.Get("/", pages.HomeHandler)
	r.Get("/currency/{id}", pages.CurrencyHandler)
	r.Mount("/achievements", pages.AchievementsRouter())
	r.Mount("/admin", pages.AdminRouter())
	r.Mount("/api", pages.APIRouter())
	r.Mount("/apps", pages.AppsRouter())
	r.Mount("/badges", pages.BadgesRouter())
	r.Mount("/bundles", pages.BundlesRouter())
	r.Mount("/categories", pages.CategoriesRouter())
	r.Mount("/changes", pages.ChangesRouter())
	r.Mount("/chat", pages.ChatRouter())
	r.Mount("/chat-bot", pages.ChatBotRouter())
	r.Mount("/commits", pages.CommitsRouter())
	r.Mount("/contact", pages.ContactRouter())
	r.Mount("/coop", pages.CoopRouter())
	r.Mount("/depots", pages.DepotsRouter())
	r.Mount("/developers", pages.DevelopersRouter())
	r.Mount("/donate", pages.DonateRouter())
	r.Mount("/experience", pages.ExperienceRouter())
	r.Mount("/forgot", pages.ForgotRouter())
	r.Mount("/franchise", pages.FranchiseRouter())
	r.Mount("/genres", pages.GenresRouter())
	r.Mount("/groups", pages.GroupsRouter())
	r.Mount("/health-check", pages.HealthCheckRouter())
	r.Mount("/home", pages.HomeRouter())
	r.Mount("/info", pages.InfoRouter())
	r.Mount("/login", pages.LoginRouter())
	r.Mount("/logout", pages.LogoutRouter())
	r.Mount("/lp", pages.LandingPagesRouter())
	r.Mount("/new-releases", pages.NewReleasesRouter())
	r.Mount("/news", pages.NewsRouter())
	r.Mount("/packages", pages.PackagesRouter())
	r.Mount("/patreon", pages.PatreonRouter())
	r.Mount("/players", pages.PlayersRouter())
	r.Mount("/price-changes", pages.PriceChangeRouter())
	r.Mount("/product-keys", pages.ProductKeysRouter())
	r.Mount("/publishers", pages.PublishersRouter())
	r.Mount("/queues", pages.QueuesRouter())
	r.Mount("/sales", pages.SalesRouter())
	r.Mount("/settings", pages.SettingsRouter())
	r.Mount("/signup", pages.SignupRouter())
	r.Mount("/stats", pages.StatsRouter())
	r.Mount("/steam-api", pages.SteamAPIRouter())
	r.Mount("/tags", pages.TagsRouter())
	r.Mount("/upcoming", pages.UpcomingRouter())
	r.Mount("/websocket", pages.WebsocketsRouter())

	// Sitemaps, Google doesnt like having a sitemap in a sub directory
	r.Route("/", func(r chi.Router) {
		r.Get("/sitemap.xml", pages.SiteMapIndexHandler)
		r.Get("/sitemap-pages.xml", pages.SiteMapPagesHandler)
		r.Get("/sitemap-games-by-score.xml", pages.SiteMapGamesByScoreHandler)
		r.Get("/sitemap-games-by-players.xml", pages.SiteMapGamesByPlayersHandler)
		r.Get("/sitemap-players-by-level.xml", pages.SiteMapPlayersByLevel)
		r.Get("/sitemap-players-by-games.xml", pages.SiteMapPlayersByGamesCount)
		r.Get("/sitemap-groups.xml", pages.SiteMapGroups)
		r.Get("/sitemap-badges.xml", pages.SiteMapBadges)
	})

	// Profiling
	if config.IsLocal() {
		r.Mount("/debug", middleware.Profiler())
	}

	// Files
	r.Get("/browserconfig.xml", pages.RootFileHandler)
	r.Get("/robots.txt", pages.RootFileHandler)
	r.Get("/site.webmanifest", pages.RootFileHandler)
	r.Get("/ads.txt", pages.RootFileHandler)

	// Redirects
	r.Get("/sitemap/index.xml", pages.RedirectHandler("/sitemap.xml"))
	r.Get("/trending", pages.RedirectHandler("/apps/trending"))

	// File server
	fileServer(r, "/assets", http.Dir("./assets"))

	// 404
	r.NotFound(pages.Error404Handler)

	log.Info("Starting Webserver")
	err := http.ListenAndServe(config.ListenOn(), r)
	log.Critical(err)

	helpers.KeepAlive()
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func fileServer(r chi.Router, path string, root http.FileSystem) {

	if strings.ContainsAny(path, "{}*") {
		log.Info("Invalid URL " + path)
		return
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusFound).ServeHTTP)
		path += "/"
	}
	path += "*"

	if strings.Contains(path, "..") {
		log.Info("Invalid URL " + path)
		return
	}

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pages.SetCacheHeaders(w, time.Hour*24*28)
		fs.ServeHTTP(w, r)
	}))
}
