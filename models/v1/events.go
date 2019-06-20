package models

// Event is a simple string alias
type Event string

var (
	// Create represents a create event
	Create Event = "create"
	// Update represents a update event
	Update Event = "update"
	// Archive represents a delete event
	Archive Event = "delete"
)
