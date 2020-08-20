package log

import (
	"context"
	"fmt"

	"cloud.google.com/go/logging"
	"github.com/gamedb/gamedb/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogName string

const (
	// Binaries
	LogNameAPI       LogName = "binary-api"
	LogNameBackend   LogName = "binary-backend"
	LogNameChatbot   LogName = "binary-chatbot"
	LogNameConsumers LogName = "binary-consumers"
	LogNameCrons     LogName = "binary-crons"
	LogNameSteam     LogName = "binary-steam"
	LogNameFrontend  LogName = "binary-frontend"
	LogNameTest      LogName = "binary-test"
	LogNameScaler    LogName = "binary-scaler"
)

func InitZap(logName LogName) {

	var logger *zap.Logger

	if config.IsLocal() {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}

	if config.IsLocal() {

		// grpcZap.ReplaceGrpcLoggerV2(logger)

	} else {

		googleClient, err := logging.NewClient(context.Background(), config.Config.GoogleProject.Get())
		if err != nil {
			fmt.Println(err)
		}

		logger = logger.WithOptions(zap.Hooks(func(e zapcore.Entry) error {

			if googleClient != nil {

				var level logging.Severity
				var message = e.Message

				if e.Level > zapcore.InfoLevel {
					message += "\n" + e.Stack
				}

				switch e.Level {
				case zapcore.DebugLevel:
					level = logging.Debug
				case zapcore.InfoLevel:
					level = logging.Info
				case zapcore.WarnLevel:
					level = logging.Warning
				case zapcore.ErrorLevel:
					level = logging.Error
				case zapcore.DPanicLevel:
					level = logging.Critical
				case zapcore.PanicLevel:
					level = logging.Critical
				default:
					level = logging.Debug
				}

				googleClient.Logger(string(logName)).Log(logging.Entry{
					Timestamp: e.Time,
					Severity:  level,
					Payload:   message,
					Labels: map[string]string{
						"env":     config.Config.Environment.Get(),
						"commits": config.Config.Commits.Get(),
						"hash":    config.Config.CommitHash.Get(),
						"key":     config.Config.SteamAPIKey.Get(),
						"ip":      config.Config.IP.Get(),
					},
				})
			}
			return nil
		}))
	}

	zap.ReplaceGlobals(logger)
}
