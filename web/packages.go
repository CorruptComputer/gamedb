package web

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

const (
	packagesLimit = 100
)

func PackagesHandler(w http.ResponseWriter, r *http.Request) {

	// Get page number
	page, err := strconv.Atoi(r.URL.Query().Get("p"))
	if err != nil {
		page = 1
	}

	var wg sync.WaitGroup

	// Get total changes
	var total int
	wg.Add(1)
	go func() {

		total, err = mysql.CountPackages()
		if err != nil {
			logger.Error(err)
		}

		wg.Done()

	}()

	// Get changes
	var packages []mysql.Package
	wg.Add(1)
	go func() {

		packages, err = mysql.GetLatestPackages(packagesLimit, page)
		if err != nil {
			logger.Error(err)
		}

		wg.Done()

	}()

	// Wait
	wg.Wait()

	// Template
	template := packagesTemplate{}
	template.Fill(r, "Packages")
	template.Packages = packages
	template.Total = total
	template.Pagination = Pagination{
		path:  "/packages?p=",
		page:  page,
		limit: packagesLimit,
		total: total,
	}

	returnTemplate(w, r, "packages", template)
}

type packagesTemplate struct {
	GlobalTemplate
	Packages   []mysql.Package
	Pagination Pagination
	Total      int
}
