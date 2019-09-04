package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Jleagle/session-go/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/cors"
	"github.com/justinas/nosurf"
)

func MiddlewareCSRF(h http.Handler) http.Handler {
	return nosurf.New(h)
}

// todo, check this is alright
func MiddlewareCors() func(next http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins: []string{config.Config.GameDBDomain.Get()}, // Use this to allow specific origin hosts
		AllowedMethods: []string{"GET", "POST"},
	}).Handler
}

func MiddlewareRealIP(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		rip := r.Header.Get(http.CanonicalHeaderKey("X-Real-IP"))
		if rip != "" {
			r.RemoteAddr = rip
		}
		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func MiddlewareTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		r.Header.Set("start-time", strconv.FormatInt(time.Now().UnixNano(), 10))

		next.ServeHTTP(w, r)
	})
}

func MiddlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.IsLocal() {
			log.Info(log.LogNameRequests, r.Method+" "+r.URL.String())
		}
		next.ServeHTTP(w, r)
	})
}

func MiddlewareAuthCheck() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			loggedIn, err := helpers.IsLoggedIn(r)
			log.Err(err)

			if loggedIn && err == nil {
				next.ServeHTTP(w, r)
				return
			}

			err = session.SetFlash(r, helpers.SessionBad, "Please login")
			log.Err(err, r)

			http.Redirect(w, r, "/login", http.StatusFound)
			return
		})
	}
}

func MiddlewareAdminCheck(handler http.HandlerFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if helpers.IsAdmin(r) {
				next.ServeHTTP(w, r)
				return
			}

			handler(w, r)
		})
	}
}
