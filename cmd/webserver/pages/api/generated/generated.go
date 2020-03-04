// Package generated provides primitives to interact the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen DO NOT EDIT.
package generated

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi"
	"net/http"
	"strings"
)

// AppSchema defines model for app-schema.
type AppSchema struct {
	Categories      []int   `json:"categories"`
	Developers      []int   `json:"developers"`
	Genres          []int   `json:"genres"`
	Id              int     `json:"id"`
	MetacriticScore int32   `json:"metacritic_score"`
	Name            string  `json:"name"`
	PlayersMax      int     `json:"players_max"`
	PlayersWeekAvg  float64 `json:"players_week_avg"`
	PlayersWeekMax  int     `json:"players_week_max"`
	Prices          []struct {
		Currency        string `json:"currency"`
		DiscountPercent int32  `json:"discountPercent"`
		Final           int32  `json:"final"`
		Individual      int32  `json:"individual"`
		Initial         int32  `json:"initial"`
	} `json:"prices"`
	Publishers      []int   `json:"publishers"`
	ReleaseDate     int64   `json:"release_date"`
	ReviewsNegative int     `json:"reviews_negative"`
	ReviewsPositive int     `json:"reviews_positive"`
	ReviewsScore    float64 `json:"reviews_score"`
	Tags            []int   `json:"tags"`
}

// MessageSchema defines model for message-schema.
type MessageSchema struct {
	Message string `json:"message"`
}

// PaginationSchema defines model for pagination-schema.
type PaginationSchema struct {
	Limit        int64 `json:"limit"`
	Offset       int64 `json:"offset"`
	PagesCurrent int   `json:"pagesCurrent"`
	PagesTotal   int   `json:"pagesTotal"`
	Total        int64 `json:"total"`
}

// PlayerSchema defines model for player-schema.
type PlayerSchema struct {
	Avatar    string `json:"avatar"`
	Badges    int    `json:"badges"`
	Comments  int    `json:"comments"`
	Continent string `json:"continent"`
	Country   string `json:"country"`
	Friends   int    `json:"friends"`
	Games     int    `json:"games"`
	Groups    int    `json:"groups"`
	Id        string `json:"id"`
	Level     int    `json:"level"`
	Name      string `json:"name"`
	Playtime  int    `json:"playtime"`
	State     string `json:"state"`
	VanityUrl string `json:"vanity_url"`
}

// PriceSchema defines model for price-schema.
type PriceSchema struct {
	Currency        string `json:"currency"`
	DiscountPercent int32  `json:"discountPercent"`
	Final           int32  `json:"final"`
	Individual      int32  `json:"individual"`
	Initial         int32  `json:"initial"`
}

// LimitParam defines model for limit-param.
type LimitParam int

// AppResponse defines model for app-response.
type AppResponse AppSchema

// AppsResponse defines model for apps-response.
type AppsResponse struct {
	Apps       []AppSchema      `json:"apps"`
	Pagination PaginationSchema `json:"pagination"`
}

// MessageResponse defines model for message-response.
type MessageResponse MessageSchema

// PaginationResponse defines model for pagination-response.
type PaginationResponse PaginationSchema

// PlayerResponse defines model for player-response.
type PlayerResponse PlayerSchema

// PlayersResponse defines model for players-response.
type PlayersResponse struct {
	Pagination PaginationSchema `json:"pagination"`
	Players    []PlayerSchema   `json:"players"`
}

// GetAppsParams defines parameters for GetApps.
type GetAppsParams struct {
	Key        string    `json:"key"`
	Offset     *int      `json:"offset,omitempty"`
	Limit      *int      `json:"limit,omitempty"`
	Sort       *string   `json:"sort,omitempty"`
	Order      *string   `json:"order,omitempty"`
	Ids        *[]int    `json:"ids,omitempty"`
	Tags       *[]int    `json:"tags,omitempty"`
	Genres     *[]int    `json:"genres,omitempty"`
	Categories *[]int    `json:"categories,omitempty"`
	Developers *[]int    `json:"developers,omitempty"`
	Publishers *[]int    `json:"publishers,omitempty"`
	Platforms  *[]string `json:"platforms,omitempty"`
}

// GetAppsIdParams defines parameters for GetAppsId.
type GetAppsIdParams struct {
	Key string `json:"key"`
}

