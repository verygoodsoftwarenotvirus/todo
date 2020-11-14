package bitmask

import (
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
)

const (
	// cycleCookieSecretPermission signifies whether or not the admin in question can cycle cookie secrets.
	cycleCookieSecretPermission AdminPermissionsBitmask = 1 << iota
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

func init() {
	gob.Register(AdminPermissionsBitmask(0))
}

// AdminPermissionsBitmask is a bitmask for keeping track of admin user permissions.
type AdminPermissionsBitmask uint32

// NewPermissionBitmask builds a new PermissionChecker.
func NewPermissionBitmask(x uint32) AdminPermissionsBitmask {
	return AdminPermissionsBitmask(x)
}

// Value implements the driver.Valuer interface.
func (p AdminPermissionsBitmask) Value() (driver.Value, error) {
	return driver.Value(int64(p)), nil
}

// Scan implements the sql.Scanner interface.
func (p *AdminPermissionsBitmask) Scan(value interface{}) error {
	b, ok := value.(int32)
	if !ok {
		*p = AdminPermissionsBitmask(0)
	}

	*p = AdminPermissionsBitmask(b)

	return nil
}

var _ json.Marshaler = (*AdminPermissionsBitmask)(nil)

// MarshalJSON implements the json.Marshaler interface.
func (p *AdminPermissionsBitmask) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint32(*p))
}

var _ json.Unmarshaler = (*AdminPermissionsBitmask)(nil)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (p *AdminPermissionsBitmask) UnmarshalJSON(data []byte) error {
	var v uint32
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	*p = AdminPermissionsBitmask(v)

	return nil
}

// CanCycleCookieSecrets determines whether or not a user can cycle cookie secrets.
func (p AdminPermissionsBitmask) CanCycleCookieSecrets() bool {
	return p&cycleCookieSecretPermission != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission2() bool {
	return p&reservedUnusedPermission2 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission3() bool {
	return p&reservedUnusedPermission3 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission4() bool {
	return p&reservedUnusedPermission4 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission5() bool {
	return p&reservedUnusedPermission5 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission6() bool {
	return p&reservedUnusedPermission6 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission7() bool {
	return p&reservedUnusedPermission7 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission8() bool {
	return p&reservedUnusedPermission8 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission9() bool {
	return p&reservedUnusedPermission9 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission10() bool {
	return p&reservedUnusedPermission10 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission11() bool {
	return p&reservedUnusedPermission11 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission12() bool {
	return p&reservedUnusedPermission12 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission13() bool {
	return p&reservedUnusedPermission13 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission14() bool {
	return p&reservedUnusedPermission14 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission15() bool {
	return p&reservedUnusedPermission15 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission16() bool {
	return p&reservedUnusedPermission16 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission17() bool {
	return p&reservedUnusedPermission17 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission18() bool {
	return p&reservedUnusedPermission18 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission19() bool {
	return p&reservedUnusedPermission19 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission20() bool {
	return p&reservedUnusedPermission20 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission21() bool {
	return p&reservedUnusedPermission21 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission22() bool {
	return p&reservedUnusedPermission22 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission23() bool {
	return p&reservedUnusedPermission23 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission24() bool {
	return p&reservedUnusedPermission24 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission25() bool {
	return p&reservedUnusedPermission25 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission26() bool {
	return p&reservedUnusedPermission26 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission27() bool {
	return p&reservedUnusedPermission27 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission28() bool {
	return p&reservedUnusedPermission28 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission29() bool {
	return p&reservedUnusedPermission29 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission30() bool {
	return p&reservedUnusedPermission30 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission31() bool {
	return p&reservedUnusedPermission31 != 0
}

func (p AdminPermissionsBitmask) hasReservedUnusedPermission32() bool {
	return p&reservedUnusedPermission32 != 0
}
