package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
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
func BuildItemCreationEventEntry(item *types.Item) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: ItemCreationEvent,
		Context: map[string]interface{}{
			ItemAssignmentKey:     item.ID,
			CreationAssignmentKey: item,
		},
	}
}

// BuildItemUpdateEventEntry builds an entry creation input for when an item is updated.
func BuildItemUpdateEventEntry(userID, itemID uint64, changes []types.FieldChangeSummary) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: ItemUpdateEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:   userID,
			ItemAssignmentKey:    itemID,
			ChangesAssignmentKey: changes,
		},
	}
}

// BuildItemArchiveEventEntry builds an entry creation input for when an item is archived.
func BuildItemArchiveEventEntry(userID, itemID uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: ItemArchiveEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: userID,
			ItemAssignmentKey:  itemID,
		},
	}
}
