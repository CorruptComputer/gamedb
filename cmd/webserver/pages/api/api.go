package api

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/Jleagle/session-go/session"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/influxdata/influxdb1-client"
	"github.com/jinzhu/gorm"
)

var (
	ops = limiter.ExpirableOptions{DefaultExpirationTTL: time.Second}
	lmt = limiter.New(&ops).SetMax(1).SetBurst(2)

	// Errors
	ErrNoKey         = errors.New("no key")
	ErrOverLimit     = errors.New("over rate limit")
	ErrInvalidKey    = errors.New("invalid key")
	ErrWrongLevelKey = errors.New("wrong level key")
	ErrInvalidOffset = errors.New("invalid offset")
	ErrInvalidLimit  = errors.New("invalid limit")

	// Core params
	ParamAPIKey    = APICallParam{Name: "key", Type: "string"}
	ParamPage      = APICallParam{Name: "page", Type: "int"}
	ParamLimit     = APICallParam{Name: "limit", Type: "int"}
	ParamSortField = APICallParam{Name: "sort_field", Type: "string"}
	ParamSortOrder = APICallParam{Name: "sort_order", Type: "string"}

	// Extra params
	ParamID          = APICallParam{Name: "id", Type: "int"}
	ParamPlayers     = APICallParam{Name: "players", Type: "int"}
	ParamScore       = APICallParam{Name: "score", Type: "int"}
	ParamCategory    = APICallParam{Name: "category", Type: "int"}
	ParamReleaseDate = APICallParam{Name: "release_date", Type: "int"}
	ParamTrending    = APICallParam{Name: "trending", Type: "int"}
)

type APIRequest struct {
	request *http.Request
}

func NewAPICall(r *http.Request) (api APIRequest, err error) {

	x := APIRequest{request: r}

	key, err := x.geKey()
	if err != nil {
		return x, err
	}

	// Rate limit
	err = tollbooth.LimitByKeys(lmt, []string{key})
	if err != nil {
		// return id, offset, limit, errOverLimit // todo
	}

	// Check user ahs access to api
	level, err := sql.GetUserFromKeyCache(key)
	if err != nil {
		return x, err
	}
	if level.PatreonLevel < 3 {
		return x, ErrWrongLevelKey
	}

	if err != nil {
		return x, err
	}

	return x, nil
}

func (r APIRequest) geKey() (key string, err error) {

	key = r.request.URL.Query().Get("key")
	if key == "" {
		key, err = session.Get(r.request, helpers.SessionUserAPIKey)
		if err != nil {
			return key, err
		}
		if key == "" {
			return key, ErrNoKey
		}
	}

	if len(key) != 20 {
		return key, ErrInvalidKey
	}

	return key, err
}

func (r APIRequest) saveToInflux(success bool) (err error) {

	key, err := r.geKey()
	if err != nil {
		return err
	}

	fields := map[string]interface{}{}

	if success {
		fields["success"] = 1
	} else {
		fields["error"] = 1
	}

	_, err = helpers.InfluxWrite(helpers.InfluxRetentionPolicyAllTime, client.Point{
		Measurement: string(helpers.InfluxMeasurementAPICalls),
		Tags: map[string]string{
			"path": r.request.URL.Path,
			"key":  key,
		},
		Fields:    fields,
		Time:      time.Now(),
		Precision: "u",
	})

	return err
}

func (r APIRequest) getQueryString(key string) string {
	return r.request.URL.Query().Get(key)
}

func (r APIRequest) getQueryInt(key string) (int64, error) {
	return strconv.ParseInt(r.request.URL.Query().Get(key), 10, 64)
}

func (r APIRequest) SetSQLLimitOffset(db *gorm.DB) (*gorm.DB, error) {

	var err error

	// Limit
	limit, err := r.getQueryInt("limit")
	if err != nil {
		return db, err
	}
	if limit <= 0 {
		return db, errors.New("invalid limit")
	}

	db = db.Limit(limit)

	// Offset
	offset, err := r.getQueryInt("offset")
	if err != nil {
		return db, err
	}
	if limit <= 0 {
		return db, errors.New("invalid offset")
	}

	db = db.Offset(offset)

	return db, db.Error
}

func (r APIRequest) setSQLOrder(db *gorm.DB, allowed []string) (*gorm.DB, error) {

	field := r.getQueryString(ParamSortField.Name)
	if !helpers.SliceHasString(allowed, field) {
		return db, errors.New("invalid limit")
	}

	switch r.getQueryString(ParamSortField.Name) {
	case "ascending", "asc", "1":
		db = db.Order(field + " ASC")
	case "descending", "desc", "0", "-1":
		db = db.Order(field + " DESC")
	default:
		db = db.Order("id asc")
	}

	return db, db.Error
}

//
type APICallParam struct {
	Name    string
	Type    string
	Default string
}

func (p APICallParam) InputType() string {
	if helpers.SliceHasString([]string{"int", "uint"}, p.Type) {
		return "number"
	}
	return "text"
}

//
type APICall struct {
	Title   string
	Version int
	Path    string
	Params  []APICallParam
	Handler http.HandlerFunc
}

func (c APICall) Hashtag() string {
	return regexp.MustCompile("[^a-zA-Z0-9]+").ReplaceAllString(c.Title, "")
}

func (c APICall) GetPath() string {
	return "/" + c.VersionString() + "/" + c.Path
}

func (c APICall) VersionString() string {
	if c.Version == 0 {
		c.Version = 1
	}
	return "v" + strconv.Itoa(c.Version)
}
