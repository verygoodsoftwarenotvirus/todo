package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// UserAssignmentKey is the key we use to indicate that an audit log entry is associated with a user.
	UserAssignmentKey = "user_id"

	// UserCreationEvent events indicate a user was created.
	UserCreationEvent = "user_created"
	// UserVerifyTwoFactorSecretEvent events indicate a user was created.
	UserVerifyTwoFactorSecretEvent = "user_two_factor_secret_verified"
	// UserUpdateTwoFactorSecretEvent events indicate a user updated their two factor secret.
	UserUpdateTwoFactorSecretEvent = "user_two_factor_secret_changed"
	// UserUpdateEvent events indicate a user was updated.
	UserUpdateEvent = "user_updated"
	// UserUpdatePasswordEvent events indicate a user updated their two factor secret.
	UserUpdatePasswordEvent = "user_password_updated"
	// UserArchiveEvent events indicate a user was archived.
	UserArchiveEvent = "user_archived"
)

// BuildUserCreationEventEntry builds an entry creation input for when a user is created.
func BuildUserCreationEventEntry(userID uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: UserCreationEvent,
		Context: map[string]interface{}{
			UserAssignmentKey: userID,
		},
	}
}

// BuildUserVerifyTwoFactorSecretEventEntry builds an entry creation input for when a user verifies their two factor secret.
func BuildUserVerifyTwoFactorSecretEventEntry(userID uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: UserVerifyTwoFactorSecretEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: userID,
		},
	}
}

// BuildUserUpdateTwoFactorSecretEventEntry builds an entry creation input for when a user updates their two factor secret.
func BuildUserUpdateTwoFactorSecretEventEntry(userID uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: UserUpdateTwoFactorSecretEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: userID,
		},
	}
}

// BuildUserUpdatePasswordEventEntry builds an entry creation input for when a user updates their authentication.
func BuildUserUpdatePasswordEventEntry(userID uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: UserUpdatePasswordEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: userID,
		},
	}
}

// BuildUserUpdateEventEntry builds an entry creation input for when a user is updated.
func BuildUserUpdateEventEntry(userID uint64, changes []types.FieldChangeSummary) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: UserUpdateEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:   userID,
			ChangesAssignmentKey: changes,
		},
	}
}

// BuildUserArchiveEventEntry builds an entry creation input for when a user is archived.
func BuildUserArchiveEventEntry(userID uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: UserArchiveEvent,
		Context: map[string]interface{}{
			UserAssignmentKey:  userID,
			ActorAssignmentKey: userID,
		},
	}
}
