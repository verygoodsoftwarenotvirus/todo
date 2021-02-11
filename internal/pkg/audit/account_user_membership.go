package audit

import "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

const (
	// UserAddedToAccountEvent events indicate a user created a membership.
	UserAddedToAccountEvent = "user_added_to_account"
	// AccountUserMembershipArchiveEvent events indicate a user deleted a membership.
	AccountUserMembershipArchiveEvent = "user_removed_from_account"
	// AccountMarkedAsDefaultEvent events indicate a user deleted a membership.
	AccountMarkedAsDefaultEvent = "account_marked_as_default"
)

// BuildUserAddedToAccountEventEntry builds an entry creation input for when a membership is created.
func BuildUserAddedToAccountEventEntry(addedBy, added, accountID uint64, reason string) *types.AuditLogEntryCreationInput {
	contextMap := map[string]interface{}{
		ActorAssignmentKey:   addedBy,
		AccountAssignmentKey: accountID,
		UserAssignmentKey:    added,
	}

	if reason != "" {
		contextMap[ReasonKey] = reason
	}

	return &types.AuditLogEntryCreationInput{
		EventType: UserAddedToAccountEvent,
		Context:   contextMap,
	}
}

// BuildUserRemovedFromAccountEventEntry builds an entry creation input for when a membership is archived.
func BuildUserRemovedFromAccountEventEntry(removedBy, removed, accountID uint64, reason string) *types.AuditLogEntryCreationInput {
	contextMap := map[string]interface{}{
		ActorAssignmentKey:   removedBy,
		AccountAssignmentKey: accountID,
		UserAssignmentKey:    removed,
	}

	if reason != "" {
		contextMap[ReasonKey] = reason
	}

	return &types.AuditLogEntryCreationInput{
		EventType: AccountUserMembershipArchiveEvent,
		Context:   contextMap,
	}
}

// BuildUserMarkedAccountAsDefaultEventEntry builds an entry creation input for when a membership is created.
func BuildUserMarkedAccountAsDefaultEventEntry(performedBy, userID, accountID uint64, reason string) *types.AuditLogEntryCreationInput {
	contextMap := map[string]interface{}{
		ActorAssignmentKey:   performedBy,
		UserAssignmentKey:    userID,
		AccountAssignmentKey: accountID,
	}

	if reason != "" {
		contextMap[ReasonKey] = reason
	}

	return &types.AuditLogEntryCreationInput{
		EventType: AccountMarkedAsDefaultEvent,
		Context:   contextMap,
	}
}
