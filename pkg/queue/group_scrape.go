package queue

import (
	"encoding/json"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/gocolly/colly"
	influx "github.com/influxdata/influxdb1-client"
	"github.com/powerslacker/ratelimit"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	groupXMLRateLimit    = ratelimit.New(1, ratelimit.WithCustomDuration(1, time.Second*60), ratelimit.WithoutSlack)
	groupScrapeRateLimit = ratelimit.New(1, ratelimit.WithCustomDuration(1, time.Second), ratelimit.WithoutSlack)
)

type groupMessage struct {
	baseMessage
	Message groupMessageInner `json:"message"`
}

type groupMessageInner struct {
	IDs []string `json:"ids"`
}

type groupQueueScrape struct {
}

//noinspection GoNilness
func (q groupQueueScrape) processMessages(msgs []amqp.Delivery) {

	msg := msgs[0]

	var err error

	message := groupMessage{}
	message.OriginalQueue = queueGroups

	err = helpers.Unmarshal(msg.Body, &message)
	if err != nil {
		log.Err(err, msg.Body)
		ackFail(msg, &message)
		return
	}

	//
	for _, groupID := range message.Message.IDs {

		group, err := mongo.GetGroup(groupID)
		if err != nil && err != mongo.ErrNoDocuments {
			log.Err(err, groupID)
			ackRetry(msg, &message)
			return
		}

		// Skip if updated recently
		if config.IsProd() && group.UpdatedAt.Unix() > time.Now().Add(time.Hour*-1).Unix() {
			continue
		}

		// Get `type` if missing
		if group.Type == "" {

			group.Type, err = getGroupType(groupID)
			if err != nil {
				steam.LogSteamError(err, groupID)
				ackRetry(msg, &message)
				return
			}

			// Deleted groups can not redirect to get a type.
			if group.Type == "" {
				message.ack(msg)
				return
			}
		}

		// Update group
		var found bool
		if group.Type == helpers.GroupTypeGame {
			found, err = updateGameGroup(groupID, &group)
		} else {
			found, err = updateRegularGroup(groupID, &group)
		}

		// Skip if we cant find numbers
		if !found {
			log.Info("Group counts not found", groupID)
			ackRetry(msg, &message)
			return
		}

		// Some pages dont contain the ID64, so use the API
		if group.ID64 == "" {
			err = produceGroupNew(groupID)
			if err != nil {
				steam.LogSteamError(err, groupID)
			}
			message.ack(msg)
			return
		}

		// Fix group data
		if group.Summary == "No information given." {
			group.Summary = ""
		}

		// Get trending value
		err = getGroupTrending(&group)
		if err != nil {
			log.Err(err, groupID)
			ackRetry(msg, &message)
			return
		}

		//
		var wg sync.WaitGroup

		// Read from MySQL
		wg.Add(1)
		var app sql.App
		go func() {

			defer wg.Done()

			var err error

			app, err = getAppFromGroup(group)
			err = helpers.IgnoreErrors(err, sql.ErrRecordNotFound)
			if err != nil {
				log.Err(err, group.ID64)
				ackRetry(msg, &message)
				return
			}
		}()

		//
		wg.Wait()

		if message.actionTaken {
			return
		}

		// Save to MySQL
		wg.Add(1)
		go func() {

			defer wg.Done()

			err = saveAppsGroupID(app, group)
			if err != nil {
				log.Err(err, group.ID64)
				ackRetry(msg, &message)
				return
			}
		}()

		// Save to Mongo
		wg.Add(1)
		go func() {

			defer wg.Done()

			err = saveGroup(group)
			if err != nil {
				log.Err(err, groupID)
				ackRetry(msg, &message)
				return
			}
		}()

		// Save to Influx
		wg.Add(1)
		go func() {

			defer wg.Done()

			err = saveGroupToInflux(group)
			if err != nil {
				log.Err(err, groupID)
				ackRetry(msg, &message)
				return
			}
		}()

		wg.Wait()

		if message.actionTaken {
			return
		}
	}

	// Clear memcache
	err = memcache.RemoveKeyFromMemCacheViaPubSub(message.Message.IDs...)
	log.Err(err)

	// Send websocket
	err = sendGroupWebsocket(message.Message.IDs)
	log.Err(err)

	//
	message.ack(msg)
}

