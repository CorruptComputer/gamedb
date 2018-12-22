package log

import (
	"context"
	"fmt"
	logg "log"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/logging"
	"github.com/rollbar/rollbar-go"
	"github.com/spf13/viper"
)

//noinspection GoUnusedConst
const (
	// Environments
	EnvProd  Environment = "production"
	EnvLocal Environment = "local"

	// Log names
	LogNameConsumers LogName = "consumers"
	LogNameCron      LogName = "crons"
	LogNameSteam     LogName = "steam-calls"
	LogNameGameDB    LogName = "gamedb" // Default
	LogNameRequests  LogName = "requests"
	LogNameDatastore LogName = "datastore"

	// Severities
	SeverityDebug    Severity = "debug"
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error" // Default
	SeverityCritical Severity = "critical"

	// Services
	ServiceGoogle  Service = "google"  // Default
	ServiceRollbar Service = "rollbar" //
	ServiceLocal   Service = "local"   // Default

	// Options
	//OptionStack Option = iota
)

type LogName string
type Environment string
type Service string
type Option int
type Severity string

func (s Severity) toGoole() (severity logging.Severity) {

	switch s {
	case SeverityDebug:
		return logging.Debug
	case SeverityInfo:
		return logging.Info
	case SeverityWarning:
		return logging.Warning
	case SeverityError:
		return logging.Error
	case SeverityCritical:
		return logging.Critical
	default:
		return logging.Error
	}
}

func (s Severity) toRollbar() (severity string) {

	switch s {
	case SeverityDebug:
		return rollbar.DEBUG
	case SeverityInfo:
		return rollbar.INFO
	case SeverityWarning:
		return rollbar.WARN
	case SeverityError:
		return rollbar.ERR
	case SeverityCritical:
		return rollbar.CRIT
	default:
		return rollbar.ERR
	}
}

type entry struct {
	request     *http.Request
	text        string
	error       string
	logName     LogName
	severity    Severity
	timestamp   time.Time
	environment Environment
}

func (e entry) toText(includeStack bool) string {

	var ret []string

	ret = append(ret, strings.ToUpper(string(e.severity)))

	if e.request != nil {
		ret = append(ret, e.request.Method+" "+e.request.URL.Path)
	}

	if e.text != "" {
		ret = append(ret, e.text)
	}

	if e.error != "" {
		ret = append(ret, e.error)
	}

	str := strings.Join(ret, " - ")

	if includeStack {
		str += "\n" + string(debug.Stack())
	}

	return str
}

var (
	env          Environment
	googleClient *logging.Client
	logger       = logg.New(os.Stderr, "", logg.Ltime)
)

// Called from main
func Init() {

	envString := viper.GetString("ENV")
	env = Environment(envString)

	// Setup Google
	var err error
	googleClient, err = logging.NewClient(context.Background(), viper.GetString("GOOGLE_PROJECT"))
	if err != nil {
		fmt.Println(err)
	}

	// Setup Roolbar
	rollbar.SetToken(viper.GetString("ROLLBAR_PRIVATE"))
	rollbar.SetEnvironment(envString)                  // defaults to "development"
	rollbar.SetCodeVersion("master")                   // optional Git hash/branch/tag (required for GitHub integration)
	rollbar.SetServerRoot("github.com/gamedb/website") // path of project (required for GitHub integration and non-project stacktrace collapsing)
}

func log(interfaces ...interface{}) {

	var entry = entry{
		logName:   LogNameGameDB,
		severity:  SeverityError,
		timestamp: time.Now(),
	}
	var loggingServices []Service

	// Create entry
	for _, v := range interfaces {

		switch val := v.(type) {
		case nil:
			continue
		case int:
			entry.text = strconv.Itoa(val)
		case string:
			entry.text = val
		case *http.Request:
			entry.request = val
		case error:
			if val != nil {
				entry.error = val.Error()
			}
		case LogName:
			entry.logName = val
		case Severity:
			entry.severity = val
		case Environment:
			entry.environment = val
		case time.Time:
			entry.timestamp = val
		case Service:
			loggingServices = append(loggingServices, val)
		case Option:
		default:
			Err("Invalid value given to Err")
		}
	}

	if entry.text == "" && entry.error == "" && entry.request == nil {
		return
	}

	if entry.environment != "" && entry.environment != env {
		return
	}

	if len(loggingServices) == 0 {
		loggingServices = append(loggingServices, ServiceGoogle, ServiceLocal)
	}

	// Send entry
	for _, v := range loggingServices {

		// Local
		if v == ServiceLocal {
			logger.Println(entry.toText(false))
		}

		// Google
		if v == ServiceGoogle {
			googleClient.Logger(string(env) + "-" + string(entry.logName)).Log(logging.Entry{
				Severity:  entry.severity.toGoole(),
				Timestamp: entry.timestamp,
				Payload:   entry.toText(true),
			})
		}

		// Rollbar
		if v == ServiceRollbar {
			rollbar.Log(entry.severity.toRollbar(), interfaces...)
		}
	}
}

func Err(interfaces ...interface{}) {
	log(append(interfaces, SeverityError)...)
}

func Info(interfaces ...interface{}) {
	log(append(interfaces, SeverityInfo)...)
}

func Debug(interfaces ...interface{}) {
	log(append(interfaces, SeverityDebug)...)
}

func Warning(interfaces ...interface{}) {
	log(append(interfaces, SeverityWarning)...)
}

func Critical(interfaces ...interface{}) {
	log(append(interfaces, SeverityCritical)...)
}
