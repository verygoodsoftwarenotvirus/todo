package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

const (
	// ItemAssignmentKey is the key we use to indicate that an audit log entry is associated with an item.
	ItemAssignmentKey = "item_id"

	// ItemCreationEvent events indicate a user created an item.
	ItemCreationEvent = "item_created"
	// ItemUpdateEvent events indicate a user updated an item.
	ItemUpdateEvent = "item_updated"
	// ItemArchiveEvent events indicate a user deleted an item.
	ItemArchiveEvent = "item_archived"
)

// BuildItemCreationEventEntry builds an entry creation input for when an item is created.
func BuildItemCreationEventEntry(item *types.Item, createdByUser uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: ItemCreationEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:    createdByUser,
			ItemAssignmentKey:     item.ID,
			CreationAssignmentKey: item,
			AccountAssignmentKey:  item.BelongsToAccount,
		},
	}
}

// BuildItemUpdateEventEntry builds an entry creation input for when an item is updated.
func BuildItemUpdateEventEntry(changedByUser, itemID, accountID uint64, changes []*types.FieldChangeSummary) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: ItemUpdateEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:   changedByUser,
			AccountAssignmentKey: accountID,
			ItemAssignmentKey:    itemID,
			ChangesAssignmentKey: changes,
		},
	}
}

// BuildItemArchiveEventEntry builds an entry creation input for when an item is archived.
func BuildItemArchiveEventEntry(archivedByUser, accountID, itemID uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: ItemArchiveEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:   archivedByUser,
			AccountAssignmentKey: accountID,
			ItemAssignmentKey:    itemID,
		},
	}
}