func updateGameGroup(id string, group *mongo.Group) (foundNumbers bool, err error) {

	groupScrapeRateLimit.Take()

	c := colly.NewCollector()
	c.SetRequestTimeout(time.Second * 15)

	// ID64
	c.OnHTML("a[href^=\"steam:\"]", func(e *colly.HTMLElement) {
		e.Text = helpers.RegexNonInts.ReplaceAllString(e.Text, "")
		group.ID64 = path.Base(e.Attr("href"))
	})

	// URL
	c.OnHTML("#eventsBlock a", func(e *colly.HTMLElement) {
		if strings.HasSuffix(e.Attr("href"), "/events") {
			var url = strings.TrimSuffix(e.Attr("href"), "/events")
			group.URL = path.Base(url)
		}
	})

	// Name
	c.OnHTML("#mainContents > h1", func(e *colly.HTMLElement) {
		var trimmed = strings.TrimSpace(e.Text)
		if trimmed != "" {
			group.Name = trimmed
		}
	})

	// App ID
	c.OnHTML("#rightActionBlock a", func(e *colly.HTMLElement) {
		var url = e.Attr("href")
		if strings.HasSuffix(url, "/discussions") {
			url = strings.TrimSuffix(url, "/discussions")
			url = path.Base(url)
			urli, err := strconv.Atoi(url)
			if err == nil {
				group.AppID = urli
			}
		}
	})

	// Headline
	c.OnHTML("#profileBlock > h1", func(e *colly.HTMLElement) {
		group.Headline = strings.TrimSpace(e.Text)
	})

	// Summary
	c.OnHTML("#summaryText", func(e *colly.HTMLElement) {
		var err error
		group.Summary, err = e.DOM.Html()
		log.Err(err)

		if group.Summary == "No information given." {
			group.Summary = ""
		}
	})

	// Icon
	if group.Icon == "" && group.URL != "" {
		i, err := strconv.Atoi(group.URL)
		if err == nil && i > 0 {
			app, err := sql.GetApp(i, []string{"id", "icon"})
			if err != nil {
				log.Err(group.URL, err)
			} else {
				group.Icon = app.Icon
			}
		}
	}

	// Members / Members In Chat
	c.OnHTML("#profileBlock .linkStandard", func(e *colly.HTMLElement) {
		if strings.Contains(strings.ToLower(e.Text), "chat") {
			e.Text = helpers.RegexNonInts.ReplaceAllString(e.Text, "")
			group.MembersInChat, err = strconv.Atoi(e.Text)
		} else {
			e.Text = helpers.RegexNonInts.ReplaceAllString(e.Text, "")
			group.Members, err = strconv.Atoi(e.Text)
			foundNumbers = true
		}
	})

	// Members In Game
	c.OnHTML("#profileBlock .membersInGame", func(e *colly.HTMLElement) {
		e.Text = helpers.RegexNonInts.ReplaceAllString(e.Text, "")
		group.MembersInGame, err = strconv.Atoi(e.Text)
	})

	// Members Online
	c.OnHTML("#profileBlock .membersOnline", func(e *colly.HTMLElement) {
		e.Text = helpers.RegexNonInts.ReplaceAllString(e.Text, "")
		group.MembersOnline, err = strconv.Atoi(e.Text)
	})

	// Error
	group.Error = ""

	c.OnHTML("#message h3", func(e *colly.HTMLElement) {
		group.Error = e.Text
		foundNumbers = true
	})

	//
	c.OnError(func(r *colly.Response, err error) {
		steam.LogSteamError(err)
	})

	return foundNumbers, c.Visit("https://steamcommunity.com/gid/" + id)
}

var (
	regularGroupID64Regex = regexp.MustCompile(`commentthread_Clan_([0-9]{18})_`)
)

