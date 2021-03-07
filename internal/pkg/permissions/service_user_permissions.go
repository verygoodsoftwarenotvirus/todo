package permissions

import (
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
)

const (
	// CanManageWebhooks represents a service user's ability to create webhooks.
	CanManageWebhooks ServiceUserPermissions = 1 << iota
	// CanManageAPIClients represents a service user's ability to create API clients.
	CanManageAPIClients
	unusedAccountUserPermission3
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

// NewServiceUserPermissions builds a new ServiceUserPermissions.
func NewServiceUserPermissions(x uint32) ServiceUserPermissions {
	return ServiceUserPermissions(x)
}

// NewServiceUserPermissionChecker builds a new ServiceUserPermissionChecker.
func NewServiceUserPermissionChecker(x uint32) ServiceUserPermissionChecker {
	return NewServiceUserPermissions(x)
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

// HasPermission determines whether or not a user can view items.
func (p ServiceUserPermissions) HasPermission(perm ServiceUserPermissions) bool {
	return p&perm != 0
}

// CanManageWebhooks determines whether or not a user can create items.
func (p ServiceUserPermissions) CanManageWebhooks() bool {
	return p.HasPermission(CanManageWebhooks)
}

// CanManageAPIClients determines whether or not a user can create items.
func (p ServiceUserPermissions) CanManageAPIClients() bool {
	return p.HasPermission(CanManageAPIClients)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission3() bool {
	return p.HasPermission(unusedAccountUserPermission3)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission4() bool {
	return p.HasPermission(unusedAccountUserPermission4)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission5() bool {
	return p.HasPermission(unusedAccountUserPermission5)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission6() bool {
	return p.HasPermission(unusedAccountUserPermission6)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission7() bool {
	return p.HasPermission(unusedAccountUserPermission7)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission8() bool {
	return p.HasPermission(unusedAccountUserPermission8)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission9() bool {
	return p.HasPermission(unusedAccountUserPermission9)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission10() bool {
	return p.HasPermission(unusedAccountUserPermission10)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission11() bool {
	return p.HasPermission(unusedAccountUserPermission11)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission12() bool {
	return p.HasPermission(unusedAccountUserPermission12)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission13() bool {
	return p.HasPermission(unusedAccountUserPermission13)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission14() bool {
	return p.HasPermission(unusedAccountUserPermission14)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission15() bool {
	return p.HasPermission(unusedAccountUserPermission15)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission16() bool {
	return p.HasPermission(unusedAccountUserPermission16)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission17() bool {
	return p.HasPermission(unusedAccountUserPermission17)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission18() bool {
	return p.HasPermission(unusedAccountUserPermission18)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission19() bool {
	return p.HasPermission(unusedAccountUserPermission19)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission20() bool {
	return p.HasPermission(unusedAccountUserPermission20)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission21() bool {
	return p.HasPermission(unusedAccountUserPermission21)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission22() bool {
	return p.HasPermission(unusedAccountUserPermission22)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission23() bool {
	return p.HasPermission(unusedAccountUserPermission23)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission24() bool {
	return p.HasPermission(unusedAccountUserPermission24)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission25() bool {
	return p.HasPermission(unusedAccountUserPermission25)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission26() bool {
	return p.HasPermission(unusedAccountUserPermission26)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission27() bool {
	return p.HasPermission(unusedAccountUserPermission27)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission28() bool {
	return p.HasPermission(unusedAccountUserPermission28)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission29() bool {
	return p.HasPermission(unusedAccountUserPermission29)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission30() bool {
	return p.HasPermission(unusedAccountUserPermission30)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission31() bool {
	return p.HasPermission(unusedAccountUserPermission31)
}

func (p ServiceUserPermissions) hasReservedUnusedPermission32() bool {
	return p.HasPermission(unusedAccountUserPermission32)
}
