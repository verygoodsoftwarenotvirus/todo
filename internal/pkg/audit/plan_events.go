package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// PlanAssignmentKey is the key we use to indicate that an audit log entry is associated with an plan.
	PlanAssignmentKey = "plan_id"
)

// BuildPlanCreationEventEntry builds an entry creation input for when an plan is created.
func BuildPlanCreationEventEntry(plan *types.Plan) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: PlanCreationEvent,
		Context: map[string]interface{}{
			PlanAssignmentKey:     plan.ID,
			CreationAssignmentKey: plan,
		},
	}
}

// BuildPlanUpdateEventEntry builds an entry creation input for when an plan is updated.
func BuildPlanUpdateEventEntry(userID, planID uint64, changes []types.FieldChangeSummary) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: PlanUpdateEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:   userID,
			PlanAssignmentKey:    planID,
			ChangesAssignmentKey: changes,
		},
	}
}

// BuildPlanArchiveEventEntry builds an entry creation input for when an plan is archived.
func BuildPlanArchiveEventEntry(userID, planID uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: PlanArchiveEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: userID,
			PlanAssignmentKey:  planID,
		},
	}
}
