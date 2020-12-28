package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// PlanAssignmentKey is the key we use to indicate that an audit log entry is associated with an item.
	PlanAssignmentKey = "item_id"
)

// BuildPlanCreationEventEntry builds an entry creation input for when an item is created.
func BuildPlanCreationEventEntry(item *types.Plan) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: PlanCreationEvent,
		Context: map[string]interface{}{
			PlanAssignmentKey:     item.ID,
			CreationAssignmentKey: item,
		},
	}
}

// BuildPlanUpdateEventEntry builds an entry creation input for when an item is updated.
func BuildPlanUpdateEventEntry(userID, itemID uint64, changes []types.FieldChangeSummary) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: PlanUpdateEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:   userID,
			PlanAssignmentKey:    itemID,
			ChangesAssignmentKey: changes,
		},
	}
}

// BuildPlanArchiveEventEntry builds an entry creation input for when an item is archived.
func BuildPlanArchiveEventEntry(userID, itemID uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: PlanArchiveEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: userID,
			PlanAssignmentKey:  itemID,
		},
	}
}
