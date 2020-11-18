package permissions

// AdminPermissionChecker returns whether or not a given permission applies to a user.
type AdminPermissionChecker interface {
	CanCycleCookieSecrets() bool
	CanBanUsers() bool
}

// AdminPermissionsSummary summarizes a user's permissions.
type AdminPermissionsSummary struct {
	CanCycleCookieSecrets bool `json:"canCycleCookieSecret"`
	CanBanUsers           bool `json:"canBanUsers"`
}
