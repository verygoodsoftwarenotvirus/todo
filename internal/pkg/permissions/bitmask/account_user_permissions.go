package bitmask

import (
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
)

const (
	canCreateItemPermissions ServiceUserPermissions = 1 << iota
	canUpdateItemPermissions
	canDeleteItemPermissions
	unusedAccountUserPermission4
	unusedAccountUserPermission5
	unusedAccountUserPermission6
	unusedAccountUserPermission7
	unusedAccountUserPermission8
	unusedAccountUserPermission9
	unusedAccountUserPermission10
	unusedAccountUserPermission11
	unusedAccountUserPermission12
	unusedAccountUserPermission13
	unusedAccountUserPermission14
	unusedAccountUserPermission15
	unusedAccountUserPermission16
	unusedAccountUserPermission17
	unusedAccountUserPermission18
	unusedAccountUserPermission19
	unusedAccountUserPermission20
	unusedAccountUserPermission21
	unusedAccountUserPermission22
	unusedAccountUserPermission23
	unusedAccountUserPermission24
	unusedAccountUserPermission25
	unusedAccountUserPermission26
	unusedAccountUserPermission27
	unusedAccountUserPermission28
	unusedAccountUserPermission29
	unusedAccountUserPermission30
	unusedAccountUserPermission31
	unusedAccountUserPermission32
)

func init() {
	gob.Register(ServiceUserPermissions(0))
}

// ServiceUserPermissions is a bitmask for keeping track of admin user permissions.
type ServiceUserPermissions uint32

// NewAccountUserPermissions builds a new ServiceUserPermissions.
func NewAccountUserPermissions(x uint32) ServiceUserPermissions {
	return ServiceUserPermissions(x)
}

// Value implements the driver.Valuer interface.
func (p ServiceUserPermissions) Value() (driver.Value, error) {
	return driver.Value(int64(p)), nil
}

// Scan implements the sql.Scanner interface.
func (p *ServiceUserPermissions) Scan(value interface{}) error {
	b, ok := value.(int32)
	if !ok {
		*p = ServiceUserPermissions(0)
	}

	*p = ServiceUserPermissions(b)

	return nil
}

var _ json.Marshaler = (*ServiceUserPermissions)(nil)

// MarshalJSON implements the json.Marshaler interface.
func (p *ServiceUserPermissions) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint32(*p))
}

var _ json.Unmarshaler = (*ServiceUserPermissions)(nil)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (p *ServiceUserPermissions) UnmarshalJSON(data []byte) error {
	var v uint32
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	*p = ServiceUserPermissions(v)

	return nil
}

// CanCreateItems determines whether or not a user can create items.
func (p ServiceUserPermissions) CanCreateItems() bool {
	return p&canCreateItemPermissions != 0
}

// CanUpdateItems determines whether or not a user can update items.
func (p ServiceUserPermissions) CanUpdateItems() bool {
	return p&canUpdateItemPermissions != 0
}

// CanArchiveItems determines whether or not a user can archive items.
func (p ServiceUserPermissions) CanArchiveItems() bool {
	return p&canDeleteItemPermissions != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission4() bool {
	return p&unusedAccountUserPermission4 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission5() bool {
	return p&unusedAccountUserPermission5 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission6() bool {
	return p&unusedAccountUserPermission6 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission7() bool {
	return p&unusedAccountUserPermission7 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission8() bool {
	return p&unusedAccountUserPermission8 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission9() bool {
	return p&unusedAccountUserPermission9 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission10() bool {
	return p&unusedAccountUserPermission10 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission11() bool {
	return p&unusedAccountUserPermission11 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission12() bool {
	return p&unusedAccountUserPermission12 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission13() bool {
	return p&unusedAccountUserPermission13 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission14() bool {
	return p&unusedAccountUserPermission14 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission15() bool {
	return p&unusedAccountUserPermission15 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission16() bool {
	return p&unusedAccountUserPermission16 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission17() bool {
	return p&unusedAccountUserPermission17 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission18() bool {
	return p&unusedAccountUserPermission18 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission19() bool {
	return p&unusedAccountUserPermission19 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission20() bool {
	return p&unusedAccountUserPermission20 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission21() bool {
	return p&unusedAccountUserPermission21 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission22() bool {
	return p&unusedAccountUserPermission22 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission23() bool {
	return p&unusedAccountUserPermission23 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission24() bool {
	return p&unusedAccountUserPermission24 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission25() bool {
	return p&unusedAccountUserPermission25 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission26() bool {
	return p&unusedAccountUserPermission26 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission27() bool {
	return p&unusedAccountUserPermission27 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission28() bool {
	return p&unusedAccountUserPermission28 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission29() bool {
	return p&unusedAccountUserPermission29 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission30() bool {
	return p&unusedAccountUserPermission30 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission31() bool {
	return p&unusedAccountUserPermission31 != 0
}

func (p ServiceUserPermissions) hasReservedUnusedPermission32() bool {
	return p&unusedAccountUserPermission32 != 0
}
