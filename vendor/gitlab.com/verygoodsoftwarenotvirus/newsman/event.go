package newsman

// Event represents something worth notifying a websocket or config about
type Event struct {
	EventType string      `json:"event_type"`
	Data      interface{} `json:"data"`
	Topics    []string    `json:"topics"`
}
