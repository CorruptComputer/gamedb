package pages

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
	"github.com/google/go-github/v27/github"
)

const (
	commitsLimit = 100
)

func CommitsRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", commitsHandler)
	r.Get("/commits.json", commitsAjaxHandler)
	return r
}

func commitsHandler(w http.ResponseWriter, r *http.Request) {

	t := commitsTemplate{}
	t.fill(w, r, "Commits", "")

	var err error
	t.Total, err = getTotalCommits()
	log.Err(err)

	err = returnTemplate(w, r, "commits", t)
	log.Err(err, r)
}

type commitsTemplate struct {
	GlobalTemplate
	Total int
}

func commitsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	query.limit(r)

	client, ctx := helpers.GetGithub()

	commits, _, err := client.Repositories.ListCommits(ctx, "gamedb", "website", &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			Page:    query.getPage(commitsLimit),
			PerPage: commitsLimit,
		},
	})

	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the commits.", Error: err})
		return
	}

	// Get total
	total, err := getTotalCommits()
	log.Err(err)

	//
	response := DataTablesAjaxResponse{}
	response.RecordsTotal = int64(total)
	response.RecordsFiltered = int64(total)
	response.Draw = query.Draw
	response.limit(r)

	var deployed bool
	for _, commit := range commits {

		if strings.HasPrefix(commit.Commit.GetMessage(), "Merge pull request #") {
			continue
		}

		if commit.GetSHA() == config.Config.CommitHash.Get() {
			deployed = true
		}

		response.AddRow([]interface{}{
			helpers.InsertNewLines(commit.Commit.GetMessage()),
			commit.Commit.Author.Date.Unix(),
			deployed,
			commit.GetHTMLURL(),
			commit.GetSHA() == config.Config.CommitHash.Get(),
			commit.GetSHA()[0:7],
		})
	}

	response.output(w, r)
}

func getTotalCommits() (total int, err error) {

	client, ctx := helpers.GetGithub()

	var item = helpers.MemcacheTotalCommits

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &total, func() (interface{}, error) {

		operation := func() (err error) {

			contributors, _, err := client.Repositories.ListContributorsStats(ctx, "gamedb", "gamedb")
			for _, v := range contributors {
				total += v.GetTotal()
			}
			if total == 0 {
				return errors.New("no commits found")
			}
			return nil
		}

		policy := backoff.NewExponentialBackOff()

		err = backoff.RetryNotify(operation, backoff.WithMaxRetries(policy, 4), func(err error, t time.Duration) { log.Info(err) })

		return total, err
	})

	return total, err
}
