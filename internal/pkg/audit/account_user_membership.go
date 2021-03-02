package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions/bitmask"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// UserAddedToAccountEvent events indicate a user created a membership.
	UserAddedToAccountEvent = "user_added_to_account"
	// UserRemovedFromAccountEvent events indicate a user deleted a membership.
	UserRemovedFromAccountEvent = "user_removed_from_account"
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
		EventType: UserRemovedFromAccountEvent,
		Context:   contextMap,
	}
}

// BuildUserMarkedAccountAsDefaultEventEntry builds an entry creation input for when a membership is created.
func BuildUserMarkedAccountAsDefaultEventEntry(performedBy, userID, accountID uint64) *types.AuditLogEntryCreationInput {
	contextMap := map[string]interface{}{
		ActorAssignmentKey:   performedBy,
		UserAssignmentKey:    userID,
		AccountAssignmentKey: accountID,
	}

	return &types.AuditLogEntryCreationInput{
		EventType: AccountMarkedAsDefaultEvent,
		Context:   contextMap,
	}
}

// BuildModifyUserPermissionsEventEntry builds an entry creation input for when a membership is created.
func BuildModifyUserPermissionsEventEntry(userID, accountID, modifiedBy uint64, newPermissions bitmask.ServiceUserPermissions, reason string) *types.AuditLogEntryCreationInput {
	contextMap := map[string]interface{}{
		ActorAssignmentKey:   modifiedBy,
		AccountAssignmentKey: accountID,
		UserAssignmentKey:    userID,
		PermissionsKey:       newPermissions,
		ReasonKey:            reason,
	}

	if reason != "" {
		contextMap[ReasonKey] = reason
	}

	return &types.AuditLogEntryCreationInput{
		EventType: UserAddedToAccountEvent,
		Context:   contextMap,
	}
}

// BuildTransferAccountOwnershipEventEntry builds an entry creation input for when a membership is created.
func BuildTransferAccountOwnershipEventEntry(oldOwner, newOwner, changedBy, accountID uint64, reason string) *types.AuditLogEntryCreationInput {
	contextMap := map[string]interface{}{
		AccountAssignmentKey: accountID,
	}

	if reason != "" {
		contextMap[ReasonKey] = reason
	}

	return &types.AuditLogEntryCreationInput{
		EventType: UserAddedToAccountEvent,
		Context:   contextMap,
	}
}
