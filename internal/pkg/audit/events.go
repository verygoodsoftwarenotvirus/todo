package audit

const (
	// UserBannedEvent events indicate an admin cycled the cookie secret.
	UserBannedEvent = "user_banned"
	// AccountTerminatedEvent events indicate an admin cycled the cookie secret.
	AccountTerminatedEvent = "account_terminated"
	// CycleCookieSecretEvent events indicate an admin cycled the cookie secret.
	CycleCookieSecretEvent = "cookie_secret_cycled"
	// SuccessfulLoginEvent events indicate a user successfully authenticated into the service via username + password + 2fa.
	SuccessfulLoginEvent = "user_logged_in"
	// LogoutEvent events indicate a user successfully logged out.
	LogoutEvent = "user_logged_out"
	// BannedUserLoginAttemptEvent events indicate a user successfully authenticated into the service via username + password + 2fa.
	BannedUserLoginAttemptEvent = "banned_user_login_attempt"
	// UnsuccessfulLoginBadPasswordEvent events indicate a user attempted to authenticate into the service, but failed because of an invalid password.
	UnsuccessfulLoginBadPasswordEvent = "user_login_failed_bad_password"
	// UnsuccessfulLoginBad2FATokenEvent events indicate a user attempted to authenticate into the service, but failed because of a faulty two factor token.
	UnsuccessfulLoginBad2FATokenEvent = "user_login_failed_bad_2FA_token"

	// AccountSubscriptionPlanCreationEvent events indicate a user created a plan.
	AccountSubscriptionPlanCreationEvent = "plan_created"
	// AccountSubscriptionPlanUpdateEvent events indicate a user updated a plan.
	AccountSubscriptionPlanUpdateEvent = "plan_updated"
	// AccountSubscriptionPlanArchiveEvent events indicate a user deleted a plan.
	AccountSubscriptionPlanArchiveEvent = "plan_archived"

	// AccountCreationEvent events indicate a user created an account.
	AccountCreationEvent = "account_created"
	// AccountUpdateEvent events indicate a user updated an account.
	AccountUpdateEvent = "account_updated"
	// AccountArchiveEvent events indicate a user deleted an account.
	AccountArchiveEvent = "account_archived"

	// ItemCreationEvent events indicate a user created an item.
	ItemCreationEvent = "item_created"
	// ItemUpdateEvent events indicate a user updated an item.
	ItemUpdateEvent = "item_updated"
	// ItemArchiveEvent events indicate a user deleted an item.
	ItemArchiveEvent = "item_archived"

	// OAuth2ClientCreationEvent events indicate a user created an item.
	OAuth2ClientCreationEvent = "oauth2_client_created"
	// OAuth2ClientArchiveEvent events indicate a user deleted an item.
	OAuth2ClientArchiveEvent = "oauth2_client_archived"
	// WebhookCreationEvent events indicate a user created an item.
	WebhookCreationEvent = "webhook_created"
	// WebhookUpdateEvent events indicate a user updated an item.
	WebhookUpdateEvent = "webhook_updated"
	// WebhookArchiveEvent events indicate a user deleted an item.
	WebhookArchiveEvent = "webhook_archived"
	// UserCreationEvent events indicate a user was created.
	UserCreationEvent = "user_account_created"
	// UserVerifyTwoFactorSecretEvent events indicate a user was created.
	UserVerifyTwoFactorSecretEvent = "user_two_factor_secret_verified"
	// UserUpdateTwoFactorSecretEvent events indicate a user updated their two factor secret.
	UserUpdateTwoFactorSecretEvent = "user_two_factor_secret_changed"
	// UserUpdatePasswordEvent events indicate a user updated their two factor secret.
	UserUpdatePasswordEvent = "user_password_updated"
	// UserArchiveEvent events indicate a user was archived.
	UserArchiveEvent = "user_archived"
)