func updateRegularGroup(id string, group *mongo.Group) (foundMembers bool, err error) {

	groupScrapeRateLimit.Take()

	group.AppID = 0

	c := colly.NewCollector()
	c.SetRequestTimeout(time.Second * 60)

	// ID64
	c.OnHTML("[id^=commentthread_Clan_]", func(e *colly.HTMLElement) {
		matches := regularGroupID64Regex.FindStringSubmatch(e.Attr("id"))
		if len(matches) > 1 {
			group.ID64 = matches[1]
		}
	})

	// Abbreviation
	c.OnHTML("div.grouppage_header_name span.grouppage_header_abbrev", func(e *colly.HTMLElement) {
		group.Abbr = strings.TrimPrefix(e.Text, "/ ")
	})

	// Name - Must be after `Abbreviation` as we delete it here.
	c.OnHTML("div.grouppage_header_name", func(e *colly.HTMLElement) {
		e.DOM.Children().Remove()
		var trimmed = strings.TrimSpace(strings.TrimPrefix(e.DOM.Text(), "/ "))
		if trimmed != "" {
			group.Name = trimmed
		}
	})

	// URL
	c.OnHTML("form#join_group_form", func(e *colly.HTMLElement) {
		group.URL = path.Base(e.Attr("action"))
	})

	// Headline
	c.OnHTML("div.group_content.group_summary h1", func(e *colly.HTMLElement) {
		group.Headline = strings.TrimSpace(e.Text)
	})

	// Summary
	c.OnHTML("div.formatted_group_summary", func(e *colly.HTMLElement) {
		summary, err := e.DOM.Html()
		log.Err(err)
		if err == nil {
			group.Summary = strings.TrimSpace(summary)
		}
	})

	// Icon
	c.OnHTML("div.grouppage_logo img", func(e *colly.HTMLElement) {
		group.Icon = strings.TrimPrefix(e.Attr("src"), helpers.AvatarBase)
	})

	// Members
	c.OnHTML("div.membercount.members .count", func(e *colly.HTMLElement) {
		group.Members, err = strconv.Atoi(helpers.RegexNonInts.ReplaceAllString(e.Text, ""))
		foundMembers = true
	})

	// Members In Game
	c.OnHTML("div.membercount.ingame .count", func(e *colly.HTMLElement) {
		group.MembersInGame, err = strconv.Atoi(helpers.RegexNonInts.ReplaceAllString(e.Text, ""))
	})

	// Members Online
	c.OnHTML("div.membercount.online .count", func(e *colly.HTMLElement) {
		group.MembersOnline, err = strconv.Atoi(helpers.RegexNonInts.ReplaceAllString(e.Text, ""))
	})

	// Members In Chat
	c.OnHTML("div.joinchat_membercount .count", func(e *colly.HTMLElement) {
		group.MembersInChat, err = strconv.Atoi(helpers.RegexNonInts.ReplaceAllString(e.Text, ""))
	})

	// Error
	group.Error = ""

	c.OnHTML("#message h3", func(e *colly.HTMLElement) {
		group.Error = e.Text
		foundMembers = true
	})

	//
	c.OnError(func(r *colly.Response, err error) {
		steam.LogSteamError(err)
	})

	return foundMembers, c.Visit("https://steamcommunity.com/gid/" + id)
}

func getGroupTrending(group *mongo.Group) (err error) {

	// Trend value - https://stackoverflow.com/questions/41361734/get-difference-since-30-days-ago-in-influxql-influxdb

	subBuilder := influxql.NewBuilder()
	subBuilder.AddSelect("difference(last(members_count))", "")
	subBuilder.SetFrom(influxHelper.InfluxGameDB, influxHelper.InfluxRetentionPolicyAllTime.String(), influxHelper.InfluxMeasurementGroups.String())
	subBuilder.AddWhere("group_id", "=", group.ID64)
	subBuilder.AddWhere("time", ">=", "NOW() - 21d")
	subBuilder.AddGroupByTime("1h")

	builder := influxql.NewBuilder()
	builder.AddSelect("cumulative_sum(difference)", "")
	builder.SetFromSubQuery(subBuilder)

	resp, err := influxHelper.InfluxQuery(builder.String())
	if err != nil {
		return err
	}

	var trendTotal int64

	// Get the last value, todo, put into influx helper, like the ones below
	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {
		values := resp.Results[0].Series[0].Values
		if len(values) > 0 {

			last := values[len(values)-1]

			trendTotal, err = last[1].(json.Number).Int64()
			if err != nil {
				return err
			}
		}
	}

	group.Trending = trendTotal
	return nil
}

