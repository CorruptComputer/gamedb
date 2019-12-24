package queue

import (
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/cenkalti/backoff/v3"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/helpers/twitch"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue/framework"
	"github.com/gamedb/gamedb/pkg/sql"
	influx "github.com/influxdata/influxdb1-client"
	"github.com/nicklaw5/helix"
)

type AppPlayerMessage struct {
	IDs []int `json:"ids"`
}

func appPlayersHandler(messages []*framework.Message) {

	for _, message := range messages {

		payload := AppPlayerMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			return
		}

		// Get apps
		appMap := map[int]sql.App{}
		apps, err := sql.GetAppsByID(payload.IDs, []string{"id", "twitch_id"})
		if err != nil {
			log.Err(err, payload.IDs)
			sendToRetryQueue(message)
			return
		}

		for _, v := range apps {
			appMap[v.ID] = v
		}

		for _, appID := range payload.IDs {

			app, ok := appMap[appID]
			if ok {

				var wg sync.WaitGroup

				// Reads
				wg.Add(1)
				var viewers int
				go func() {

					defer wg.Done()

					var err error
					viewers, err = getAppTwitchStreamers(app.TwitchID)
					if err != nil {
						log.Err(err, payload.IDs)
						sendToRetryQueue(message)
						return
					}
				}()

				wg.Add(1)
				var appPlayersWeek int64
				go func() {

					defer wg.Done()

					var err error
					appPlayersWeek, err = getAppTopPlayersWeek(appID)
					if err != nil {
						log.Err(err, payload.IDs)
						sendToRetryQueue(message)
						return
					}
				}()

				wg.Add(1)
				var appPlayersWeekAverage float64
				go func() {

					defer wg.Done()

					var err error
					appPlayersWeekAverage, err = getAppAveragePlayersWeek(appID)
					if err != nil {
						log.Err(err, payload.IDs)
						sendToRetryQueue(message)
						return
					}
				}()

				wg.Add(1)
				var appPlayersAlltime int64
				go func() {

					defer wg.Done()

					var err error
					appPlayersAlltime, err = getAppTopPlayersAlltime(appID)
					if err != nil {
						log.Err(err, payload.IDs)
						sendToRetryQueue(message)
						return
					}
				}()

				wg.Add(1)
				var appTrend int64
				go func() {

					defer wg.Done()

					var err error
					appTrend, err = getAppTrendValue(appID)
					if err != nil {
						log.Err(err, payload.IDs)
						sendToRetryQueue(message)
						return
					}
				}()

				wg.Add(1)
				var appPlayersNow int
				go func() {

					defer wg.Done()

					var err error
					appPlayersNow, err = getAppOnlinePlayers(appID)
					if err != nil {
						steam.LogSteamError(err, payload.IDs)
						sendToRetryQueue(message)
						return
					}
				}()

				wg.Wait()

				if message.ActionTaken {
					continue
				}

				// Save counts to Influx
				wg.Add(1)
				go func() {

					defer wg.Done()

					err = saveAppPlayerToInflux(appID, viewers, appPlayersNow)
					if err != nil {
						log.Err(err, payload.IDs)
						sendToRetryQueue(message)
						return
					}
				}()

				// Save to MySQL
				wg.Add(1)
				go func() {

					defer wg.Done()

					err = updateAppPlayerInfoRow(appID, appTrend, appPlayersWeek, appPlayersAlltime, appPlayersWeekAverage)
					if err != nil {
						log.Err(err, payload.IDs)
						sendToRetryQueue(message)
						return
					}
				}()

				wg.Wait()

				if message.ActionTaken {
					continue
				}
			}
		}

		//
		message.Ack()
	}
}

func getAppTwitchStreamers(twitchID int) (viewers int, err error) {

	if twitchID > 0 {

		client, err := twitch.GetTwitch()
		if err != nil {
			return 0, err
		}

		var resp *helix.StreamsResponse

		// Retrying as this call can fail
		operation := func() (err error) {

			resp, err = client.GetStreams(&helix.StreamsParams{First: 100, GameIDs: []string{strconv.Itoa(twitchID)}, Language: []string{"en"}})
			return err
		}

		policy := backoff.NewExponentialBackOff()

		err = backoff.RetryNotify(operation, backoff.WithMaxRetries(policy, 3), func(err error, t time.Duration) { log.Info(err) })
		if err != nil {
			return 0, err
		}

		for _, v := range resp.Data.Streams {
			viewers += v.ViewerCount
		}
	}

	return viewers, nil
}

