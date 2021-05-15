package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

const (
	// UserAddedToAccountEvent events indicate a user created a membership.
	UserAddedToAccountEvent = "user_added_to_account"
	// UserAccountPermissionsModifiedEvent events indicate a user created a membership.
	UserAccountPermissionsModifiedEvent = "user_account_permissions_modified"
	// UserRemovedFromAccountEvent events indicate a user deleted a membership.
	UserRemovedFromAccountEvent = "user_removed_from_account"
	// AccountMarkedAsDefaultEvent events indicate a user deleted a membership.
	AccountMarkedAsDefaultEvent = "account_marked_as_default"
	// AccountTransferredEvent events indicate a user deleted a membership.
	AccountTransferredEvent = "account_transferred"
)

// BuildUserAddedToAccountEventEntry builds an entry creation input for when a membership is created.
func BuildUserAddedToAccountEventEntry(addedBy uint64, input *types.AddUserToAccountInput) *types.AuditLogEntryCreationInput {
	contextMap := map[string]interface{}{
		ActorAssignmentKey:   addedBy,
		AccountAssignmentKey: input.AccountID,
		UserAssignmentKey:    input.UserID,
		PermissionsKey:       input.UserAccountPermissions,
	}

	if input.Reason != "" {
		contextMap[ReasonKey] = input.Reason
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
func BuildModifyUserPermissionsEventEntry(userID, accountID, modifiedBy uint64, newPermissions permissions.ServiceUserPermission, reason string) *types.AuditLogEntryCreationInput {
	contextMap := map[string]interface{}{
		ActorAssignmentKey:   modifiedBy,
		AccountAssignmentKey: accountID,
		UserAssignmentKey:    userID,
		PermissionsKey:       newPermissions,
	}

	if reason != "" {
		contextMap[ReasonKey] = reason
	}

	return &types.AuditLogEntryCreationInput{
		EventType: UserAccountPermissionsModifiedEvent,
		Context:   contextMap,
	}
}

// BuildTransferAccountOwnershipEventEntry builds an entry creation input for when a membership is created.
func BuildTransferAccountOwnershipEventEntry(accountID, changedBy uint64, input *types.AccountOwnershipTransferInput) *types.AuditLogEntryCreationInput {
	contextMap := map[string]interface{}{
		ActorAssignmentKey:   changedBy,
		"old_owner":          input.CurrentOwner,
		"new_owner":          input.NewOwner,
		AccountAssignmentKey: accountID,
	}

	if input.Reason != "" {
		contextMap[ReasonKey] = input.Reason
	}

	return &types.AuditLogEntryCreationInput{
		EventType: AccountTransferredEvent,
		Context:   contextMap,
	}
}
