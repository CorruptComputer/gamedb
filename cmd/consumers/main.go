package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
)

func main() {

	// debug.SetGCPercent(1)

	if config.IsLocal() {
		log.Info("Starting consumers profiling")
		go func() {
			log.Err(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	log.Info("Starting consumers")

	for queueName, q := range queue.QueueRegister {
		q.Name = queueName
		go q.ConsumeMessages()
	}

	helpers.KeepAlive()
}
