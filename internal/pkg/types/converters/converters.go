package converters

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// ConvertAuditLogEntryCreationInputToEntry converts an AuditLogEntryCreationInput to an AuditLogEntry.
func ConvertAuditLogEntryCreationInputToEntry(e *types.AuditLogEntryCreationInput) *types.AuditLogEntry {
	return &types.AuditLogEntry{
		EventType: e.EventType,
		Context:   e.Context,
	}
}
