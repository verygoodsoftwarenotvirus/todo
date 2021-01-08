package bitmask

import (
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
)

const (
	canCreateItemPermissions AccountUserPermissions = 1 << iota
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
	gob.Register(AccountUserPermissions(0))
}

// AccountUserPermissions is a bitmask for keeping track of admin user permissions.
type AccountUserPermissions uint32

// NewAccountUserPermissions builds a new AccountUserPermissions.
func NewAccountUserPermissions(x uint32) AccountUserPermissions {
	return AccountUserPermissions(x)
}

// Value implements the driver.Valuer interface.
func (p AccountUserPermissions) Value() (driver.Value, error) {
	return driver.Value(int64(p)), nil
}

// Scan implements the sql.Scanner interface.
func (p *AccountUserPermissions) Scan(value interface{}) error {
	b, ok := value.(int32)
	if !ok {
		*p = AccountUserPermissions(0)
	}

	*p = AccountUserPermissions(b)

	return nil
}

var _ json.Marshaler = (*AccountUserPermissions)(nil)

// MarshalJSON implements the json.Marshaler interface.
func (p *AccountUserPermissions) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint32(*p))
}

var _ json.Unmarshaler = (*AccountUserPermissions)(nil)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (p *AccountUserPermissions) UnmarshalJSON(data []byte) error {
	var v uint32
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	*p = AccountUserPermissions(v)

	return nil
}

// CanCycleCookieSecrets determines whether or not a user can cycle cookie secrets.
func (p AccountUserPermissions) CanCreateItems() bool {
	return p&canCreateItemPermissions != 0
}

// CanBanUsers determines whether or not a user can ban users.
func (p AccountUserPermissions) CanUpdateItems() bool {
	return p&canUpdateItemPermissions != 0
}

// CanTerminateAccounts determines whether or not a user can terminate accounts.
func (p AccountUserPermissions) CanDeleteItems() bool {
	return p&canDeleteItemPermissions != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission4() bool {
	return p&unusedAccountUserPermission4 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission5() bool {
	return p&unusedAccountUserPermission5 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission6() bool {
	return p&unusedAccountUserPermission6 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission7() bool {
	return p&unusedAccountUserPermission7 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission8() bool {
	return p&unusedAccountUserPermission8 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission9() bool {
	return p&unusedAccountUserPermission9 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission10() bool {
	return p&unusedAccountUserPermission10 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission11() bool {
	return p&unusedAccountUserPermission11 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission12() bool {
	return p&unusedAccountUserPermission12 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission13() bool {
	return p&unusedAccountUserPermission13 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission14() bool {
	return p&unusedAccountUserPermission14 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission15() bool {
	return p&unusedAccountUserPermission15 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission16() bool {
	return p&unusedAccountUserPermission16 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission17() bool {
	return p&unusedAccountUserPermission17 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission18() bool {
	return p&unusedAccountUserPermission18 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission19() bool {
	return p&unusedAccountUserPermission19 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission20() bool {
	return p&unusedAccountUserPermission20 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission21() bool {
	return p&unusedAccountUserPermission21 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission22() bool {
	return p&unusedAccountUserPermission22 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission23() bool {
	return p&unusedAccountUserPermission23 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission24() bool {
	return p&unusedAccountUserPermission24 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission25() bool {
	return p&unusedAccountUserPermission25 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission26() bool {
	return p&unusedAccountUserPermission26 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission27() bool {
	return p&unusedAccountUserPermission27 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission28() bool {
	return p&unusedAccountUserPermission28 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission29() bool {
	return p&unusedAccountUserPermission29 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission30() bool {
	return p&unusedAccountUserPermission30 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission31() bool {
	return p&unusedAccountUserPermission31 != 0
}

func (p AccountUserPermissions) hasReservedUnusedPermission32() bool {
	return p&unusedAccountUserPermission32 != 0
}
