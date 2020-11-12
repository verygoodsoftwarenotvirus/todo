package converters

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"
)

// ConvertAuditLogEntryCreationInputToEntry converts an AuditLogEntryCreationInput to an AuditLogEntry.
func ConvertAuditLogEntryCreationInputToEntry(e *models.AuditLogEntryCreationInput) *models.AuditLogEntry {
	return &models.AuditLogEntry{
		EventType: e.EventType,
		Context:   e.Context,
	}
}
