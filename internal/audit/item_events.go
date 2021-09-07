package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/segmentio/ksuid"
)

const (
	// ItemAssignmentKey is the key we use to indicate that an audit log entry is associated with an item.
	ItemAssignmentKey = "item_id"
	// ItemCreationEvent is the event type used to indicate an item was created.
	ItemCreationEvent = "item_created"
	// ItemUpdateEvent is the event type used to indicate an item was updated.
	ItemUpdateEvent = "item_updated"
	// ItemArchiveEvent is the event type used to indicate an item was archived.
	ItemArchiveEvent = "item_archived"
)

// BuildItemCreationEventEntry builds an entry creation input for when an item is created.
func BuildItemCreationEventEntry(item *types.Item, createdByUser string) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		ID:        ksuid.New().String(),
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
func BuildItemUpdateEventEntry(changedByUser, itemID, accountID string, changes []*types.FieldChangeSummary) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		ID:        ksuid.New().String(),
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
func BuildItemArchiveEventEntry(archivedByUser, accountID, itemID string) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		ID:        ksuid.New().String(),
		EventType: ItemArchiveEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:   archivedByUser,
			AccountAssignmentKey: accountID,
			ItemAssignmentKey:    itemID,
		},
	}
}
