package queue

import (
	"errors"
	"strconv"
	"time"

	"github.com/gamedb/website/config"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/websockets"
	"github.com/streadway/amqp"
)

type changeMessage struct {
	ID          int                      `json:"id"`
	PICSChanges RabbitMessageChangesPICS `json:"PICSChanges"`
}

type changeQueue struct {
	baseQueue
}

func (q changeQueue) processMessage(msg amqp.Delivery) {

	var err error
	var payload = baseMessage{
		Message: changeMessage{},
	}

	err = helpers.Unmarshal(msg.Body, &payload)
	if err != nil {
		logError(err)
		payload.ack(msg)
		return
	}

	message, ok := payload.Message.(changeMessage)
	if !ok {
		logError(errors.New("can not type assert changeMessage"))
		payload.ack(msg)
		return
	}

	logInfo("Consuming change " + strconv.Itoa(message.ID) + ", attempt " + strconv.Itoa(payload.Attempt))

	// Group products by change id
	changes := map[int]*db.Change{}

	for _, v := range message.PICSChanges.AppChanges {
		if _, ok := changes[v.ChangeNumber]; ok {
			changes[v.ChangeNumber].Apps = append(changes[v.ChangeNumber].Apps, db.ChangeItem{ID: v.ID})

		} else {
			changes[v.ChangeNumber] = &db.Change{
				CreatedAt: time.Now(),
				ChangeID:  v.ChangeNumber,
				Apps:      []db.ChangeItem{{v.ID, ""}},
			}
		}
	}

	for _, v := range message.PICSChanges.PackageChanges {
		if _, ok := changes[v.ChangeNumber]; ok {
			changes[v.ChangeNumber].Packages = append(changes[v.ChangeNumber].Packages, db.ChangeItem{ID: v.ID})
		} else {
			changes[v.ChangeNumber] = &db.Change{
				CreatedAt: time.Now(),
				ChangeID:  v.ChangeNumber,
				Packages:  []db.ChangeItem{{v.ID, ""}},
			}
		}
	}

	// Get apps slice
	var appsSlice []int
	for _, v := range message.PICSChanges.AppChanges {
		appsSlice = append(appsSlice, v.ID)
	}

	var packagesSlice []int
	for _, v := range message.PICSChanges.PackageChanges {
		packagesSlice = append(packagesSlice, v.ID)
	}

	// Get mysql rows
	appRows, err := db.GetAppsByID(appsSlice, []string{"id", "name"})
	logError(err)

	packageRows, err := db.GetPackages(packagesSlice, []string{"id", "name"})
	logError(err)

	// Make map
	appRowsMap := map[int]db.App{}
	for _, v := range appRows {
		appRowsMap[v.ID] = v
	}

	packageRowsMap := map[int]db.Package{}
	for _, v := range packageRows {
		packageRowsMap[v.ID] = v
	}

	// Fill in the change item names
	for changeID, change := range changes {

		for k, changeItem := range change.Apps {
			if val, ok := appRowsMap[changeItem.ID]; ok {
				changes[changeID].Apps[k].Name = val.GetName()
			}
		}
		for k, changeItem := range change.Packages {
			if val, ok := packageRowsMap[changeItem.ID]; ok {
				changes[changeID].Packages[k].Name = val.GetName()
			}
		}
	}

	// Make changes into slice for bulk add
	var changesSlice []db.Kind
	for _, v := range changes {
		changesSlice = append(changesSlice, *v)
	}

	// Save change to DS
	if config.Config.IsProd() {
		err = db.BulkSaveKinds(changesSlice, db.KindChange, true)
		if err != nil {
			logError(err)
			payload.ackRetry(msg)
			return
		}
	}

	// Send websocket
	page, err := websockets.GetPage(websockets.PageChanges)
	if err != nil {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	if page.HasConnections() {

		// Make websocket
		var ws [][]interface{}
		for _, v := range changes {

			ws = append(ws, v.OutputForJSON())
		}

		page.Send(ws)
	}

	payload.ack(msg)
}

type RabbitMessageChangesPICS struct {
	LastChangeNumber    int  `json:"LastChangeNumber"`
	CurrentChangeNumber int  `json:"CurrentChangeNumber"`
	RequiresFullUpdate  bool `json:"RequiresFullUpdate"`
	PackageChanges      map[string]struct {
		ID           int  `json:"ID"`
		ChangeNumber int  `json:"ChangeNumber"`
		NeedsToken   bool `json:"NeedsToken"`
	} `json:"PackageChanges"`
	AppChanges map[string]struct {
		ID           int  `json:"ID"`
		ChangeNumber int  `json:"ChangeNumber"`
		NeedsToken   bool `json:"NeedsToken"`
	} `json:"AppChanges"`
	JobID steamKitJob `json:"JobID"`
}