func getAppOnlinePlayers(appID int) (count int, err error) {

	// var regexIntsOnly = regexp.MustCompile("[^0-9]+")
	//
	// c := colly.NewCollector(
	//   colly.AllowURLRevisit()
	// )
	// c.SetRequestTimeout(time.Second * 5)
	//
	// // ID
	// c.OnHTML(".apphub_NumInApp", func(e *colly.HTMLElement) {
	// 	e.Text = regexIntsOnly.ReplaceAllString(e.Text, "")
	// 	log.Info(e.Text)
	// })
	//
	// //
	// c.OnError(func(r *colly.Response, err error) {
	// 	helpers.LogSteamError(err)
	// })
	//
	// err2 := c.Visit("https://steamcommunity.com/app/440")
	// log.Err(err2)

	client := steam.GetSteamUnlimited()

	count, b, err := client.GetNumberOfCurrentPlayers(appID)
	err = steam.AllowSteamCodes(err, b, []int{404})
	return count, err
}

func saveAppPlayerToInflux(appID int, viewers int, players int) (err error) {

	_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementApps),
		Tags: map[string]string{
			"app_id": strconv.Itoa(appID),
		},
		Fields: map[string]interface{}{
			"player_count":   players,
			"twitch_viewers": viewers,
		},
		Time:      time.Now(),
		Precision: "m",
	})

	return err
}

func getAppTopPlayersWeek(appID int) (val int64, err error) {

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.SetFrom(influxHelper.InfluxGameDB, influxHelper.InfluxRetentionPolicyAllTime.String(), influxHelper.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW() - 7d")
	builder.AddWhere("app_id", "=", appID)
	builder.SetFillNone()

	resp, err := influxHelper.InfluxQuery(builder.String())
	if err != nil {
		return 0, err
	}

	return influxHelper.GetFirstInfluxInt(resp), nil
}

func getAppAveragePlayersWeek(appID int) (val float64, err error) {

	builder := influxql.NewBuilder()
	builder.AddSelect("mean(player_count)", "mean_player_count")
	builder.SetFrom(influxHelper.InfluxGameDB, influxHelper.InfluxRetentionPolicyAllTime.String(), influxHelper.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW() - 7d")
	builder.AddWhere("app_id", "=", appID)

	resp, err := influxHelper.InfluxQuery(builder.String())
	if err != nil {
		return 0, err
	}

	return influxHelper.GetFirstInfluxFloat(resp), nil
}

func getAppTopPlayersAlltime(appID int) (val int64, err error) {

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.SetFrom(influxHelper.InfluxGameDB, influxHelper.InfluxRetentionPolicyAllTime.String(), influxHelper.InfluxMeasurementApps.String())
	builder.AddWhere("app_id", "=", appID)
	builder.SetFillNone()

	resp, err := influxHelper.InfluxQuery(builder.String())
	if err != nil {
		return 0, err
	}

	return influxHelper.GetFirstInfluxInt(resp), nil
}

func getAppTrendValue(appID int) (trend int64, err error) {

	// Trend value - https://stackoverflow.com/questions/41361734/get-difference-since-30-days-ago-in-influxql-influxdb
	subBuilder := influxql.NewBuilder()
	subBuilder.AddSelect("difference(last(player_count))", "")
	subBuilder.SetFrom(influxHelper.InfluxGameDB, influxHelper.InfluxRetentionPolicyAllTime.String(), influxHelper.InfluxMeasurementApps.String())
	subBuilder.AddWhere("app_id", "=", appID)
	subBuilder.AddWhere("time", ">=", "NOW() - 7d")
	subBuilder.AddGroupByTime("1h")

	builder := influxql.NewBuilder()
	builder.AddSelect("cumulative_sum(difference)", "")
	builder.SetFromSubQuery(subBuilder)

	resp, err := influxHelper.InfluxQuery(builder.String())
	if err != nil {
		return 0, err
	}

	var trendTotal int64

	// Get the last value, todo, put into influx helper, like the ones below
	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {
		values := resp.Results[0].Series[0].Values
		if len(values) > 0 {

			last := values[len(values)-1]

			trendTotal, err = last[1].(json.Number).Int64()
			if err != nil {
				return 0, err
			}
		}
	}

	return trendTotal, nil
}

func updateAppPlayerInfoRow(appID int, trend int64, week int64, alltime int64, average float64) (err error) {

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"player_trend":        trend,
		"player_peak_week":    week,
		"player_peak_alltime": alltime,
		"player_avg_week":     average,
	}

	gorm.Table("apps").Where("id = ?", appID).Updates(data)

	return gorm.Error
}
