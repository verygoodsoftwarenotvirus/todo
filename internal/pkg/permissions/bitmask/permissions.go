package bitmask

import "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"

const (
	// cycleCookieSecretPermission signifies whether or not the admin in question can cycle cookie secrets.
	cycleCookieSecretPermission permission = 1 << iota
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
	completeAdministrativePrivilegesPermission
)

type permission uint32

// NewPermissionMask builds a new PermissionChecker.
func NewPermissionMask(x uint32) permissions.PermissionChecker {
	return permission(x)
}

func (p permission) CanCycleCookieSecrets() bool {
	return p&cycleCookieSecretPermission != 0
}

func (p permission) hasReservedUnusedPermission2() bool {
	return p&reservedUnusedPermission2 != 0
}

func (p permission) hasReservedUnusedPermission3() bool {
	return p&reservedUnusedPermission3 != 0
}

func (p permission) hasReservedUnusedPermission4() bool {
	return p&reservedUnusedPermission4 != 0
}

func (p permission) hasReservedUnusedPermission5() bool {
	return p&reservedUnusedPermission5 != 0
}

func (p permission) hasReservedUnusedPermission6() bool {
	return p&reservedUnusedPermission6 != 0
}

func (p permission) hasReservedUnusedPermission7() bool {
	return p&reservedUnusedPermission7 != 0
}

func (p permission) hasReservedUnusedPermission8() bool {
	return p&reservedUnusedPermission8 != 0
}

func (p permission) hasReservedUnusedPermission9() bool {
	return p&reservedUnusedPermission9 != 0
}

func (p permission) hasReservedUnusedPermission10() bool {
	return p&reservedUnusedPermission10 != 0
}

func (p permission) hasReservedUnusedPermission11() bool {
	return p&reservedUnusedPermission11 != 0
}

func (p permission) hasReservedUnusedPermission12() bool {
	return p&reservedUnusedPermission12 != 0
}

func (p permission) hasReservedUnusedPermission13() bool {
	return p&reservedUnusedPermission13 != 0
}

func (p permission) hasReservedUnusedPermission14() bool {
	return p&reservedUnusedPermission14 != 0
}

func (p permission) hasReservedUnusedPermission15() bool {
	return p&reservedUnusedPermission15 != 0
}

func (p permission) hasReservedUnusedPermission16() bool {
	return p&reservedUnusedPermission16 != 0
}

func (p permission) hasReservedUnusedPermission17() bool {
	return p&reservedUnusedPermission17 != 0
}

func (p permission) hasReservedUnusedPermission18() bool {
	return p&reservedUnusedPermission18 != 0
}

func (p permission) hasReservedUnusedPermission19() bool {
	return p&reservedUnusedPermission19 != 0
}

func (p permission) hasReservedUnusedPermission20() bool {
	return p&reservedUnusedPermission20 != 0
}

func (p permission) hasReservedUnusedPermission21() bool {
	return p&reservedUnusedPermission21 != 0
}

func (p permission) hasReservedUnusedPermission22() bool {
	return p&reservedUnusedPermission22 != 0
}

func (p permission) hasReservedUnusedPermission23() bool {
	return p&reservedUnusedPermission23 != 0
}

func (p permission) hasReservedUnusedPermission24() bool {
	return p&reservedUnusedPermission24 != 0
}

func (p permission) hasReservedUnusedPermission25() bool {
	return p&reservedUnusedPermission25 != 0
}

func (p permission) hasReservedUnusedPermission26() bool {
	return p&reservedUnusedPermission26 != 0
}

func (p permission) hasReservedUnusedPermission27() bool {
	return p&reservedUnusedPermission27 != 0
}

func (p permission) hasReservedUnusedPermission28() bool {
	return p&reservedUnusedPermission28 != 0
}

func (p permission) hasReservedUnusedPermission29() bool {
	return p&reservedUnusedPermission29 != 0
}

func (p permission) hasReservedUnusedPermission30() bool {
	return p&reservedUnusedPermission30 != 0
}

func (p permission) hasReservedUnusedPermission31() bool {
	return p&reservedUnusedPermission31 != 0
}

func (p permission) IsCompleteAdmin() bool {
	return p&completeAdministrativePrivilegesPermission != 0
}
