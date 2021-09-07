package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/segmentio/ksuid"
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
func BuildAccountCreationEventEntry(account *types.Account, createdByUser string) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		ID:        ksuid.New().String(),
		EventType: AccountCreationEvent,
		Context: map[string]interface{}{
			AccountAssignmentKey:  account.ID,
			UserAssignmentKey:     account.BelongsToUser,
			ActorAssignmentKey:    createdByUser,
			CreationAssignmentKey: account,
		},
	}
}

// BuildAccountUpdateEventEntry builds an entry creation input for when an account is updated.
func BuildAccountUpdateEventEntry(userID, accountID, changedByUser string, changes []*types.FieldChangeSummary) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		ID:        ksuid.New().String(),
		EventType: AccountUpdateEvent,
		Context: map[string]interface{}{
			UserAssignmentKey:    userID,
			AccountAssignmentKey: accountID,
			ChangesAssignmentKey: changes,
			ActorAssignmentKey:   changedByUser,
		},
	}
}

// BuildAccountArchiveEventEntry builds an entry creation input for when an account is archived.
func BuildAccountArchiveEventEntry(userID, accountID, archivedByUser string) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		ID:        ksuid.New().String(),
		EventType: AccountArchiveEvent,
		Context: map[string]interface{}{
			UserAssignmentKey:    userID,
			AccountAssignmentKey: accountID,
			ActorAssignmentKey:   archivedByUser,
		},
	}
}
