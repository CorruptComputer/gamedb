package crons

import (
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
)

type AppPlayers struct {
}

func (c AppPlayers) ID() CronEnum {
	return CronAppPlayers
}

func (c AppPlayers) Name() string {
	return "Check apps for players"
}

func (c AppPlayers) Config() sql.ConfigType {
	return sql.ConfAddedAllAppPlayers
}

func (c AppPlayers) Work() {

	log.Info("Queueing apps for player checks")

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Critical(err)
		return
	}

	gorm = gorm.Select([]string{"id"})
	gorm = gorm.Order("id ASC")
	gorm = gorm.Model(&[]sql.App{})

	var appIDs []int
	gorm = gorm.Pluck("id", &appIDs)
	if gorm.Error != nil {
		log.Critical(gorm.Error)
	}

	log.Info("Found " + strconv.Itoa(len(appIDs)) + " apps")

	// Chunk appIDs
	var chunks [][]int
	for i := 0; i < len(appIDs); i += 10 {
		end := i + 10

		if end > len(appIDs) {
			end = len(appIDs)
		}

		chunks = append(chunks, appIDs[i:end])
	}

	log.Info("Chunking")

	for _, chunk := range chunks {

		err = queue.ProduceAppPlayers(chunk)
		log.Err(err)
	}

	log.Info("Finished chunking")

	//
	err = sql.SetConfig(sql.ConfAddedAllAppPlayers, strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	page := websockets.GetPage(websockets.PageAdmin)
	page.Send(websockets.AdminPayload{Message: string(sql.ConfAddedAllAppPlayers) + " complete"})

	cronLogInfo("App players cron complete")
}

type ClearUpcomingCache struct {
}

func (c ClearUpcomingCache) ID() CronEnum {
	return CronClearUpcomingCache
}

func (c ClearUpcomingCache) Name() string {
	return "Clear upcoming apps cache"
}

func (c ClearUpcomingCache) Config() sql.ConfigType {
	return sql.ConfClearUpcomingCache
}

func (c ClearUpcomingCache) Work() {

	var mc = helpers.GetMemcache()
	var err error

	err = mc.Delete(helpers.MemcacheUpcomingAppsCount.Key)
	err = helpers.IgnoreErrors(err, helpers.ErrCacheMiss)
	log.Err(err)

	err = mc.Delete(helpers.MemcacheUpcomingPackagesCount.Key)
	err = helpers.IgnoreErrors(err, helpers.ErrCacheMiss)
	log.Err(err)
}
