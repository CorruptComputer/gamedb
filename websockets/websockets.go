package websockets

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gamedb/website/log"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
)

type WebsocketPage string

const (
	PageChanges  WebsocketPage = "changes"
	PageChat     WebsocketPage = "chat"
	PageNews     WebsocketPage = "news"
	PagePrices   WebsocketPage = "prices"
	PageAdmin    WebsocketPage = "admin"
	PageApp      WebsocketPage = "app"
	PagePackage  WebsocketPage = "package"
	PagePackages WebsocketPage = "packages"
	PageProfile  WebsocketPage = "profile"
)

var (
	pages      map[WebsocketPage]Page
	pagesSlice = []WebsocketPage{PageChanges, PageChat, PageNews, PagePrices, PageAdmin, PageApp, PagePackage, PagePackages, PageProfile}
	upgrader   = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

var ErrInvalidPage = errors.New("invalid page")

func init() {
	pages = map[WebsocketPage]Page{}
	for _, v := range pagesSlice {
		pages[v] = Page{
			name:        v,
			connections: map[uuid.UUID]*websocket.Conn{},
		}
	}
}

func WebsocketsHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	page, err := GetPage(WebsocketPage(id))
	if err != nil {

		bytes, err := json.Marshal(websocketPayload{Error: "Invalid page"})
		log.Err(err)

		_, err = w.Write(bytes)
		log.Err(err)
		return
	}

	// Upgrade the connection
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if !strings.Contains(err.Error(), "websocket: not a websocket handshake") {
			log.Err(err)
		}
		return
	}

	err = page.setConnection(connection)
	if err != nil {
		log.Err(err)
	}
}

func GetPage(page WebsocketPage) (p Page, err error) {

	if val, ok := pages[page]; ok {
		return val, nil
	}
	return p, ErrInvalidPage
}

type Page struct {
	name        WebsocketPage
	connections map[uuid.UUID]*websocket.Conn
}

func (p Page) HasConnections() bool {
	return len(p.connections) > 0
}

func (p *Page) setConnection(conn *websocket.Conn) error {

	id, err := uuid.NewV4()
	if err != nil {
		return err
	}

	p.connections[id] = conn

	return nil
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

				err := v.Close()
				log.Err(err)
				delete(p.connections, k)

			} else {
				log.Err(err)
			}
		}
	}
}

type websocketPayload struct {
	Data  interface{}
	Page  WebsocketPage
	Error string
}
