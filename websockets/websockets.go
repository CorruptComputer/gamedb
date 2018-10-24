package websockets

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"strings"

	"github.com/gamedb/website/logging"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
)

const (
	PageChanges = "changes"
	PageChat    = "chat"
	PageNews    = "news"
	PagePrices  = "prices"
	PageProfile = "profile"
)

var pages map[string]Page

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var ErrInvalidPage = errors.New("invalid page")

func init() {
	pages = map[string]Page{}
	for _, v := range []string{PageChanges, PageChat, PageNews, PagePrices, PageProfile} {
		pages[v] = Page{
			name:        v,
			connections: map[int]*websocket.Conn{},
		}
	}
}

func WebsocketsHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	page, err := GetPage(id)
	if err != nil {

		bytes, err := json.Marshal(websocketPayload{Error: "Invalid page"})
		logging.Error(err)

		w.Write(bytes)
		return
	}

	// Upgrade the connection
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if !strings.Contains(err.Error(), "websocket: not a websocket handshake") {
			logging.Error(err)
		}
		return
	}

	page.SetConnection(connection)
}

func GetPage(page string) (p Page, err error) {

	if val, ok := pages[page]; ok {
		return val, nil
	}
	return p, ErrInvalidPage
}

type Page struct {
	name        string
	connections map[int]*websocket.Conn
}

func (p Page) HasConnections() bool {
	return len(p.connections) > 0
}

func (p *Page) SetConnection(conn *websocket.Conn) {
	p.connections[rand.Int()] = conn
}

func (p *Page) Send(data interface{}) {

	if !p.HasConnections() {
		return
	}

	payload := websocketPayload{}
	payload.Page = p.name
	payload.Data = data

	for k, v := range p.connections {
		err := v.WriteJSON(payload)
		if err != nil {

			// Clean up old connections
			if strings.Contains(err.Error(), "broken pipe") {

				v.Close()
				delete(p.connections, k)

			} else {
				logging.Error(err)
			}
		}
	}
}

type websocketPayload struct {
	Data  interface{}
	Page  string
	Error string
}
