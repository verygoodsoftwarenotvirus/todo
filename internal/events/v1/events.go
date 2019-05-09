package v1

// Event is a simple string alias
type Event string

var (
	// Create represents a create event
	Create Event = "create"
	// Update represents a update event
	Update Event = "update"
	// Delete represents a delete event
	Delete Event = "delete"
)
