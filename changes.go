package main

import (
	"net/http"
	"time"

	"github.com/Jleagle/go-helpers/logger"
	"github.com/go-chi/chi"
	"github.com/kr/pretty"
	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/queue"
)

func changesHandler(w http.ResponseWriter, r *http.Request) {

	template := changesTemplate{}
	template.SetSession(r)

	// Get changes
	changes, err := datastore.GetLatestChanges(100)
	if err != nil {
		logger.Error(err)
	}

	queue.ChangeProducer(&datastore.Change{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ChangeID:  123,
		Apps:      []int{922, 923},
		Packages:  []int{8491, 31794},
	})

	// Get apps/packages
	appIDs := make([]int, 0)
	packageIDs := make([]int, 0)
	for _, v := range changes {
		appIDs = append(appIDs, v.Apps...)
		packageIDs = append(packageIDs, v.Packages...)
	}

	// Get apps for all changes
	appsMap := make(map[int]mysql.App)
	apps, err := mysql.GetApps(appIDs, []string{"id", "name"})

	for _, v := range apps {
		appsMap[v.ID] = v
	}

	// Get packages for all changes
	packagesMap := make(map[int]mysql.Package)
	packages, err := mysql.GetPackages(packageIDs, []string{"id", "name"})

	for _, v := range packages {
		packagesMap[v.ID] = v
	}

	pretty.Println(appsMap)

	// todo, sort packagesMap by id

	// Template
	template.Changes = changes
	template.Apps = appsMap
	template.Packages = packagesMap

	returnTemplate(w, "changes", template)
}

// todo, Just pass through a new struct with all the correct info instead of changes and maps to get names
type changesTemplate struct {
	GlobalTemplate
	Changes  []datastore.Change
	Apps     map[int]mysql.App
	Packages map[int]mysql.Package
}

func changeHandler(w http.ResponseWriter, r *http.Request) {

	change, err := datastore.GetChange(chi.URLParam(r, "id"))
	if err != nil {
		logger.Error(err)
		if err.Error() == "datastore: no such entity" {
			returnErrorTemplate(w, 404, "We can't find this change in our database, there may not be one with this ID.")
			return
		}
	}

	template := changeTemplate{}
	template.SetSession(r)
	template.Change = change

	returnTemplate(w, "change", template)
}

type changeTemplate struct {
	GlobalTemplate
	Change *datastore.Change
}
