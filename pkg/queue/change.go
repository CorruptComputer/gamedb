package queue

import (
	"sort"
	"strconv"
	"strings"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/websockets"
	"go.mongodb.org/mongo-driver/bson"
)

type ChangesMessage struct {
	AppIDs     map[uint32]uint32 `json:"app_ids"`
	PackageIDs map[uint32]uint32 `json:"package_ids"`
}

func changesHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := ChangesMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			return
		}

		// Group products by change ID
		changes := map[uint32]*mongo.Change{}

		for changeNumber, appID := range payload.AppIDs {
			if _, ok := changes[changeNumber]; ok {
				changes[changeNumber].Apps = append(changes[changeNumber].Apps, int(appID))
			} else {
				changes[changeNumber] = &mongo.Change{
					CreatedAt: message.FirstSeen(),
					ID:        int(changeNumber),
					Apps:      []int{int(appID)},
				}
			}
		}

		for changeNumber, packageID := range payload.PackageIDs {
			if _, ok := changes[changeNumber]; ok {
				changes[changeNumber].Packages = append(changes[changeNumber].Packages, int(packageID))
			} else {
				changes[changeNumber] = &mongo.Change{
					CreatedAt: message.FirstSeen(),
					ID:        int(changeNumber),
					Packages:  []int{int(packageID)},
				}
			}
		}

		// Convert map to slice sor soeting
		var changeSlice []*mongo.Change
		for _, v := range changes {
			changeSlice = append(changeSlice, v)
		}

		sort.Slice(changeSlice, func(i, j int) bool {
			return changeSlice[i].ID < changeSlice[j].ID
		})

		// Save to Mongo
		err = saveChangesToMongo(changeSlice)
		if err != nil && !strings.Contains(err.Error(), "duplicate key error collection") {
			log.Err(err)
			sendToRetryQueue(message)
			return
		}

		// Get apps and packages for all changes in message
		appMap, packageMap, err := getChangesAppsAndPackages(changeSlice)
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			return
		}

		// Send websocket
		err = sendChangesWebsocket(changeSlice, appMap, packageMap)
		log.Err(err)

		// Send to Discord
		// err = sendChangeToDiscord(changeSlice, appMap, packageMap)
		// log.Err(err)

		message.Ack(false)
	}
}

func saveChangesToMongo(changes []*mongo.Change) (err error) {

	if len(changes) == 0 {
		return nil
	}

	var changesDocuments []mongo.Document
	for _, v := range changes {

		changesDocuments = append(changesDocuments, mongo.Change{
			ID:        v.ID,
			CreatedAt: v.CreatedAt,
			Apps:      v.Apps,
			Packages:  v.Packages,
		})
	}

	_, err = mongo.InsertMany(mongo.CollectionChanges, changesDocuments)
	return err
}

func getChangesAppsAndPackages(changes []*mongo.Change) (appMap map[int]string, packageMap map[int]string, err error) {

	appMap = map[int]string{}
	packageMap = map[int]string{}

	var appIDs []int
	var packageIDs []int

	for _, v := range changes {
		appIDs = append(appIDs, v.Apps...)
		packageIDs = append(packageIDs, v.Packages...)
	}

	// Apps & packages for all changes
	apps, err := mongo.GetAppsByID(appIDs, bson.M{"_id": 1, "name": 1})
	if err != nil {
		log.Err(err)
	}

	for _, v := range apps {
		appMap[v.ID] = v.GetName()
	}

	packages, err := mongo.GetPackagesByID(packageIDs, bson.M{"id": 1, "name": 1})
	if err != nil {
		return appMap, packageMap, err
	}

	for _, v := range packages {
		packageMap[v.ID] = v.GetName()
	}

	return appMap, packageMap, err
}

func sendChangesWebsocket(changes []*mongo.Change, appMap map[int]string, packageMap map[int]string) (err error) {

	var ws [][]interface{}
	for _, v := range changes {

		ws = append(ws, v.OutputForJSON(appMap, packageMap))
	}

	wsPayload := ChangesPayload{Data: ws}
	return ProduceWebsocket(wsPayload, websockets.PageChanges)
}

// todo, add packages to return
func sendChangeToDiscord(changes []*mongo.Change, appMap map[int]string, packageMap map[int]string) (err error) {

	if config.IsProd() {

		for _, change := range changes {

			var apps []string

			for _, v := range change.Apps {
				if val, ok := appMap[v]; ok {
					apps = append(apps, val)
				}
			}

			if len(apps) > 0 {

				var msg = "Change " + strconv.Itoa(change.ID) + ": " + strings.Join(apps, ", ")
				_, err := discordClient.ChannelMessageSend("574563721045606431", msg)
				log.Err(err)
			}
		}
	}

	return nil
}
