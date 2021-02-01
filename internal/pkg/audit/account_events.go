package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// AccountAssignmentKey is the key we use to indicate that an audit log entry is associated with an account.
	AccountAssignmentKey = "account_id"

	// AccountCreationEvent events indicate a user created an account.
	AccountCreationEvent = "account_created"
	// AccountUpdateEvent events indicate a user updated an account.
	AccountUpdateEvent = "account_updated"
	// AccountArchiveEvent events indicate a user deleted an account.
	AccountArchiveEvent = "account_archived"
)

// BuildAccountCreationEventEntry builds an entry creation input for when an account is created.
func BuildAccountCreationEventEntry(account *types.Account) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: AccountCreationEvent,
		Context: map[string]interface{}{
			AccountAssignmentKey:  account.ID,
			CreationAssignmentKey: account,
		},
	}
}

// BuildAccountUpdateEventEntry builds an entry creation input for when an account is updated.
func BuildAccountUpdateEventEntry(userID, accountID uint64, changes []types.FieldChangeSummary) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: AccountUpdateEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:   userID,
			AccountAssignmentKey: accountID,
			ChangesAssignmentKey: changes,
		},
	}
}

// BuildAccountArchiveEventEntry builds an entry creation input for when an account is archived.
func BuildAccountArchiveEventEntry(userID, accountID uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: AccountArchiveEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:   userID,
			AccountAssignmentKey: accountID,
		},
	}
}
