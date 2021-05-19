package converters

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

// ConvertAuditLogEntryCreationInputToEntry converts an AuditLogEntryCreationInput to an AuditLogEntry.
func ConvertAuditLogEntryCreationInputToEntry(e *types.AuditLogEntryCreationInput) *types.AuditLogEntry {
	return &types.AuditLogEntry{
		EventType: e.EventType,
		Context:   e.Context,
	}
}

// ConvertAccountToAccountUpdateInput creates an AccountUpdateInput struct from an item.
func ConvertAccountToAccountUpdateInput(x *types.Account) *types.AccountUpdateInput {
	return &types.AccountUpdateInput{
		Name:          x.Name,
		BelongsToUser: x.BelongsToUser,
	}
}

// ConvertItemToItemUpdateInput creates an ItemUpdateInput struct from an item.
func ConvertItemToItemUpdateInput(x *types.Item) *types.ItemUpdateInput {
	return &types.ItemUpdateInput{
		Name:    x.Name,
		Details: x.Details,
	}
}
