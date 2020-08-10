package tasks

import (
	"strconv"
	"strings"
	"sync"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"go.mongodb.org/mongo-driver/bson"
)

type StatsDevelopers struct {
	BaseTask
}

func (c StatsDevelopers) ID() string {
	return "update-developer-stats"
}

func (c StatsDevelopers) Name() string {
	return "Update developers"
}

func (c StatsDevelopers) Group() TaskGroup {
	return ""
}

func (c StatsDevelopers) Cron() TaskTime {
	return CronTimeDevelopers
}

func (c StatsDevelopers) work() (err error) {

	// Get current developers, to delete old ones
	allDevelopers, err := mysql.GetAllDevelopers([]string{"id", "name"})
	if err != nil {
		return err
	}

	developersToDelete := map[int]bool{}
	for _, v := range allDevelopers {
		developersToDelete[v.ID] = true
	}

	var developersNameMap = map[int]string{}
	for _, v := range allDevelopers {
		developersNameMap[v.ID] = strings.TrimSpace(v.GetName())
	}

	// Get apps from mysql
	appsWithDevelopers, err := mongo.GetNonEmptyArrays(0, 0, "developers", bson.M{"developers": 1, "prices": 1, "reviews_score": 1})
	if err != nil {
		return err
	}

	log.Info("Found " + strconv.Itoa(len(appsWithDevelopers)) + " apps with developers")

	newDevelopers := make(map[int]*statsRow)
	for _, app := range appsWithDevelopers {

		// if len(app.Developers) == 0 {
		// 	appDevelopers = []string{""}
		// }

		// For each developer in an app
		for _, appDeveloperID := range app.Developers {

			delete(developersToDelete, appDeveloperID)

			var developersName string
			if val, ok := developersNameMap[appDeveloperID]; ok {
				developersName = val
			} else {
				continue
			}

			if _, ok := newDevelopers[appDeveloperID]; ok {
				newDevelopers[appDeveloperID].count++
				newDevelopers[appDeveloperID].totalScore += app.ReviewsScore
			} else {
				newDevelopers[appDeveloperID] = &statsRow{
					name:       developersName,
					count:      1,
					totalPrice: map[steamapi.ProductCC]int{},
					totalScore: app.ReviewsScore,
				}
			}

			for _, code := range i18n.GetProdCCs(true) {
				price := app.Prices.Get(code.ProductCode)
				if price.Exists {
					newDevelopers[appDeveloperID].totalPrice[code.ProductCode] += price.Final
				}
			}
		}
	}

	var limit int
	var wg sync.WaitGroup

	// Delete old developers
	limit++
	wg.Add(1)
	go func() {

		defer func() {
			limit--
			wg.Done()
		}()

		var devsToDeleteSlice []int
		for k := range developersToDelete {
			devsToDeleteSlice = append(devsToDeleteSlice, k)
		}

		err := mysql.DeleteDevelopers(devsToDeleteSlice)
		log.Err(err)
	}()

	wg.Wait()

	gorm, err := mysql.GetMySQLClient()
	if err != nil {
		return err
	}

	// Update current developers
	var count = 1
	for k, v := range newDevelopers {

		if limit >= 2 {
			wg.Wait()
		}

		limit++
		wg.Add(1)
		go func(developerInt int, v *statsRow) {

			defer func() {
				limit--
				wg.Done()
			}()

			var developer mysql.Developer

			gorm = gorm.Unscoped().FirstOrInit(&developer, mysql.Developer{ID: developerInt})
			log.Err(gorm.Error)

			developer.Name = v.name
			developer.Apps = v.count
			developer.MeanPrice = v.getMeanPrice()
			developer.MeanScore = v.getMeanScore()
			developer.DeletedAt = nil

			gorm = gorm.Unscoped().Save(&developer)
			log.Err(gorm.Error)

		}(k, v)

		count++
	}
	wg.Wait()

	// Clear cache
	return memcache.Delete(
		memcache.MemcacheDeveloperKeyNames.Key,
	)
}
