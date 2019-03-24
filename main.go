package main

import (
	"math/rand"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gamedb/website/config"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/queue"
	"github.com/gamedb/website/social"
	"github.com/gamedb/website/web"
	_ "github.com/go-sql-driver/mysql"
	"github.com/robfig/cron"
)

//noinspection GoUnusedGlobalVariable
var version string

func main() {

	config.Config.CommitHash = version

	rand.Seed(time.Now().UnixNano())

	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		log.Err("GOOGLE_APPLICATION_CREDENTIALS not found")
		os.Exit(1)
	}

	// Preload connections
	helpers.GetMemcache()
	_, err := db.GetMySQLClient()
	log.Critical(err)

	// Web server
	if config.Config.EnableWebserver.GetBool() {
		go func() {
			log.Info("Starting web server")
			err := web.Serve()
			log.Critical(err)
		}()
	}

	// Consumers
	if config.Config.EnableConsumers.GetBool() {
		go func() {
			log.Info("Starting consumers")
			queue.RunConsumers()
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

	// Prod crons
	if config.Config.IsProd() {

		c := cron.New()

		// Daily
		err = c.AddFunc("0 0 0 * * *", web.ClearUpcomingCache)
		log.Critical(err)

		err = c.AddFunc("0 0 0 * * *", web.CronRanks)
		log.Critical(err)

		err = c.AddFunc("0 0 1 * * *", web.CronGenres)
		log.Critical(err)

		err = c.AddFunc("0 0 2 * * *", web.CronTags)
		log.Critical(err)

		err = c.AddFunc("0 0 3 * * *", web.CronPublishers)
		log.Critical(err)

		err = c.AddFunc("0 0 4 * * *", web.CronDevelopers)
		log.Critical(err)

		err = c.AddFunc("0 0 5 * * *", web.CronDonations)
		log.Critical(err)

		err = c.AddFunc("0 0 12 * * *", social.UploadInstagram)
		log.Critical(err)

		// Every 3 hours
		err = c.AddFunc("0 0 */3 * * *", checkForPlayers)
		log.Critical(err)

		c.Start()
	}

	// Block forever for goroutines to run
	x := make(chan os.Signal)
	signal.Notify(x, syscall.SIGTERM, os.Interrupt, os.Kill)

	wg := &sync.WaitGroup{} // Must be pointer
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		for range x {

			//noinspection GoDeferInLoop
			defer wg.Done()

			sql, err := db.GetMySQLClient()
			if err != nil {
				log.Err(err)
				return
			}

			err = sql.Close()
			log.Err(err)

			return
		}
	}(wg)

	wg.Wait()
}

// This is here because you can't queue things from db package
func checkForPlayers() {

	log.Info("Queueing apps for player checks")

	gorm, err := db.GetMySQLClient()
	if err != nil {
		log.Critical(err)
		return
	}

	gorm = gorm.Select([]string{"id"})
	gorm = gorm.Order("id ASC")
	gorm = gorm.Model(&[]db.App{})

	var appIDs []int
	gorm = gorm.Pluck("id", &appIDs)
	if gorm.Error != nil {
		log.Critical(gorm.Error)
	}

	appIDs = append(appIDs, 0) // Steam client

	// Chunk appIDs
	var chunks [][]int
	for i := 0; i < len(appIDs); i += 10 {
		end := i + 10

		if end > len(appIDs) {
			end = len(appIDs)
		}

		chunks = append(chunks, appIDs[i:end])
	}

	for _, chunk := range chunks {

		err = queue.ProduceAppPlayers(chunk)
		log.Err(err)
	}
}
