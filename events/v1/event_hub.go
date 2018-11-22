package events

import (
	"sync"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models"
)

// EventHub maintains the set of active clients and broadcasts events to the clients.
type EventHub struct {
	lock sync.Mutex
	// Registered clients.
	clients map[*EventListener]bool

	// Inbound messages from the clients.
	newEvents chan models.Event

	eventsCache [models.DefaultLimit]models.Event

	// Unregister requests from clients.
	unregister chan *EventListener
}

func NewEventHub() *EventHub {
	h := &EventHub{
		newEvents: make(chan models.Event),
		clients:   make(map[*EventListener]bool),
	}
	go h.run()
	return h
}

func (h *EventHub) AddEvent(event string, data interface{}) {
	h.newEvents <- models.Event{Event: event, Data: data}
}

func (h *EventHub) AttachListener(el *EventListener) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.clients[el] = true
}

func (h *EventHub) DetachListener(el *EventListener) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if _, ok := h.clients[el]; ok {
		delete(h.clients, el)
		close(el.events)
	}
}

func (h *EventHub) SetCache(events []models.Event) {
	h.lock.Lock()
	defer h.lock.Unlock()
	newCache := [models.DefaultLimit]models.Event{}
	for i, x := range events {
		if i == models.DefaultLimit {
			break
		}
		newCache[i] = x
	}
	h.eventsCache = newCache
}

func (h *EventHub) run() {
	for {
		select {
		case event := <-h.newEvents:
			h.lock.Lock()
			for el := range h.clients {
				select {
				case el.events <- event:
				default:
					close(el.events)
					delete(h.clients, el)
				}
			}
			h.lock.Unlock()
		}
	}
}