// GetPlayersParams defines parameters for GetPlayers.
type GetPlayersParams struct {
	Key       string    `json:"key"`
	Offset    *int      `json:"offset,omitempty"`
	Limit     *int      `json:"limit,omitempty"`
	Sort      *string   `json:"sort,omitempty"`
	Order     *string   `json:"order,omitempty"`
	Continent *[]string `json:"continent,omitempty"`
	Country   *[]string `json:"country,omitempty"`
}

// GetPlayersIdParams defines parameters for GetPlayersId.
type GetPlayersIdParams struct {
	Key string `json:"key"`
}

// PostPlayersIdParams defines parameters for PostPlayersId.
type PostPlayersIdParams struct {
	Key string `json:"key"`
}

type ServerInterface interface {
	// List Apps (GET /apps)
	GetApps(w http.ResponseWriter, r *http.Request)
	// Retrieve App (GET /apps/{id})
	GetAppsId(w http.ResponseWriter, r *http.Request)
	// List Players (GET /players)
	GetPlayers(w http.ResponseWriter, r *http.Request)
	// Retrieve Player (GET /players/{id})
	GetPlayersId(w http.ResponseWriter, r *http.Request)
	// Update Player (POST /players/{id})
	PostPlayersId(w http.ResponseWriter, r *http.Request)
}

// ParamsForGetApps operation parameters from context
func ParamsForGetApps(ctx context.Context) *GetAppsParams {
	return ctx.Value("GetAppsParams").(*GetAppsParams)
}

// GetApps operation middleware
func GetAppsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var err error

		ctx = context.WithValue(ctx, "key-cookie.Scopes", []string{""})

		ctx = context.WithValue(ctx, "key-header.Scopes", []string{""})

		ctx = context.WithValue(ctx, "key-query.Scopes", []string{""})

		// Parameter object where we will unmarshal all parameters from the context
		var params GetAppsParams

		// ------------- Required query parameter "key" -------------
		if paramValue := r.URL.Query().Get("key"); paramValue != "" {

		} else {
			http.Error(w, "Query argument key is required, but not found", http.StatusBadRequest)
			return
		}

		err = runtime.BindQueryParameter("form", true, true, "key", r.URL.Query(), &params.Key)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter key: %s", err), http.StatusBadRequest)
			return
		}

		// ------------- Optional query parameter "offset" -------------
		if paramValue := r.URL.Query().Get("offset"); paramValue != "" {

		}

		err = runtime.BindQueryParameter("form", true, false, "offset", r.URL.Query(), &params.Offset)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter offset: %s", err), http.StatusBadRequest)
			return
		}

		// ------------- Optional query parameter "limit" -------------
		if paramValue := r.URL.Query().Get("limit"); paramValue != "" {

		}

		err = runtime.BindQueryParameter("form", true, false, "limit", r.URL.Query(), &params.Limit)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter limit: %s", err), http.StatusBadRequest)
			return
		}

		// ------------- Optional query parameter "sort" -------------
		if paramValue := r.URL.Query().Get("sort"); paramValue != "" {

		}

		err = runtime.BindQueryParameter("form", true, false, "sort", r.URL.Query(), &params.Sort)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter sort: %s", err), http.StatusBadRequest)
			return
		}

		// ------------- Optional query parameter "order" -------------
		if paramValue := r.URL.Query().Get("order"); paramValue != "" {

		}

		err = runtime.BindQueryParameter("form", true, false, "order", r.URL.Query(), &params.Order)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter order: %s", err), http.StatusBadRequest)
			return
		}

		// ------------- Optional query parameter "ids" -------------
		if paramValue := r.URL.Query().Get("ids"); paramValue != "" {

		}

		err = runtime.BindQueryParameter("form", true, false, "ids", r.URL.Query(), &params.Ids)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter ids: %s", err), http.StatusBadRequest)
			return
		}

		// ------------- Optional query parameter "tags" -------------
		if paramValue := r.URL.Query().Get("tags"); paramValue != "" {

		}

		err = runtime.BindQueryParameter("form", true, false, "tags", r.URL.Query(), &params.Tags)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter tags: %s", err), http.StatusBadRequest)
			return
		}

		// ------------- Optional query parameter "genres" -------------
		if paramValue := r.URL.Query().Get("genres"); paramValue != "" {

		}

		err = runtime.BindQueryParameter("form", true, false, "genres", r.URL.Query(), &params.Genres)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter genres: %s", err), http.StatusBadRequest)
			return
		}

		// ------------- Optional query parameter "categories" -------------
		if paramValue := r.URL.Query().Get("categories"); paramValue != "" {

		}

		err = runtime.BindQueryParameter("form", true, false, "categories", r.URL.Query(), &params.Categories)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter categories: %s", err), http.StatusBadRequest)
			return
		}

		// ------------- Optional query parameter "developers" -------------
		if paramValue := r.URL.Query().Get("developers"); paramValue != "" {

		}

		err = runtime.BindQueryParameter("form", true, false, "developers", r.URL.Query(), &params.Developers)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter developers: %s", err), http.StatusBadRequest)
			return
		}

		// ------------- Optional query parameter "publishers" -------------
		if paramValue := r.URL.Query().Get("publishers"); paramValue != "" {

		}

		err = runtime.BindQueryParameter("form", true, false, "publishers", r.URL.Query(), &params.Publishers)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter publishers: %s", err), http.StatusBadRequest)
			return
		}

		// ------------- Optional query parameter "platforms" -------------
		if paramValue := r.URL.Query().Get("platforms"); paramValue != "" {

		}

		err = runtime.BindQueryParameter("form", true, false, "platforms", r.URL.Query(), &params.Platforms)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter platforms: %s", err), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, "GetAppsParams", &params)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ParamsForGetAppsId operation parameters from context
