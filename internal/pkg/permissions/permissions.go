package permissions

// SiteAdminPermissionChecker returns whether or not a given permission applies to a user.
type SiteAdminPermissionChecker interface {
	CanCycleCookieSecrets() bool
	CanBanUsers() bool
	CanTerminateAccounts() bool
}

// SiteAdminPermissionsSummary summarizes a user's permissions.
type SiteAdminPermissionsSummary struct {
	CanCycleCookieSecrets bool `json:"canCycleCookieSecret"`
	CanBanUsers           bool `json:"canBanUsers"`
	CanTerminateAccounts  bool `json:"canTerminateAccounts"`
}

// UserAccountPermissionChecker returns whether or not a given permission applies to a user.
type UserAccountPermissionChecker interface {
	CanCreateItems() bool
	CanUpdateItems() bool
	CanDeleteItems() bool
}
