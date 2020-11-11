package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	// ActorAssignmentKey is the key we use to indicate which user performed a given action.
	ActorAssignmentKey = "performed_by"
	// ChangesAssignmentKey is the key we use to indicate which changes occurred during an update.
	ChangesAssignmentKey = "changes"
	// CreationAssignmentKey is the key we use to indicate which object was created for creation events.
	CreationAssignmentKey = "created"
)

// BuildCycleCookieSecretEvent builds an entry creation input for when a cookie secret is cycled.
func BuildCycleCookieSecretEvent(userID uint64) *models.AuditLogEntryCreationInput {
	return &models.AuditLogEntryCreationInput{
		EventType: CycleCookieSecretEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: userID,
		},
	}
}

// BuildSuccessfulLoginEventEntry builds an entry creation input for when a user successfully logs in.
func BuildSuccessfulLoginEventEntry(userID uint64) *models.AuditLogEntryCreationInput {
	return &models.AuditLogEntryCreationInput{
		EventType: SuccessfulLoginEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: userID,
		},
	}
}

// BuildUnsuccessfulLoginBadPasswordEventEntry builds an entry creation input for when a user fails to log in because of a bad password.
func BuildUnsuccessfulLoginBadPasswordEventEntry(userID uint64) *models.AuditLogEntryCreationInput {
	return &models.AuditLogEntryCreationInput{
		EventType: UnsuccessfulLoginBadPasswordEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: userID,
		},
	}
}

// BuildUnsuccessfulLoginBad2FATokenEventEntry builds an entry creation input for when a user fails to log in because of a bad two factor token.
func BuildUnsuccessfulLoginBad2FATokenEventEntry(userID uint64) *models.AuditLogEntryCreationInput {
	return &models.AuditLogEntryCreationInput{
		EventType: UnsuccessfulLoginBad2FATokenEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: userID,
		},
	}
}

// BuildLogoutEventEntry builds an entry creation input for when a user logs out.
func BuildLogoutEventEntry(userID uint64) *models.AuditLogEntryCreationInput {
	return &models.AuditLogEntryCreationInput{
		EventType: LogoutEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: userID,
		},
	}
}