func ParamsForGetAppsId(ctx context.Context) *GetAppsIdParams {
	return ctx.Value("GetAppsIdParams").(*GetAppsIdParams)
}

// GetAppsId operation middleware
func GetAppsIdCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var err error

		// ------------- Path parameter "id" -------------
		var id int32

		err = runtime.BindStyledParameter("simple", false, "id", chi.URLParam(r, "id"), &id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter id: %s", err), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, "id", id)

		ctx = context.WithValue(ctx, "key-cookie.Scopes", []string{""})

		ctx = context.WithValue(ctx, "key-header.Scopes", []string{""})

		ctx = context.WithValue(ctx, "key-query.Scopes", []string{""})

		// Parameter object where we will unmarshal all parameters from the context
		var params GetAppsIdParams

		// ------------- Required query parameter "key" -------------
		if paramValue := r.URL.Query().Get("key"); paramValue != "" {

		} else {
			http.Error(w, "Query argument key is required, but not found", http.StatusBadRequest)
			return
		}

		err = runtime.BindQueryParameter("form", true, true, "key", r.URL.Query(), &params.Key)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter key: %s", err), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, "GetAppsIdParams", &params)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ParamsForGetPlayers operation parameters from context
func ParamsForGetPlayers(ctx context.Context) *GetPlayersParams {
	return ctx.Value("GetPlayersParams").(*GetPlayersParams)
}

// GetPlayers operation middleware
func GetPlayersCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var err error

		ctx = context.WithValue(ctx, "key-header.Scopes", []string{""})

		ctx = context.WithValue(ctx, "key-query.Scopes", []string{""})

		ctx = context.WithValue(ctx, "key-cookie.Scopes", []string{""})

		// Parameter object where we will unmarshal all parameters from the context
		var params GetPlayersParams

		// ------------- Required query parameter "key" -------------
		if paramValue := r.URL.Query().Get("key"); paramValue != "" {

		} else {
			http.Error(w, "Query argument key is required, but not found", http.StatusBadRequest)
			return
		}

		err = runtime.BindQueryParameter("form", true, true, "key", r.URL.Query(), &params.Key)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter key: %s", err), http.StatusBadRequest)
			return
		}

		// ------------- Optional query parameter "offset" -------------
		if paramValue := r.URL.Query().Get("offset"); paramValue != "" {

		}

		err = runtime.BindQueryParameter("form", true, false, "offset", r.URL.Query(), &params.Offset)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter offset: %s", err), http.StatusBadRequest)
			return
		}

		// ------------- Optional query parameter "limit" -------------
		if paramValue := r.URL.Query().Get("limit"); paramValue != "" {

		}

		err = runtime.BindQueryParameter("form", true, false, "limit", r.URL.Query(), &params.Limit)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter limit: %s", err), http.StatusBadRequest)
			return
		}

		// ------------- Optional query parameter "sort" -------------
		if paramValue := r.URL.Query().Get("sort"); paramValue != "" {

		}

		err = runtime.BindQueryParameter("form", true, false, "sort", r.URL.Query(), &params.Sort)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter sort: %s", err), http.StatusBadRequest)
			return
		}

		// ------------- Optional query parameter "order" -------------
		if paramValue := r.URL.Query().Get("order"); paramValue != "" {

		}

		err = runtime.BindQueryParameter("form", true, false, "order", r.URL.Query(), &params.Order)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter order: %s", err), http.StatusBadRequest)
			return
		}

		// ------------- Optional query parameter "continent" -------------
		if paramValue := r.URL.Query().Get("continent"); paramValue != "" {

		}

		err = runtime.BindQueryParameter("form", true, false, "continent", r.URL.Query(), &params.Continent)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter continent: %s", err), http.StatusBadRequest)
			return
		}

		// ------------- Optional query parameter "country" -------------
		if paramValue := r.URL.Query().Get("country"); paramValue != "" {

		}

		err = runtime.BindQueryParameter("form", true, false, "country", r.URL.Query(), &params.Country)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter country: %s", err), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, "GetPlayersParams", &params)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ParamsForGetPlayersId operation parameters from context
