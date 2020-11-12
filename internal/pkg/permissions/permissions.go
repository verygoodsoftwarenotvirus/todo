package permissions

// PermissionChecker returns whether or not a given permission applies to a user.
type PermissionChecker interface {
	CanCycleCookieSecrets() bool
}
