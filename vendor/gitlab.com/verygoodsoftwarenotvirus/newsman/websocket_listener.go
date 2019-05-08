// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package newsman

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeTimeout = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = time.Minute

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = time.Minute / 2
)

// WebsocketListener is a middleman between the websocket connection and the newsman.
type WebsocketListener struct {
	newsman        *Newsman
	conn           WebsocketConnector
	incoming       chan Event
	listenerConfig *ListenerConfig
}

// NewWebsocketListener builds a websocket-based listener
func NewWebsocketListener(newsman *Newsman, conn WebsocketConnector, listenerConfig *ListenerConfig) Listener {
	wsl := &WebsocketListener{
		newsman:        newsman,
		conn:           conn,
		listenerConfig: listenerConfig,
		incoming:       make(chan Event, 256),
	}
	return wsl
}

// Listen sends messages from newsman to the websocket connection.
func (c *WebsocketListener) Listen() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if c.conn != nil {
			_ = c.conn.Close()
		}
	}()
	for {
		select {
		case event, ok := <-c.incoming:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if !ok {
				// The channel is closed.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				log.Println("websocket listener leaving audience: channel is closed")
				c.newsman.tuneOut <- c
				return
			}

			if err := c.conn.WriteJSON(event.Data); err != nil {
				log.Printf("websocket listener leaving audience: %v\n", err)
				c.newsman.tuneOut <- c
				return
			}

			// Add queued chat messages to the current websocket event.
			n := len(c.incoming)
			for i := 0; i < n; i++ {
				x := <-c.incoming
				if err := c.conn.WriteJSON(x.Data); err != nil {
					log.Printf("websocket listener leaving audience: %v\n", err)
					c.newsman.tuneOut <- c
					return
				}
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("websocket listener leaving audience: %v", err)
				c.newsman.tuneOut <- c
				return
			}
		}
	}
}

// Channel implements our listener interface
func (c *WebsocketListener) Channel() chan Event {
	return c.incoming
}

// Config implements our listener interface
func (c *WebsocketListener) Config() *ListenerConfig {
	return c.listenerConfig
}
