package items

import (
	"log"
	"net/http"
	"sync"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second    // Time allowed to write a message to the peer.
	pongWait   = 60 * time.Second    // Time allowed to read the next pong message from the peer.
	pingPeriod = (pongWait * 9) / 10 // Send pings to peer with this period. Must be less than pongWait.
)

// client is a middleman between the websocket connection and the ItemHub.
type client struct {
	itemHub *ItemHub
	conn    *websocket.Conn
	items   chan models.Item
}

// serveMessages serves messages from the ItemHub to the websocket connection.
func (c *client) serveMessages() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case item, ok := <-c.items:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(item); err != nil {
				return
			}

			for i := 0; i < len(c.items); i++ {
				if err := c.conn.WriteJSON(<-c.items); err != nil {
					return
				}
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWebSocket handles websocket requests from the peer.
func (is *ItemsService) Feed(res http.ResponseWriter, req *http.Request) {
	conn, err := is.upgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Println(err)
		return
	}
	cli := &client{itemHub: is.itemHub, conn: conn, items: make(chan models.Item, 32)}
	cli.itemHub.newClients <- cli

	go cli.serveMessages()
}

// ItemHub maintains the set of active clients and broadcasts messages to the
// clients.
type ItemHub struct {
	// Registered clients.
	clients    map[*client]bool
	clientLock sync.Mutex

	// Inbound messages from the clients.
	newItems chan models.Item

	itemsCache []models.Item

	// Register requests from the clients.
	newClients chan *client

	// Unregister requests from clients.
	unregister chan *client
}

func newItemHub() *ItemHub {
	h := &ItemHub{
		newItems:   make(chan models.Item),
		newClients: make(chan *client),
		unregister: make(chan *client),
		clients:    make(map[*client]bool),
	}
	go h.run()
	return h
}

func (h *ItemHub) AddItem(item models.Item) {
	h.newItems <- item
}

func (h *ItemHub) SetCache(items []models.Item) {
	h.clientLock.Lock()
	defer h.clientLock.Unlock()
	h.itemsCache = items
}

func (h *ItemHub) run() {
	for {
		select {
		case c := <-h.newClients:
			h.clientLock.Lock()
			h.clients[c] = true
			// for _, item := range h.itemsCache {
			// 	c.items <- item
			// }
			h.clientLock.Unlock()
		case c := <-h.unregister:
			h.clientLock.Lock()
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.items)
			}
			h.clientLock.Unlock()
		case item := <-h.newItems:
			h.clientLock.Lock()
			for c := range h.clients {
				select {
				case c.items <- item:
				default:
					close(c.items)
					delete(h.clients, c)
				}
			}
			h.clientLock.Unlock()
		}
	}
}
