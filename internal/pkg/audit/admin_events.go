package audit

import "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

// BuildUserBanEventEntry builds an entry creation input for when an item is created.
func BuildUserBanEventEntry(banGiver, banRecipient uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: UserBannedEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: banGiver,
			UserAssignmentKey:  banRecipient,
		},
	}
}
