package pages

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

func WebsocketsRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/{id:[a-z-]+}", websocketsHandler)
	return r
}

func websocketsHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	page := websockets.GetPage(websockets.WebsocketPage(id))
	if page == nil {
		returnErrorTemplate(w, r, errorTemplate{Message: "Invalid websocket ID", Code: 404})
		return
	}
	if page.GetName() == "" {

		bytes, err := json.Marshal(websockets.WebsocketPayload{Error: "Invalid page"})
		if err != nil {
			log.ErrS(err)
		}

		_, err = w.Write(bytes)
		if err != nil {
			log.ErrS(err)
		}
		return
	}

	// Upgrade the connection
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if strings.Contains(err.Error(), "websocket: not a websocket handshake") {
			return
		}
		if strings.Contains(err.Error(), "'websocket' token not found in 'Upgrade'") {
			return
		}
		log.ErrS(err)
		return
	}

	page.AddConnection(connection)
}
