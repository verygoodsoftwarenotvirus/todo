package audit

import "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

const (
	// AccountUserMembershipAssignmentKey is the key we use to indicate that an audit log entry is associated with an item.
	AccountUserMembershipAssignmentKey = "item_id"

	// AccountUserMembershipCreationEvent events indicate a user created an item.
	AccountUserMembershipCreationEvent = "item_created"
	// AccountUserMembershipUpdateEvent events indicate a user updated an item.
	AccountUserMembershipUpdateEvent = "item_updated"
	// AccountUserMembershipArchiveEvent events indicate a user deleted an item.
	AccountUserMembershipArchiveEvent = "item_archived"
)

// BuildAccountUserMembershipCreationEventEntry builds an entry creation input for when an item is created.
func BuildAccountUserMembershipCreationEventEntry(item *types.AccountUserMembership) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: AccountUserMembershipCreationEvent,
		Context: map[string]interface{}{
			AccountUserMembershipAssignmentKey: item.ID,
			CreationAssignmentKey:              item,
		},
	}
}

// BuildAccountUserMembershipUpdateEventEntry builds an entry creation input for when an item is updated.
func BuildAccountUserMembershipUpdateEventEntry(userID, itemID uint64, changes []types.FieldChangeSummary) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: AccountUserMembershipUpdateEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:                 userID,
			AccountUserMembershipAssignmentKey: itemID,
			ChangesAssignmentKey:               changes,
		},
	}
}

// BuildAccountUserMembershipArchiveEventEntry builds an entry creation input for when an item is archived.
func BuildAccountUserMembershipArchiveEventEntry(userID, itemID uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: AccountUserMembershipArchiveEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:                 userID,
			AccountUserMembershipAssignmentKey: itemID,
		},
	}
}
