package tasks

import (
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/robfig/cron/v3"
)

func init() {
	for _, v := range tasks {
		TaskRegister[v.ID()] = v
	}
}

var (
	Parser = cron.NewParser(cron.Minute | cron.Hour)
	tasks  = []TaskInterface{
		AppPlayers{},
		AppQueueAll{},
		AutoPlayerRefreshes{},
		ClearUpcomingCache{},
		DevCodeRun{},
		Developers{},
		Genres{},
		Instagram{},
		MemcacheClear{},
		SetBadgeCache{},
		PackagesQueueAll{},
		PlayerRanks{},
		PlayersQueueAll{},
		Publishers{},
		SteamClientPlayers{},
		Tags{},
		Wishlists{},
	}
)

var TaskRegister = map[string]TaskInterface{}

type TaskInterface interface {
	ID() string
	Name() string
	Cron() string
	work()

	// Base
	Next() (t time.Time)
	Prev() (t time.Time)
	Bad() bool
}

type BaseTask struct {
	TaskInterface
}

func (task BaseTask) Next() (t time.Time) {

	sched, err := Parser.Parse(task.Cron())
	if err != nil {
		return t
	}
	return sched.Next(time.Now())
}

func (task BaseTask) Prev() (d time.Time) {

	sched, err := Parser.Parse(task.Cron())
	if err != nil {
		return d
	}
	next := sched.Next(time.Now())
	nextNext := sched.Next(next)
	diff := nextNext.Sub(next)

	return next.Add(-diff)
}

func (task BaseTask) Bad() (b bool) {

	if task.Cron() == "" {
		return false
	}

	config, err := GetTaskConfig(task)
	if err == nil {
		i, err := strconv.ParseInt(config.Value, 10, 64)
		if err == nil {
			return task.Prev().Unix() > i
		}
	}

	return true
}

//
func Run(task TaskInterface) {

	log.Info("Cron started: " + task.Name())

	// Send websocket
	page := websockets.GetPage(websockets.PageAdmin)
	page.Send(websockets.AdminPayload{TaskID: task.ID(), Action: "started"})

	// Do work
	task.work()

	// Save config row
	err := sql.SetConfig(sql.ConfigID("task-"+task.ID()), strconv.FormatInt(time.Now().Unix(), 10))
	log.Err(err)

	// Send websocket
	page = websockets.GetPage(websockets.PageAdmin)
	page.Send(websockets.AdminPayload{
		TaskID: task.ID(),
		Action: "finished",
		Time:   BaseTask{task}.Next().Unix(),
	})

	//
	log.Info("Cron complete: " + task.Name())
}

func GetTaskConfig(task TaskInterface) (config sql.Config, err error) {
	return sql.GetConfig(sql.ConfigID("task-" + task.ID()))
}

//
func statsLogger(tableName string, count int, total int, rowName string) {
	log.Info("Updating " + tableName + " - " + strconv.Itoa(count) + " / " + strconv.Itoa(total) + ": " + rowName)
}
