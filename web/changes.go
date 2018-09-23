package web

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/steam-authority/steam-authority/db"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/structs"
)

const (
	changesLimit = 100
)

func ChangesHandler(w http.ResponseWriter, r *http.Request) {

	// Get page number
	page, err := strconv.Atoi(r.URL.Query().Get("p"))
	if err != nil {
		page = 1
	}

	// Get changes
	var changes []structs.ChangesChangeTemplate
	resp, err := db.GetLatestChanges(changesLimit, page)
	logger.Error(err)

	for _, v := range resp {
		changes = append(changes, structs.ChangesChangeTemplate{
			Change: v,
		})
	}

	//
	var wg = sync.WaitGroup{}

	// Get apps
	wg.Add(1)
	go func() {

		// Get app IDs
		var appIDs []int
		for _, v := range changes {
			appIDs = append(appIDs, v.Change.Apps...)
		}

		// Get apps for all changes
		appsMap := make(map[int]db.App)
		apps, err := db.GetApps(appIDs, []string{"id", "name", "icon"})
		logger.Error(err)

		// Make app map
		for _, v := range apps {
			appsMap[v.ID] = v
		}

		// Add app to changes
		for k, v := range changes {

			for _, vv := range v.Change.Apps {

				if val, ok := appsMap[vv]; ok {

					changes[k].Apps = append(changes[k].Apps, val)
				}
			}
		}

		wg.Done()

	}()

	// Get packages
	wg.Add(1)
	go func() {

		// Get package IDs
		var packageIDs []int
		for _, v := range changes {
			packageIDs = append(packageIDs, v.Change.Packages...)
		}

		// Get packages for all changes
		packagesMap := make(map[int]db.Package)
		packages, err := db.GetPackages(packageIDs, []string{"id", "name"})
		logger.Error(err)

		// Make package map
		for _, v := range packages {
			packagesMap[v.ID] = v
		}

		// Add app to changes
		for k, v := range changes {

			for _, vv := range v.Change.Apps {

				if val, ok := packagesMap[vv]; ok {

					changes[k].Packages = append(changes[k].Packages, val)
				}
			}
		}

		wg.Done()

	}()

	// Wait
	wg.Wait()

	// Template
	t := changesTemplate{}
	t.Fill(w, r, "Changes")
	t.Changes = changes

	// todo, stop using limit offset for datastore
	t.Pagination = Pagination{
		path:  "/changes?p=",
		page:  page,
		limit: changesLimit,
		total: changesLimit * 100, // 100 Pages
	}

	returnTemplate(w, r, "changes", t)
}

type changesTemplate struct {
	GlobalTemplate
	Changes    []structs.ChangesChangeTemplate
	Pagination Pagination
}
