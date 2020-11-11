package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	// UserAssignmentKey is the key we use to indicate that an audit log entry is associated with a user.
	UserAssignmentKey = "user_id"
)

// BuildUserCreationEventEntry builds an entry creation input for when a user is created.
func BuildUserCreationEventEntry(user *models.User) *models.AuditLogEntryCreationInput {
	return &models.AuditLogEntryCreationInput{
		EventType: UserCreationEvent,
		Context: map[string]interface{}{
			UserAssignmentKey:     user.ID,
			CreationAssignmentKey: user,
		},
	}
}

// BuildUserVerifyTwoFactorSecretEventEntry builds an entry creation input for when a user verifies their two factor secret.
func BuildUserVerifyTwoFactorSecretEventEntry(userID uint64) *models.AuditLogEntryCreationInput {
	return &models.AuditLogEntryCreationInput{
		EventType: UserVerifyTwoFactorSecretEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: userID,
		},
	}
}

// BuildUserUpdateTwoFactorSecretEventEntry builds an entry creation input for when a user updates their two factor secret.
func BuildUserUpdateTwoFactorSecretEventEntry(userID uint64) *models.AuditLogEntryCreationInput {
	return &models.AuditLogEntryCreationInput{
		EventType: UserUpdateTwoFactorSecretEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: userID,
		},
	}
}

// BuildUserUpdatePasswordEventEntry builds an entry creation input for when a user updates their password.
func BuildUserUpdatePasswordEventEntry(userID uint64) *models.AuditLogEntryCreationInput {
	return &models.AuditLogEntryCreationInput{
		EventType: UserUpdatePasswordEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: userID,
		},
	}
}

// BuildUserArchiveEventEntry builds an entry creation input for when a user is archived.
func BuildUserArchiveEventEntry(userID uint64) *models.AuditLogEntryCreationInput {
	return &models.AuditLogEntryCreationInput{
		EventType: UserArchiveEvent,
		Context: map[string]interface{}{
			UserAssignmentKey:  userID,
			ActorAssignmentKey: userID,
		},
	}
}
