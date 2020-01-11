package log

import (
	"context"
	"fmt"
	l "log"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/logging"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/getsentry/sentry-go"
	"github.com/logrusorgru/aurora"
	"github.com/rollbar/rollbar-go"
)

const (
	LogNameChatbot   LogName = "binary-chatbot"
	LogNameConsumers LogName = "binary-consumers"
	LogNameCrons     LogName = "binary-crons"
	LogNameSteam     LogName = "binary-steam"
	LogNameWebserver LogName = "binary-webserver"
	LogNameTest      LogName = "binary-test"

	LogNameMongo         LogName = "mongo"
	LogNameInflux        LogName = "influx"
	LogNameRequests      LogName = "requests"
	LogNameSQL           LogName = "sql"
	LogNameTriggerUpdate LogName = "trigger-update"

	// Severities
	SeverityDebug    Severity = 1
	SeverityInfo     Severity = 2
	SeverityWarning  Severity = 3
	SeverityError    Severity = 4
	SeverityCritical Severity = 5

	OptionNoStack Option = iota
)

type LogName string
type Service string
type Option int
type Severity int

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

func (s Severity) toRollbar() string {

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

func (s Severity) string() string {

	switch s {
	case SeverityDebug:
		return "Debug"
	case SeverityInfo:
		return "Info"
	case SeverityWarning:
		return "Warning"
	case SeverityError:
		return "Error"
	case SeverityCritical:
		return "Critical"
	default:
		return "Error"
	}
}

type entry struct {
	request   *http.Request
	texts     []string
	error     error
	logNames  []LogName
	severity  Severity
	timestamp time.Time
	noStack   bool
}

func (e entry) toText(severity Severity) string {

	var ret []string

	// // Severity
	// ret = append(ret, e.severity.string())
	//
	// // Environment
	// if !config.IsLocal() {
	// 	ret = append(ret, config.Config.Environment.Get())
	// 	ret = append(ret, path.Base(os.Args[0]))
	// }

	// Request
	if e.request != nil {
		ret = append(ret, e.request.Method+" "+e.request.URL.Path)
	}

	// Texts
	ret = append(ret, e.texts...)

	// Error
	if e.error != nil {
		ret = append(ret, e.error.Error())
	}

	// Join
	str := strings.Join(ret, " - ")

	// Stack
	if severity > 3 && !e.noStack {
		str += "\n" + string(debug.Stack())
	}

	return str
}

var (
	googleClient *logging.Client
	logger       *l.Logger
	defaultLogs  []LogName
)

func Initialise(logs []LogName) {

	defaultLogs = logs

	if config.IsLocal() {
		logger = l.New(os.Stderr, "", 0)
	} else {
		logger = l.New(os.Stderr, "", l.LstdFlags)
	}

	var err error

	// Google
	googleClient, err = logging.NewClient(context.Background(), config.Config.GoogleProject.Get())
	if err != nil {
		fmt.Println(err)
	}

	// Rollbar
	rollbar.SetToken(config.Config.RollbarSecret.Get())
	rollbar.SetEnvironment(config.Config.Environment.Get())
	rollbar.SetServerHost("gamedb.online")
	rollbar.SetServerRoot("github.com/gamedb/gamedb")

	// Sentry
	err = sentry.Init(sentry.ClientOptions{
		Dsn:              config.Config.SentryDSN.Get(),
		AttachStacktrace: true,
		Environment:      config.Config.Environment.Get(),
		Release:          config.Config.CommitHash.Get(),
	})
	if err != nil {
		fmt.Println(err)
	}
}

func log(interfaces ...interface{}) {

	var entry = entry{
		logNames:  defaultLogs,
		severity:  SeverityError,
		timestamp: time.Now(),
	}
	var loggingServices []Service

	// Create entry
	for _, v := range interfaces {

		switch val := v.(type) {
		case nil:
			continue
		case error:
			entry.error = val
		case float32:
			entry.texts = append(entry.texts, strconv.FormatFloat(float64(val), 'f', -1, 32))
		case float64:
			entry.texts = append(entry.texts, strconv.FormatFloat(val, 'f', -1, 64))
		case []byte:
			entry.texts = append(entry.texts, string(val))
		case []string:
			entry.texts = append(entry.texts, strings.Join(val, ","))
		case time.Time:
			entry.timestamp = val
		case time.Duration:
			entry.texts = append(entry.texts, val.String())
		case net.IP:
			entry.texts = append(entry.texts, val.String())
		case *http.Request:
			entry.request = val
		case LogName:
			entry.logNames = append(entry.logNames, val)
		case Severity:
			entry.severity = val
		case Service:
			loggingServices = append(loggingServices, val)
		case Option:
			if val == OptionNoStack {
				entry.noStack = true
			}
		default:
			entry.texts = append(entry.texts, fmt.Sprint(val))
		}
	}

	if len(entry.texts) > 0 || entry.error != nil {

		// Local
		switch entry.severity {
		case SeverityCritical:
			logger.Println(aurora.Red(aurora.Bold(entry.toText(entry.severity))))
		case SeverityError:
			logger.Println(aurora.Red(entry.toText(entry.severity)))
		case SeverityWarning:
			logger.Println(aurora.Yellow(entry.toText(entry.severity)))
		case SeverityInfo:
			logger.Println(entry.toText(entry.severity))
		case SeverityDebug:
			logger.Println(aurora.Green(entry.toText(entry.severity)))
		default:
			logger.Println(entry.toText(entry.severity))
		}

		if !config.IsLocal() {

			// Google
			for _, logName := range entry.logNames {

				googleClient.Logger(string(logName)).Log(logging.Entry{
					Severity:  entry.severity.toGoole(),
					Timestamp: entry.timestamp,
					Payload:   entry.toText(entry.severity),
					Labels: map[string]string{
						"env":  config.Config.Environment.Get(),
						"hash": config.Config.CommitHash.Get(),
						"key":  config.GetSteamKeyTag(),
					},
				})
			}

			if entry.severity >= SeverityWarning {

				// Rollbar
				rollbar.Log(entry.severity.toRollbar(), entry.toText(entry.severity))

				// Sentry
				if entry.error != nil {
					sentry.CaptureException(entry.error)
				}
				sentry.Flush(time.Second * 5)
			}
		}
	}
}

func Critical(interfaces ...interface{}) {
	go log(append(interfaces, SeverityCritical)...)
}

func Err(interfaces ...interface{}) {
	go log(append(interfaces, SeverityError)...)
}

func Warning(interfaces ...interface{}) {
	go log(append(interfaces, SeverityWarning)...)
}

func Info(interfaces ...interface{}) {
	go log(append(interfaces, SeverityInfo)...)
}

func Debug(interfaces ...interface{}) {
	go log(append(interfaces, SeverityDebug)...)
}
