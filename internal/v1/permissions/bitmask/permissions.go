package bitmask

import "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/permissions"

type permissionMask uint32

// NewPermissionMask builds a new PermissionChecker.
func NewPermissionMask(x uint32) permissions.PermissionChecker {
	return permissionMask(x)
}

func (p permissionMask) CanCycleCookieSecrets() bool {
	return p&permissionMask(permissions.CycleCookieSecretPermission) != 0
}
