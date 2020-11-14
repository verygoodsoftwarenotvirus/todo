package bitmask

import (
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
)

const (
	// cycleCookieSecretPermission signifies whether or not the admin in question can cycle cookie secrets.
	cycleCookieSecretPermission PermissionBitmask = 1 << iota
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

func init() {
	gob.Register(PermissionBitmask(0))
}

type PermissionBitmask uint32

// NewPermissionBitmask builds a new PermissionChecker.
func NewPermissionBitmask(x uint32) PermissionBitmask {
	return PermissionBitmask(x)
}

// Value implements the driver.Valuer interface.
func (p PermissionBitmask) Value() (driver.Value, error) {
	return driver.Value(int64(p)), nil
}

// Scan implements the sql.Scanner interface.
func (p *PermissionBitmask) Scan(value interface{}) error {
	b, _ := value.(int32)
	*p = PermissionBitmask(b)
	return nil
}

var _ json.Marshaler = (*PermissionBitmask)(nil)

func (p *PermissionBitmask) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint32(*p))
}

var _ json.Unmarshaler = (*PermissionBitmask)(nil)

func (p *PermissionBitmask) UnmarshalJSON(data []byte) error {
	var v uint32
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	*p = PermissionBitmask(v)

	return nil
}

func (p PermissionBitmask) CanCycleCookieSecrets() bool {
	return p&cycleCookieSecretPermission != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission2() bool {
	return p&reservedUnusedPermission2 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission3() bool {
	return p&reservedUnusedPermission3 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission4() bool {
	return p&reservedUnusedPermission4 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission5() bool {
	return p&reservedUnusedPermission5 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission6() bool {
	return p&reservedUnusedPermission6 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission7() bool {
	return p&reservedUnusedPermission7 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission8() bool {
	return p&reservedUnusedPermission8 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission9() bool {
	return p&reservedUnusedPermission9 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission10() bool {
	return p&reservedUnusedPermission10 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission11() bool {
	return p&reservedUnusedPermission11 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission12() bool {
	return p&reservedUnusedPermission12 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission13() bool {
	return p&reservedUnusedPermission13 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission14() bool {
	return p&reservedUnusedPermission14 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission15() bool {
	return p&reservedUnusedPermission15 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission16() bool {
	return p&reservedUnusedPermission16 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission17() bool {
	return p&reservedUnusedPermission17 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission18() bool {
	return p&reservedUnusedPermission18 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission19() bool {
	return p&reservedUnusedPermission19 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission20() bool {
	return p&reservedUnusedPermission20 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission21() bool {
	return p&reservedUnusedPermission21 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission22() bool {
	return p&reservedUnusedPermission22 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission23() bool {
	return p&reservedUnusedPermission23 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission24() bool {
	return p&reservedUnusedPermission24 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission25() bool {
	return p&reservedUnusedPermission25 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission26() bool {
	return p&reservedUnusedPermission26 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission27() bool {
	return p&reservedUnusedPermission27 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission28() bool {
	return p&reservedUnusedPermission28 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission29() bool {
	return p&reservedUnusedPermission29 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission30() bool {
	return p&reservedUnusedPermission30 != 0
}

func (p PermissionBitmask) hasReservedUnusedPermission31() bool {
	return p&reservedUnusedPermission31 != 0
}

func (p PermissionBitmask) IsCompleteAdmin() bool {
	return p&completeAdministrativePrivilegesPermission != 0
}
