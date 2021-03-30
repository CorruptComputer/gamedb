package consumers

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/websockets"
	"go.uber.org/zap"
)

type WebsocketMessage struct {
	Pages   []websockets.WebsocketPage `json:"pages"`
	Message []byte                     `json:"message"`
}

func websocketHandler(message *rabbit.Message) {

	payload := WebsocketMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	for _, page := range payload.Pages {

		wsPage := websockets.GetPage(page)
		if wsPage == nil {
			continue
		}

		if wsPage.CountConnections() == 0 {
			continue
		}

		switch page {
		case websockets.PageApp, websockets.PageBundle, websockets.PagePackage:

			idPayload := IntPayload{}

			err = helpers.Unmarshal(payload.Message, &idPayload)
			if err != nil {
				log.ErrS(err)
				continue
			}

			wsPage.Send(idPayload.ID)

		case websockets.PageNews:

			newsPayload := NewsPayload{}

			err = helpers.Unmarshal(payload.Message, &newsPayload)
			if err != nil {
				log.ErrS(err)
				continue
			}

			wsPage.Send(newsPayload.RowData)

		case websockets.PageGroup:

			idPayload := StringPayload{}

			err = helpers.Unmarshal(payload.Message, &idPayload)
			if err != nil {
				log.ErrS(err)
				continue
			}

			wsPage.Send(idPayload.String)

		case websockets.PagePlayer:

			playerPayload := PlayerPayload{}

			err = helpers.Unmarshal(payload.Message, &playerPayload)
			if err != nil {
				log.ErrS(err)
				continue
			}

			wsPage.Send(playerPayload)

		case websockets.PageChatBot:

			cbPayload := ChatBotPayload{}

			err = helpers.Unmarshal(payload.Message, &cbPayload)
			if err != nil {
				log.ErrS(err)
				continue
			}

			wsPage.Send(cbPayload)

		case websockets.PageAdmin:

			adminPayload := AdminPayload{}

			err = helpers.Unmarshal(payload.Message, &adminPayload)
			if err != nil {
				log.ErrS(err)
				continue
			}

			wsPage.Send(adminPayload)

		case websockets.PageChanges:

			changePayload := ChangesPayload{}

			err = helpers.Unmarshal(payload.Message, &changePayload)
			if err != nil {
				log.ErrS(err)
				continue
			}

			wsPage.Send(changePayload.Data)

		case websockets.PagePackages:

			idPayload := IntPayload{}

			err = helpers.Unmarshal(payload.Message, &idPayload)
			if err != nil {
				log.ErrS(err)
				continue
			}

			pack, err := mongo.GetPackage(idPayload.ID)
			if err != nil {
				log.ErrS(err)
				continue
			}

			wsPage.Send(pack.OutputForJSON(steamapi.ProductCCUS))

		case websockets.PageBundles:

			idPayload := IntPayload{}

			err = helpers.Unmarshal(payload.Message, &idPayload)
			if err != nil {
				log.ErrS(err)
				continue
			}

			var bundle helpers.Bundle
			bundle, err = mongo.GetBundle(idPayload.ID)
			if err != nil {
				log.ErrS(err)
				continue
			} else {
				wsPage.Send(bundle.OutputForJSON())
			}

		case websockets.PagePrices:

			idsPayload := StringsPayload{}

			err = helpers.Unmarshal(payload.Message, &idsPayload)
			if err != nil {
				log.ErrS(err)
				continue
			}

			prices, err := mongo.GetPricesByID(idsPayload.IDs)
			if err != nil {
				log.ErrS(err)
				continue
			} else {
				for _, v := range prices {
					wsPage.Send(v.OutputForJSON())
				}
			}

		default:

			log.Err("no handler for page", zap.String("page", string(page)))
		}
	}

	//
	message.Ack()
}

type IntPayload struct {
	ID int `json:"id"`
}

type StringPayload struct {
	String string `json:"id"`
}

type StringsPayload struct {
	IDs []string `json:"id"`
}

type ChangesPayload struct {
	Data [][]interface{} `json:"d"`
}

type AdminPayload struct {
	TaskID string `json:"task_id"`
	Action string `json:"action"`
	Time   int64  `json:"time"`
}

type ChatBotPayload struct {
	RowData []interface{} `json:"row_data"`
}

type ChatPayload struct {
	I            float32 `json:"i"`
	AuthorID     string  `json:"author_id"`
	AuthorUser   string  `json:"author_user"`
	AuthorAvatar string  `json:"author_avatar"`
	Content      string  `json:"content"`
	Channel      string  `json:"channel"`
	Time         string  `json:"timestamp"`
	Embeds       bool    `json:"embeds"`
}

type PlayerPayload struct {
	ID            string `json:"id"` // string for js
	Name          string `json:"name"`
	Link          string `json:"link"`
	Avatar        string `json:"avatar"`
	CreatedAt     int64  `json:"created_at"`
	Queue         string `json:"queue"`
	CommunityLink string `json:"community_link"`
	New           bool   `json:"new"`
}

type NewsPayload struct {
	RowData []interface{} `json:"row_data"`
}
