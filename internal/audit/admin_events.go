package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/segmentio/ksuid"
)

// BuildUserBanEventEntry builds an entry creation input for when a user is banned.
func BuildUserBanEventEntry(banGiver, banRecipient, reason string) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		ID:        ksuid.New().String(),
		EventType: UserBannedEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: banGiver,
			UserAssignmentKey:  banRecipient,
			ReasonKey:          reason,
		},
	}
}

// BuildAccountTerminationEventEntry builds an entry creation input for when an account is terminated.
func BuildAccountTerminationEventEntry(terminator, terminee, reason string) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		ID:        ksuid.New().String(),
		EventType: AccountTerminatedEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: terminator,
			UserAssignmentKey:  terminee,
			ReasonKey:          reason,
		},
	}
}
