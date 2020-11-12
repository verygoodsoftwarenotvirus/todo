package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"
)

const (
	// ItemAssignmentKey is the key we use to indicate that an audit log entry is associated with an item.
	ItemAssignmentKey = "item_id"
)

// BuildItemCreationEventEntry builds an entry creation input for when an item is created.
func BuildItemCreationEventEntry(item *models.Item) *models.AuditLogEntryCreationInput {
	return &models.AuditLogEntryCreationInput{
		EventType: ItemCreationEvent,
		Context: map[string]interface{}{
			ItemAssignmentKey:     item.ID,
			CreationAssignmentKey: item,
		},
	}
}

// BuildItemUpdateEventEntry builds an entry creation input for when an item is updated.
func BuildItemUpdateEventEntry(userID, itemID uint64, changes []models.FieldChangeSummary) *models.AuditLogEntryCreationInput {
	return &models.AuditLogEntryCreationInput{
		EventType: ItemUpdateEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:   userID,
			ItemAssignmentKey:    itemID,
			ChangesAssignmentKey: changes,
		},
	}
}

// BuildItemArchiveEventEntry builds an entry creation input for when an item is archived.
func BuildItemArchiveEventEntry(userID, itemID uint64) *models.AuditLogEntryCreationInput {
	return &models.AuditLogEntryCreationInput{
		EventType: ItemArchiveEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: userID,
			ItemAssignmentKey:  itemID,
		},
	}
}
