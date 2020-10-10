package pages

import (
	"net/http"
	"strings"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/go-chi/chi"
)

func QueuesRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", queuesHandler)
	r.Get("/queues.json", queuesAjaxHandler)
	return r
}

func queuesHandler(w http.ResponseWriter, r *http.Request) {

	t := queuesTemplate{}
	t.fill(w, r, "Queues", "When new items get added to the site, they go through a queue to not overload the servers.")
	t.addAssetHighCharts()
	t.Charts = []string{"Players", "Groups", "Apps", "Packages", "Bundles", "Changes"}

	returnTemplate(w, r, "queues", t)
}

type queuesTemplate struct {
	globalTemplate
	Charts []string
}

func queuesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	var item = memcache.MemcacheQueues
	var highcharts = map[string]influx.HighChartsJSON{}

	err := memcache.GetSetInterface(item.Key, item.Expiration, &highcharts, func() (interface{}, error) {

		// just get ones with prefix of frontend
		var fields = []string{
			string(queue.QueueApps),
			string(queue.QueueBundles),
			string(queue.QueueChanges),
			string(queue.QueueGroups),
			string(queue.QueuePackages),
			string(queue.QueuePlayers),
		}

		builder := influxql.NewBuilder()
		builder.AddSelect(`sum("messages")`, "sum_messages")
		builder.SetFrom(influx.InfluxTelegrafDB, influx.InfluxRetentionPolicy14Day.String(), influx.InfluxMeasurementRabbitQueue.String())
		builder.AddWhere("time", ">=", "now() - 1h")
		builder.AddWhereRaw(`"queue" =~ /^(` + strings.Join(fields, "|") + `)/`)
		builder.AddGroupByTime("10s")
		builder.AddGroupBy("queue")
		builder.SetFillNone()

		resp, err := influx.InfluxQuery(builder)
		if err != nil {
			log.ErrS(builder.String())
			return highcharts, err
		}

		ret := map[string]influx.HighChartsJSON{}
		if len(resp.Results) > 0 {
			for _, v := range resp.Results[0].Series {
				ret[strings.Replace(v.Tags["queue"], "GameDB_Go_", "", 1)] = influx.InfluxResponseToHighCharts(v)
			}
		}

		return ret, err
	})

	if err != nil {
		log.ErrS(err)
		return
	}

	returnJSON(w, r, highcharts)
}
