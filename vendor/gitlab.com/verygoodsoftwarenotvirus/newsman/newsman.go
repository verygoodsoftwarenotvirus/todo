// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package newsman

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebsocketAuthFunc is a function newsman will call when it receives
// a websocket connection request. The onus is on the service using
// newsman to provide a valid authentication func. The default behavior
// is openly authorized
type WebsocketAuthFunc func(req *http.Request) bool

// TypeNameManipulationFunc is a function newsman will call when it is asked
// to report on something. It will determine the type's name as a string,
// call this function with that derived type name, and use the return value
// (after removing any errant asterisks) as the type name for data type notifications
type TypeNameManipulationFunc func(string) string

type Reporter interface {
	Report(event Event)
}

// Newsman maintains the set of active audience and broadcasts messages to the
// audience.
type Newsman struct {
	audienceMutex sync.RWMutex
	audience      map[Listener]bool

	events chan Event

	// How to start listening
	listeners chan Listener

	// How to stop listening
	tuneOut chan Listener

	upgrader *websocket.Upgrader

	websocketAuthFunc        WebsocketAuthFunc
	typeNameManipulationFunc TypeNameManipulationFunc
}

// NewNewsman builds a Newsman
func NewNewsman(
	websocketAuthFunc WebsocketAuthFunc,
	typeNameManipulationFunc TypeNameManipulationFunc,
) *Newsman {
	x := &Newsman{
		events:                   make(chan Event),
		listeners:                make(chan Listener),
		tuneOut:                  make(chan Listener),
		audience:                 make(map[Listener]bool),
		websocketAuthFunc:        websocketAuthFunc,
		typeNameManipulationFunc: typeNameManipulationFunc,
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}

	go x.run()
	return x
}

// TuneIn provides a convenient way to tune in to the newsman
func (newsman *Newsman) TuneIn(l Listener) {
	newsman.listeners <- l
}

// TuneOut provides a convenient way to disconnect from the newsman
func (newsman *Newsman) TuneOut(l Listener) {
	newsman.tuneOut <- l
}

// Report provides a convenient way to Report events to the newsman
func (newsman *Newsman) Report(event Event) {
	newsman.events <- event
}

// ServeWebsockets serves websockets to those who want them
func (newsman *Newsman) ServeWebsockets(res http.ResponseWriter, req *http.Request) {
	if newsman.websocketAuthFunc != nil {
		authorized := newsman.websocketAuthFunc(req)
		if !authorized {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	conn, err := newsman.upgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Println(err)
		return
	}

	cfg := ParseConfigFromURL(req.URL.Query())
	newsman.TuneIn(NewWebsocketListener(newsman, conn, cfg))
}

func (newsman *Newsman) run() {
	for {
		select {
		case listener := <-newsman.listeners:
			// log.Printf("listeners channel received: %v\n", listener)

			go listener.Listen()
			newsman.audienceMutex.Lock()
			newsman.audience[listener] = true
			newsman.audienceMutex.Unlock()

		case listener := <-newsman.tuneOut:
			// log.Printf("tuneOut channel received: %v\n", listener)

			newsman.audienceMutex.Lock()
			if _, ok := newsman.audience[listener]; ok {
				delete(newsman.audience, listener)

				close(listener.Channel())
			}
			newsman.audienceMutex.Unlock()
		case event := <-newsman.events:
			// log.Printf("events channel received: %v\n", event)

			newsman.audienceMutex.RLock()
			for listener := range newsman.audience {
				if listener.Config().IsInterested(event, newsman.typeNameManipulationFunc) {
					select {
					case listener.Channel() <- event:
					default:
						close(listener.Channel())
						delete(newsman.audience, listener)
					}
				}
			}
			newsman.audienceMutex.RUnlock()
		}
	}
}

// AudienceCount returns the current audience size
func (newsman *Newsman) AudienceCount() uint {
	newsman.audienceMutex.RLock()
	defer newsman.audienceMutex.RUnlock()
	return uint(len(newsman.audience))
}
