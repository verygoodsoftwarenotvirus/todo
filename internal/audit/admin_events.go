package audit

import "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

// BuildUserBanEventEntry builds an entry creation input for when a user is banned.
func BuildUserBanEventEntry(banGiver, banRecipient uint64, reason string) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: UserBannedEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: banGiver,
			UserAssignmentKey:  banRecipient,
			ReasonKey:          reason,
		},
	}
}

// BuildAccountTerminationEventEntry builds an entry creation input for when an account is terminated.
func BuildAccountTerminationEventEntry(terminator, terminee uint64, reason string) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: AccountTerminatedEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: terminator,
			UserAssignmentKey:  terminee,
			ReasonKey:          reason,
		},
	}
}
