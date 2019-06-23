package pages

import (
	"errors"
	"net/http"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
	"github.com/google/go-github/github"
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

	client, ctx := helpers.GetGithub()

	operation := func() (err error) {

		contributors, _, err := client.Repositories.ListContributorsStats(ctx, "gamedb", "website")
		for _, v := range contributors {
			t.Total += v.GetTotal()
		}

		if t.Total == 0 {
			return errors.New("no commits found")
		}

		return nil
	}

	policy := backoff.NewExponentialBackOff()

	err := backoff.RetryNotify(operation, backoff.WithMaxRetries(policy, 3), func(err error, t time.Duration) { log.Info(err, r) })
	if err != nil {
		log.Critical(err, r)
	}

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

	commitsResponse, _, err := client.Repositories.ListCommits(ctx, "gamedb", "website", &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			Page:    query.getPage(commitsLimit),
			PerPage: commitsLimit,
		},
	})

	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the commits.", Error: err})
		return
	}

	// Get commits
	var commits []commitStruct

	var deployed bool
	for _, commit := range commitsResponse {

		if commit.GetSHA() == config.Config.CommitHash.Get() {
			deployed = true
		}

		commits = append(commits, commitStruct{
			Message:   helpers.InsertNewLines(commit.Commit.GetMessage(), 10),
			Time:      commit.Commit.Author.Date.Unix(),
			Deployed:  deployed,
			Link:      commit.GetHTMLURL(),
			Highlight: commit.GetSHA() == config.Config.CommitHash.Get(),
			Hash:      commit.GetSHA()[0:7],
		})
	}

	// Get total
	var total int
	operation := func() (err error) {

		contributors, _, err := client.Repositories.ListContributorsStats(ctx, "gamedb", "website")
		for _, v := range contributors {
			total += v.GetTotal()
		}
		if total == 0 {
			return errors.New("no contributors found")
		}
		return nil
	}

	policy := backoff.NewExponentialBackOff()

	err = backoff.RetryNotify(operation, backoff.WithMaxRetries(policy, 2), func(err error, t time.Duration) { log.Info(err, r) })
	log.Err(err)

	//
	response := DataTablesAjaxResponse{}
	response.RecordsTotal = int64(total)
	response.RecordsFiltered = int64(total)
	response.Draw = query.Draw
	response.limit(r)

	for _, v := range commits {
		response.AddRow(v.OutputForJSON())
	}

	response.output(w, r)
}

type commitStruct struct {
	Message   string
	Deployed  bool
	Time      int64
	Link      string
	Highlight bool
	Hash      string
}

func (commit commitStruct) OutputForJSON() (output []interface{}) {

	return []interface{}{
		commit.Message,
		commit.Time,
		commit.Deployed,
		commit.Link,
		commit.Highlight,
		commit.Hash,
	}
}
