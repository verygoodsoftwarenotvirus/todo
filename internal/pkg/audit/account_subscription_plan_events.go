package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// AccountSubscriptionPlanAssignmentKey is the key we use to indicate that an audit log entry is associated with an plan.
	AccountSubscriptionPlanAssignmentKey = "plan_id"

	// AccountSubscriptionPlanCreationEvent events indicate a user created a plan.
	AccountSubscriptionPlanCreationEvent = "plan_created"
	// AccountSubscriptionPlanUpdateEvent events indicate a user updated a plan.
	AccountSubscriptionPlanUpdateEvent = "plan_updated"
	// AccountSubscriptionPlanArchiveEvent events indicate a user deleted a plan.
	AccountSubscriptionPlanArchiveEvent = "plan_archived"
)

// BuildAccountSubscriptionPlanCreationEventEntry builds an entry creation input for when an plan is created.
func BuildAccountSubscriptionPlanCreationEventEntry(plan *types.AccountSubscriptionPlan) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: AccountSubscriptionPlanCreationEvent,
		Context: map[string]interface{}{
			AccountSubscriptionPlanAssignmentKey: plan.ID,
			CreationAssignmentKey:                plan,
		},
	}
}

// BuildAccountSubscriptionPlanUpdateEventEntry builds an entry creation input for when an plan is updated.
func BuildAccountSubscriptionPlanUpdateEventEntry(userID, planID uint64, changes []types.FieldChangeSummary) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: AccountSubscriptionPlanUpdateEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:                   userID,
			AccountSubscriptionPlanAssignmentKey: planID,
			ChangesAssignmentKey:                 changes,
		},
	}
}

// BuildAccountSubscriptionPlanArchiveEventEntry builds an entry creation input for when an plan is archived.
func BuildAccountSubscriptionPlanArchiveEventEntry(userID, planID uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: AccountSubscriptionPlanArchiveEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:                   userID,
			AccountSubscriptionPlanAssignmentKey: planID,
		},
	}
}
