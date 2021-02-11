package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// ActorAssignmentKey is the key we use to indicate which user performed a given action.
	ActorAssignmentKey = "performed_by"
	// ChangesAssignmentKey is the key we use to indicate which changes occurred during an update.
	ChangesAssignmentKey = "changes"
	// CreationAssignmentKey is the key we use to indicate which object was created for creation events.
	CreationAssignmentKey = "created"
	// PermissionsKey is the key we use to indicate which permissions were applicable to an event.
	PermissionsKey = "permissions"
	// ReasonKey is the key we use to indicate the reason behind a given event.
	ReasonKey = "reason"

	// UserBannedEvent events indicate an admin cycled the cookie secret.
	UserBannedEvent = "user_banned"
	// AccountTerminatedEvent events indicate an admin cycled the cookie secret.
	AccountTerminatedEvent = "account_terminated"
	// CycleCookieSecretEvent events indicate an admin cycled the cookie secret.
	CycleCookieSecretEvent = "cookie_secret_cycled"
	// SuccessfulLoginEvent events indicate a user successfully authenticated into the service via username + authentication + 2fa.
	SuccessfulLoginEvent = "user_logged_in"
	// LogoutEvent events indicate a user successfully logged out.
	LogoutEvent = "user_logged_out"
	// BannedUserLoginAttemptEvent events indicate a user successfully authenticated into the service via username + authentication + 2fa.
	BannedUserLoginAttemptEvent = "banned_user_login_attempt"
	// UnsuccessfulLoginBadPasswordEvent events indicate a user attempted to authenticate into the service, but failed because of an invalid authentication.
	UnsuccessfulLoginBadPasswordEvent = "user_login_failed_bad_password"
	// UnsuccessfulLoginBad2FATokenEvent events indicate a user attempted to authenticate into the service, but failed because of a faulty two factor token.
	UnsuccessfulLoginBad2FATokenEvent = "user_login_failed_bad_2FA_token"
)

// BuildCycleCookieSecretEvent builds an entry creation input for when a cookie secret is cycled.
func BuildCycleCookieSecretEvent(userID uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: CycleCookieSecretEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: userID,
		},
	}
}

// BuildSuccessfulLoginEventEntry builds an entry creation input for when a user successfully logs in.
func BuildSuccessfulLoginEventEntry(userID uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: SuccessfulLoginEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: userID,
		},
	}
}

// BuildBannedUserLoginAttemptEventEntry builds an entry creation input for when a user successfully logs in.
func BuildBannedUserLoginAttemptEventEntry(userID uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: BannedUserLoginAttemptEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: userID,
		},
	}
}

// BuildUnsuccessfulLoginBadPasswordEventEntry builds an entry creation input for when a user fails to log in because of a bad authentication.
func BuildUnsuccessfulLoginBadPasswordEventEntry(userID uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: UnsuccessfulLoginBadPasswordEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: userID,
		},
	}
}

// BuildUnsuccessfulLoginBad2FATokenEventEntry builds an entry creation input for when a user fails to log in because of a bad two factor token.
func BuildUnsuccessfulLoginBad2FATokenEventEntry(userID uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: UnsuccessfulLoginBad2FATokenEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: userID,
		},
	}
}

// BuildLogoutEventEntry builds an entry creation input for when a user logs out.
func BuildLogoutEventEntry(userID uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: LogoutEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey: userID,
		},
	}
}
