package audit

const (
	// CycleCookieSecretEvent events indicate an admin cycled the cookie secret.
	CycleCookieSecretEvent = iota
	// SuccessfulLoginEvent events indicate a user successfully authenticated into the service via username + password + 2fa.
	SuccessfulLoginEvent
	// UnsuccessfulLoginBadPasswordEvent events indicate a user attempted to authenticate into the service, but failed because of an invalid password.
	UnsuccessfulLoginBadPasswordEvent
	// UnsuccessfulLoginBad2FATokenEvent events indicate a user attempted to authenticate into the service, but failed because of a faulty two factor token.
	UnsuccessfulLoginBad2FATokenEvent
	// LogoutEvent events indicate a user successfully logged out.
	LogoutEvent
	// ItemCreationEvent events indicate a user created an item.
	ItemCreationEvent
	// ItemUpdateEvent events indicate a user updated an item.
	ItemUpdateEvent
	// ItemArchiveEvent events indicate a user deleted an item.
	ItemArchiveEvent
	// OAuth2ClientCreationEvent events indicate a user created an item.
	OAuth2ClientCreationEvent
	// OAuth2ClientArchiveEvent events indicate a user deleted an item.
	OAuth2ClientArchiveEvent
	// WebhookCreationEvent events indicate a user created an item.
	WebhookCreationEvent
	// WebhookUpdateEvent events indicate a user updated an item.
	WebhookUpdateEvent
	// WebhookArchiveEvent events indicate a user deleted an item.
	WebhookArchiveEvent
	// UserCreationEvent events indicate a user was created.
	UserCreationEvent
	// UserVerifyTwoFactorSecretEvent events indicate a user was created.
	UserVerifyTwoFactorSecretEvent
	// UserUpdateTwoFactorSecretEvent events indicate a user updated their two factor secret.
	UserUpdateTwoFactorSecretEvent
	// UserUpdatePasswordEvent events indicate a user updated their two factor secret.
	UserUpdatePasswordEvent
	// UserArchiveEvent events indicate a user was archived.
	UserArchiveEvent
)
