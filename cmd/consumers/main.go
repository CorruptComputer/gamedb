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

	// Load pubsub
	log.Info("Listening to PubSub for memcache")
	go helpers.ListenToPubSubMemcache()

	// Load PPROF
	if config.IsLocal() {
		log.Info("Starting consumers profiling")
		go func() {
			log.Err(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	// Load consumers
	log.Info("Starting consumers")
	for queueName, q := range queue.QueueRegister {
		q.Name = queueName
		go q.ConsumeMessages()
	}

	// Load Steam PICS checker
	go queue.InitSteam()

	//
	helpers.KeepAlive()
}
