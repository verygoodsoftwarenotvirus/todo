package models

// ServiceDataEvent is a simple string alias
type ServiceDataEvent string

const (
	// Create represents a create event
	Create ServiceDataEvent = "create"
	// Update represents an update event
	Update ServiceDataEvent = "update"
	// Archive represents an archive event
	Archive ServiceDataEvent = "archive"
)
