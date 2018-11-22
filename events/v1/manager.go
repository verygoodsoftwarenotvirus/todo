package events

// very much heavily borrowed from the great gorilla/websocket examples:
//		https://github.com/gorilla/websocket/tree/483fb8d7c32fcb4b5636cd293a92e3935932e2f4/examples/chat

import (
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second    // Time allowed to write a message to the peer.
	pongWait   = 60 * time.Second    // Time allowed to read the next pong message from the peer.
	pingPeriod = (pongWait * 9) / 10 // Send pings to peer with this period. Must be less than pongWait.
)

// EventListener is a middleman between the websocket connection and the EventHub.
type EventListener struct {
	eventHub         *EventHub
	conn             *websocket.Conn
	events           chan models.Event
	SubscribedTopics map[string]bool
	SubscribedTypes  []interface{}
}

func NewEventListener(ev *EventHub, conn *websocket.Conn) *EventListener {
	return &EventListener{
		eventHub:         ev,
		conn:             conn,
		events:           make(chan models.Event, models.DefaultLimit),
		SubscribedTopics: map[string]bool{},
		SubscribedTypes:  []interface{}{},
	}
}

// Serve serves messages from the EventHub to the websocket connection.
func (el *EventListener) Serve() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		el.conn.Close()
	}()
	for {
		select {
		case event, ok := <-el.events:
			if _, ok := el.SubscribedTopics[event.Event]; !ok {
				continue
			}
			var typeMatches bool
			for _, i := range el.SubscribedTypes {
				if i == event.Data {
					typeMatches = true
				}
			}
			if !typeMatches {
				continue
			}

			el.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				el.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := el.conn.WriteJSON(event); err != nil {
				return
			}

			for i := 0; i < len(el.events); i++ {
				if err := el.conn.WriteJSON(<-el.events); err != nil {
					return
				}
			}
		case <-ticker.C:
			el.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := el.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
