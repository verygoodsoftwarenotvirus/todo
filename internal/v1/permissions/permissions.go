package permissions

const (
	// CycleCookieSecretPermission signifies whether or not the admin in question can cycle cookie secrets.
	CycleCookieSecretPermission uint32 = 1 << iota
	reservedUnusedPermission2
	reservedUnusedPermission3
	reservedUnusedPermission4
	reservedUnusedPermission5
	reservedUnusedPermission6
	reservedUnusedPermission7
	reservedUnusedPermission8
	reservedUnusedPermission9
	reservedUnusedPermission10
	reservedUnusedPermission11
	reservedUnusedPermission12
	reservedUnusedPermission13
	reservedUnusedPermission14
	reservedUnusedPermission15
	reservedUnusedPermission16
	reservedUnusedPermission17
	reservedUnusedPermission18
	reservedUnusedPermission19
	reservedUnusedPermission20
	reservedUnusedPermission21
	reservedUnusedPermission22
	reservedUnusedPermission23
	reservedUnusedPermission24
	reservedUnusedPermission25
	reservedUnusedPermission26
	reservedUnusedPermission27
	reservedUnusedPermission28
	reservedUnusedPermission29
	reservedUnusedPermission30
	reservedUnusedPermission31
	reservedUnusedPermission32
)

// PermissionChecker returns whether or not a given permission applies to a user.
type PermissionChecker interface {
	CanCycleCookieSecrets() bool
}
