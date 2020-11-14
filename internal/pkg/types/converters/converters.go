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

// ConvertItemToItemUpdateInput creates an ItemUpdateInput struct from an item.
func ConvertItemToItemUpdateInput(x *types.Item) *types.ItemUpdateInput {
	return &types.ItemUpdateInput{
		Name:    x.Name,
		Details: x.Details,
	}
}
