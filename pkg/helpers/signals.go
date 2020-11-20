package helpers

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gamedb/gamedb/pkg/log"
	"go.uber.org/zap"
)

func KeepAlive(callbacks ...func()) {

	var signals = []os.Signal{
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGQUIT,
		syscall.SIGINT,
		syscall.SIGKILL,
	}

	signalsChan := make(chan os.Signal, len(signals))
	signal.Notify(signalsChan, signals...)

	s := <-signalsChan // Blocks

	for _, callback := range callbacks {
		callback()
	}

	log.Info("Shutting down", zap.String("signal", s.String()))
}
