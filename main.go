package main

import (
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gamedb/website/config"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/queue2"
	"github.com/gamedb/website/social"
	"github.com/gamedb/website/web"
	_ "github.com/go-sql-driver/mysql"
)

func main() {

	rand.Seed(time.Now().UnixNano())

	// Preload connections
	helpers.GetMemcache()
	_, err := db.GetMySQLClient()
	log.Critical(err)

	// Web server
	if config.Config.EnableWebserver.GetBool() {
		go func() {
			log.Info("Starting web server")
			err := web.Serve()
			log.Err(err)
		}()
	}

	// Consumers
	if config.Config.EnableConsumers.GetBool() {
		go func() {
			log.Info("Starting consumers")
			// queue.RunConsumers()

			queue2.RunConsumers()
			// err = queue2.ProduceApps([]int{440})
			// log.Err(err)
		}()
	}

	// Log number of goroutines
	go func() {
		log.Info("Logging goroutines")
		for {
			time.Sleep(time.Minute * 10)
			log.Info("Goroutines running: "+strconv.Itoa(runtime.NumGoroutine()), log.SeverityInfo, log.ServiceGoogle)
		}
	}()

	// Instagram
	if config.Config.IsProd() {
		go func() {
			log.Info("Starting Instagram")
			social.RunInstagram()
		}()
	}

	// Block forever for goroutines to run
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