func ParamsForGetPlayersId(ctx context.Context) *GetPlayersIdParams {
	return ctx.Value("GetPlayersIdParams").(*GetPlayersIdParams)
}

// GetPlayersId operation middleware
func GetPlayersIdCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var err error

		// ------------- Path parameter "id" -------------
		var id int64

		err = runtime.BindStyledParameter("simple", false, "id", chi.URLParam(r, "id"), &id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter id: %s", err), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, "id", id)

		ctx = context.WithValue(ctx, "key-cookie.Scopes", []string{""})

		ctx = context.WithValue(ctx, "key-header.Scopes", []string{""})

		ctx = context.WithValue(ctx, "key-query.Scopes", []string{""})

		// Parameter object where we will unmarshal all parameters from the context
		var params GetPlayersIdParams

		// ------------- Required query parameter "key" -------------
		if paramValue := r.URL.Query().Get("key"); paramValue != "" {

		} else {
			http.Error(w, "Query argument key is required, but not found", http.StatusBadRequest)
			return
		}

		err = runtime.BindQueryParameter("form", true, true, "key", r.URL.Query(), &params.Key)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter key: %s", err), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, "GetPlayersIdParams", &params)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ParamsForPostPlayersId operation parameters from context
func ParamsForPostPlayersId(ctx context.Context) *PostPlayersIdParams {
	return ctx.Value("PostPlayersIdParams").(*PostPlayersIdParams)
}

// PostPlayersId operation middleware
func PostPlayersIdCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var err error

		// ------------- Path parameter "id" -------------
		var id int64

		err = runtime.BindStyledParameter("simple", false, "id", chi.URLParam(r, "id"), &id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter id: %s", err), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, "id", id)

		ctx = context.WithValue(ctx, "key-cookie.Scopes", []string{""})

		ctx = context.WithValue(ctx, "key-header.Scopes", []string{""})

		ctx = context.WithValue(ctx, "key-query.Scopes", []string{""})

		// Parameter object where we will unmarshal all parameters from the context
		var params PostPlayersIdParams

		headers := r.Header

		// ------------- Required header parameter "key" -------------
		if valueList, found := headers[http.CanonicalHeaderKey("key")]; found {
			var Key string
			n := len(valueList)
			if n != 1 {
				http.Error(w, fmt.Sprintf("Expected one value for key, got %d", n), http.StatusBadRequest)
				return
			}

			err = runtime.BindStyledParameter("simple", false, "key", valueList[0], &Key)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid format for parameter key: %s", err), http.StatusBadRequest)
				return
			}

			params.Key = Key

		} else {
			http.Error(w, fmt.Sprintf("Header parameter key is required, but not found", err), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, "PostPlayersIdParams", &params)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerFromMux(si, chi.NewRouter())
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r *chi.Mux) http.Handler {
	r.Group(func(r chi.Router) {
		r.Use(GetAppsCtx)
		r.Get("/apps", si.GetApps)
	})
	r.Group(func(r chi.Router) {
		r.Use(GetAppsIdCtx)
		r.Get("/apps/{id}", si.GetAppsId)
	})
	r.Group(func(r chi.Router) {
		r.Use(GetPlayersCtx)
		r.Get("/players", si.GetPlayers)
	})
	r.Group(func(r chi.Router) {
		r.Use(GetPlayersIdCtx)
		r.Get("/players/{id}", si.GetPlayersId)
	})
	r.Group(func(r chi.Router) {
		r.Use(PostPlayersIdCtx)
		r.Post("/players/{id}", si.PostPlayersId)
	})

	return r
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/+xZX2/bNhD/KgbXRyVy02FY9bRsA4ZsHWCs28sMLWCks3yNSKok5cYI/N0HUqIoyZT/",
	"pAm2An0KaN0//nj83R3zSDLBKsGBa0WSR1JRSRlokHZVIkN9YX8zS+QkIR9rkFsSEU4ZkKQRIRFR2RoY",
	"NVI5rGhdapK8nkeE0QdkNTMLs0LeriKit5XRR66hAEl2u11EJKhKcAXWOa2qC/eDWWeCa+C6/VRiRjUK",
	"Hn9QgpvffACvJKxIQr6J/dbi5quKjdFW0nrMQWUSK2OJJOSaz2hVkV1kPKineR9afIdKz8TKmFXR7BPq",
	"9ayiBXKrTSJSSVGB1Nht2f5FDUyds5UOTiol3Zp1z8sRO17SI2OO4mONEnKSLMkgYhtjGgCvv1UTAQOl",
	"aAHPf4bO8PQ5/t5IDHF4/kBCyO3FsqAFzJCvhGQNhCaokm5BvkBAjd0DCT5rRHwQT8zyYd5+VrJ1kZyc",
	"+aNtjpP/UPY6V4cS2MkYidal46MpADKqoRCyXXWbGHPc/jXNYQOlsXSuYgFcnu0N87AcA00ziRqzW5UJ",
	"aVOhSdhG7s0ViQJqTQnoDCotkRe9A71l9CHs0Al8Ari/pZti4DAX9V0J3iOv2V1Abdq4xGwEzei0aimB",
	"Z9tg7DmqTNRcL0Bm7VU4AYoVclqeKIs8xw3m9RkKqPFE6VHyd1v1Vlyw+1sdhJaGikp9V6Jan5+tEkqg",
	"Cm5zqveS67tvg5uWsEH4pG45FFTjBsKOnFQlFB6X2k/u6VzTtDhvlyPkMXdNUmuru7NRny4GFDBAuEvk",
	"4X0K3ILAfRpBHgAqgPAYqAAxpL3CPkWG7ffA7RpB5ATTYZ2estu0mqdlj1itFJwqXNEC1E/2pugJSjES",
	"fwrdXMJAJrhPR72NIGjjjLo+urE08DgKMPUNxBRSdEM1lUF6u6N5ARP5nAnG3CAQ+so18iFE3q4lEhmm",
	"1JVE4PmE2YKyqXgKKepq4tuglHlXpblMYY2D1Uojm+AOpVvO2tPbUI56e1vL8niq99mgPZ3uLHrAe6wc",
	"MB0Kbm+9eD3q/dNxIQ8CTF1dnG5hvhbFU4qiyQjIaol6+94g2YB3D9uLNdAcZDcnt8tuUL6HrY+MVvgb",
	"2MpoNJuRemLADurtLAAr4Xp2mtkzAUaxJAn5gAxoUcIPhfnhMhPM2/u1tJ9IRGzikrXWlUri2KRbfncp",
	"eIkcYmfUcBvq0ii+10DZ7OcfZ9eLG5NcIFXTN7+2hFsBpxWShLy5nF/OLWnptQUndmNt0TCyyTrL9Dc5",
	"ScgvoK8rm9/9t4flITD8WWpZQ//toaJagzSK/yznF2+vL/5OH6/mu1ceQX8/wx46Qg48aPRfMOahLHvp",
	"J5IJB0rIof1T9yqbDPWKwI3r5ZJQlZFmPiJpejJ4aInLmzvQOzH6cNN8tdset1Jh+20fdbaDU+13DdqL",
	"eRh0fi/mZdBSvpiXQa/6cl5Kqg3BH3biS5X38WbPRTp6Yryaz6feGjq5ePgUaAtAzRg1jN08GVj6Mr9b",
	"0fgR890xtrvJ/zO+M7Tcv7EH7e9V1oP09FR4p9D9A7RE2IBBuAG491I0Be+iFflaT778ejJoavfvPqMP",
	"74AXek2Sq+g8Jph06Prp53X3pJux90Ab4J5F763SKRxloFbpSyQhO1M/PwmNH+QneGjhH8+FCmC7EOoo",
	"uOG54P+EbijNPxPfvf8FDQH+q8qp9vD2hywLnBmSMiHuEUiyTKPBuOXW7RC1TE2ECuTGoe7HnCSOS5HR",
	"ci2UTt7Ov38dm4lll+7+DQAA//+HZiQYBh0AAA==",
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file.
func GetSwagger() (*openapi3.Swagger, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error loading Swagger: %s", err)
	}
	return swagger, nil
}
