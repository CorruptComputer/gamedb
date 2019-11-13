package pages

import (
	"net/http"

	"github.com/Jleagle/session-go/session"
	"github.com/gamedb/gamedb/cmd/webserver/pages/api"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi"
)

var endpoints = []api.APICall{
	{
		Title: "App - Players",
		Path:  "app-players",
		Params: []api.APICallParam{
			api.ParamAPIKey,
			api.ParamPage,
			api.ParamLimit,
			api.ParamSortField,
			api.ParamSortOrder,
			{Name: "id", Type: "int"},
		},
	},
	{
		Title: "App - Price Changes",
		Path:  "app-prices",
		Params: []api.APICallParam{
			api.ParamAPIKey,
			api.ParamPage,
			api.ParamLimit,
			api.ParamSortField,
			api.ParamSortOrder,
			{Name: "id", Type: "int"},
		},
	},
	{
		Title: "Apps",
		Path:  "apps",
		Params: []api.APICallParam{
			api.ParamAPIKey,
			api.ParamPage,
			api.ParamLimit,
			api.ParamSortField,
			api.ParamSortOrder,
			{Name: "id", Type: "int"},
			{Name: "category", Type: "int"},
			{Name: "tag", Type: "int"},
			{Name: "genre", Type: "int"},
			{Name: "min_players", Type: "int", Comment: "Max over last 7 days"},
			{Name: "max_players", Type: "int", Comment: "Max over last 7 days"},
			{Name: "min_release_date", Type: "timestamp", Comment: "Only includes apps with valid release dates"},
			{Name: "max_release_date", Type: "timestamp", Comment: "Only includes apps with valid release dates"},
			{Name: "min_score", Type: "int"},
			{Name: "max_score", Type: "int"},
			{Name: "min_trending", Type: "int"},
			{Name: "max_trending", Type: "int"},
			{Name: "min_avg_players", Type: "int", Comment: "Avg over last 7 days"},
			{Name: "max_avg_players", Type: "int", Comment: "Avg over last 7 days"},
		},
		Handler: APIEndpointHandler(api.AppsHandler),
	},
	{
		Title: "Articles",
		Path:  "articles",
		Params: []api.APICallParam{
			api.ParamAPIKey,
			api.ParamPage,
			api.ParamLimit,
			api.ParamSortField,
			api.ParamSortOrder,
		},
	},
	{
		Title: "Bundles",
		Path:  "bundles",
		Params: []api.APICallParam{
			api.ParamAPIKey,
			api.ParamPage,
			api.ParamLimit,
			api.ParamSortField,
			api.ParamSortOrder,
		},
	},
	{
		Title: "Changes",
		Path:  "changes",
		Params: []api.APICallParam{
			api.ParamAPIKey,
			api.ParamPage,
			api.ParamLimit,
			api.ParamSortField,
			api.ParamSortOrder,
		},
	},
	{
		Title: "Groups",
		Path:  "groups",
		Params: []api.APICallParam{
			api.ParamAPIKey,
			api.ParamPage,
			api.ParamLimit,
			api.ParamSortField,
			api.ParamSortOrder,
		},
	},
	{
		Title: "Packages",
		Path:  "packages",
		Params: []api.APICallParam{
			api.ParamAPIKey,
			api.ParamPage,
			api.ParamLimit,
			api.ParamSortField,
			api.ParamSortOrder,
		},
	},
	{
		Title: "Player - Badges",
		Path:  "player-badges",
		Params: []api.APICallParam{
			api.ParamAPIKey,
			api.ParamPage,
			api.ParamLimit,
			api.ParamSortField,
			api.ParamSortOrder,
			{Name: "id", Type: "int"},
		},
	},
	{
		Title: "Player - Games",
		Path:  "player-apps",
		Params: []api.APICallParam{
			api.ParamAPIKey,
			api.ParamPage,
			api.ParamLimit,
			api.ParamSortField,
			api.ParamSortOrder,
			{Name: "id", Type: "int"},
		},
	},
	{
		Title: "Player - History",
		Path:  "player-history",
		Params: []api.APICallParam{
			api.ParamAPIKey,
			api.ParamPage,
			api.ParamLimit,
			api.ParamSortField,
			api.ParamSortOrder,
			{Name: "id", Type: "int"},
		},
	},
	{
		Title:  "Player - Update",
		Path:   "player-update",
		Params: []api.APICallParam{},
	},
	{
		Title: "Players",
		Path:  "players",
		Params: []api.APICallParam{
			api.ParamAPIKey,
			api.ParamPage,
			api.ParamLimit,
			api.ParamSortField,
			api.ParamSortOrder,
		},
	},
	{
		Title: "Stats - Categories",
		Path:  "steam-stats",
		Params: []api.APICallParam{
			api.ParamAPIKey,
		},
	},
	{
		Title: "Stats - Genres",
		Path:  "steam-stats",
		Params: []api.APICallParam{
			api.ParamAPIKey,
		},
	},
	{
		Title: "Stats - Publishers",
		Path:  "steam-stats",
		Params: []api.APICallParam{
			api.ParamAPIKey,
		},
	},
	{
		Title: "Stats - Steam",
		Path:  "steam-stats",
		Params: []api.APICallParam{
			api.ParamAPIKey,
		},
	},
	{
		Title: "Stats - Tags",
		Path:  "steam-stats",
		Params: []api.APICallParam{
			api.ParamAPIKey,
		},
	},
}

func APIRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", apiHandler)
	r.Get("/swagger.json", APISwaggerHandler)

	for _, v := range endpoints {
		if v.Handler != nil {
			r.Get(v.GetPath(), v.Handler)
		}
	}

	return r
}

func apiHandler(w http.ResponseWriter, r *http.Request) {

	t := apiTemplate{}
	t.fill(w, r, "API", "A list of API endpoints to access Steam data & Game DB data")
	t.Calls = endpoints
	t.Key, _ = session.Get(r, helpers.SessionUserAPIKey)
	t.Base = config.Config.GameDBDomain.Get()

	returnTemplate(w, r, "api", t)
}

type apiTemplate struct {
	GlobalTemplate
	Calls []api.APICall
	Key   string
	Base  string
}

func APISwaggerHandler(w http.ResponseWriter, r *http.Request) {

	swagger := openapi3.Swagger{
		OpenAPI: "3.0",
		Info: openapi3.Info{
			Title: "Steam DB API",
			Contact: &openapi3.Contact{
				URL: "https://gamedb.online/contact",
			},
		},
		Servers: []*openapi3.Server{
			{URL: "https://gamedb.online/api"},
		},
		Paths: openapi3.Paths{
			"/prefix/{pathArg}/suffix": &openapi3.PathItem{
				Post: &openapi3.Operation{
					// Parameters: openapi3.Parameters{
					// 	{
					// 		Value: &openapi3.Parameter{
					// 			In:     "query",
					// 			Name:   "pathArg",
					// 			Schema: openapi3.NewStringSchema().WithMaxLength(2).NewRef(),
					// 		},
					// 	},
					// 	{
					// 		Value: &openapi3.Parameter{
					// 			In:     "query",
					// 			Name:   "queryArg",
					// 			Schema: openapi3.NewStringSchema().WithMaxLength(2).NewRef(),
					// 		},
					// 	},
					// },
				},
			},
		},
	}

	b, err := swagger.MarshalJSON()
	log.Err(err)

	_, err = w.Write(b)
	log.Err(err)
}

func APIEndpointHandler(callback func(api.APIRequest) (ret interface{}, err error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		call, err := api.NewAPICall(r)
		if err != nil {

			returnJSON(w, r, APIEndpointResponse{Error: err.Error()})

			err = call.SaveToInflux(false, err)
			log.Err(err, r)

			return
		}

		resp, err := callback(call)
		if err != nil {

			returnJSON(w, r, APIEndpointResponse{Error: err.Error()})

			err = call.SaveToInflux(false, err)
			log.Err(err, r)

			return
		}

		returnJSON(w, r, APIEndpointResponse{Data: resp})

		err = call.SaveToInflux(true, nil)
		log.Err(err, r)
	}
}

type APIEndpointResponse struct {
	Error string      `json:"error"`
	Data  interface{} `json:"data"`
}
