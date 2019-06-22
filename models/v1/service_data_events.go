package models

// ServiceDataEvent is a simple string alias
type ServiceDataEvent string

var (
	// Create represents a create event
	Create ServiceDataEvent = "create"
	// Update represents a update event
	Update ServiceDataEvent = "update"
	// Archive represents a delete event
	Archive ServiceDataEvent = "delete"
)