func saveGroup(group mongo.Group) (err error) {

	_, err = mongo.ReplaceOne(mongo.CollectionGroups, bson.D{{"_id", group.ID64}}, group)
	return err
}

func getAppFromGroup(group mongo.Group) (app sql.App, err error) {

	if group.Type == helpers.GroupTypeGame && group.AppID > 0 {
		app, err = sql.GetApp(group.AppID, []string{"id", "group_id"})
		if err == sql.ErrRecordNotFound {
			err = ProduceToSteam(SteamPayload{AppIDs: []int{group.AppID}, Force: true})
		}
	}

	return app, err
}

func saveAppsGroupID(app sql.App, group mongo.Group) (err error) {

	if app.ID == 0 || group.ID64 == "" || app.GroupID == group.ID64 || group.Type != helpers.GroupTypeGame {
		return nil
	}

	db, err := sql.GetMySQLClient()
	if err != nil {
		return err
	}

	db = db.Model(&app).Updates(map[string]interface{}{
		"group_id":        group.ID64,
		"group_followers": group.Members,
	})

	return db.Error
}

func saveGroupToInflux(group mongo.Group) (err error) {

	fields := map[string]interface{}{
		"members_count":   group.Members,
		"members_in_chat": group.MembersInChat,
		"members_in_game": group.MembersInGame,
		"members_online":  group.MembersOnline,
	}

	_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementGroups),
		Tags: map[string]string{
			"group_id":   group.ID64,
			"group_type": group.Type,
		},
		Fields:    fields,
		Time:      time.Now(),
		Precision: "h",
	})

	return err
}

func sendGroupWebsocket(ids []string) (err error) {

	wsPayload := websockets.PubSubIDStringsPayload{} // String as int64 too large for js
	wsPayload.IDs = ids
	wsPayload.Pages = []websockets.WebsocketPage{websockets.PageGroup}

	_, err = helpers.Publish(helpers.PubSubTopicWebsockets, wsPayload)
	return err
}

func getGroupType(id string) (string, error) {

	groupScrapeRateLimit.Take()

	resp, err := http.Get("https://steamcommunity.com/gid/" + id)
	if err != nil {
		return "", err
	}

	defer func() {
		err = resp.Body.Close()
		log.Err(err)
	}()

	u := resp.Request.URL.String()

	if strings.Contains(u, "/games/") {
		return helpers.GroupTypeGame, err
	} else if strings.Contains(u, "/groups/") {
		return helpers.GroupTypeGroup, err
	}

	return "", err
}

func updateGroupFromXML(id string, group *mongo.Group) (err error) {

	groupXMLRateLimit.Take()

	resp, b, err := steam.GetSteam().GetGroupByID(id)
	err = steam.AllowSteamCodes(err, b, nil)
	if err != nil {
		return err
	}

	group.SetID(id)
	group.ID64 = resp.ID64
	if resp.Details.Name != "" {
		group.Name = resp.Details.Name
	}
	group.URL = resp.Details.URL
	group.Headline = resp.Details.Headline
	group.Summary = resp.Details.Summary
	group.Members = int(resp.Details.MemberCount)
	group.MembersInChat = int(resp.Details.MembersInChat)
	group.MembersInGame = int(resp.Details.MembersInGame)
	group.MembersOnline = int(resp.Details.MembersOnline)
	group.Type = resp.Type

	// Try to get App ID from URL
	i, err := strconv.Atoi(resp.Details.URL)
	if err == nil && i > 0 {
		group.AppID = i
	}

	// Get working icon
	if helpers.GetResponseCode(resp.Details.AvatarFull) == 200 {

		group.Icon = strings.Replace(resp.Details.AvatarFull, helpers.AvatarBase, "", 1)

	} else if helpers.GetResponseCode(resp.Details.AvatarMedium) == 200 {

		group.Icon = strings.Replace(resp.Details.AvatarMedium, helpers.AvatarBase, "", 1)

	} else if helpers.GetResponseCode(resp.Details.AvatarIcon) == 200 {

		group.Icon = strings.Replace(resp.Details.AvatarIcon, helpers.AvatarBase, "", 1)

	} else {

		group.Icon = ""
	}

	return nil
}
