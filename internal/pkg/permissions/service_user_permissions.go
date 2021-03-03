package permissions

import (
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
)

const (
	// CanReadWebhooks represents a service user's ability to read webhooks.
	CanReadWebhooks ServiceUserPermissions = 1 << iota
	// CanCreateWebhooks represents a service user's ability to create webhooks.
	CanCreateWebhooks
	// CanUpdateWebhooks represents a service user's ability to update webhooks.
	CanUpdateWebhooks
	// CanArchiveWebhooks represents a service user's ability to delete webhooks.
	CanArchiveWebhooks
	// CanReadAPIClients represents a service user's ability to read API clients.
	CanReadAPIClients
	// CanCreateAPIClients represents a service user's ability to create API clients.
	CanCreateAPIClients
	// CanUpdateAPIClients represents a service user's ability to update API clients.
	CanUpdateAPIClients
	// CanArchiveAPIClients represents a service user's ability to delete API clients.
	CanArchiveAPIClients
	// CanReadItems represents a service user's ability to read items.
	CanReadItems
	// CanCreateItems represents a service user's ability to create items.
	CanCreateItems
	// CanUpdateItems represents a service user's ability to update items.
	CanUpdateItems
	// CanArchiveItems represents a service user's ability to delete items.
	CanArchiveItems
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

// CanReadWebhooks determines whether or not a user can view items.
func (p ServiceUserPermissions) CanReadWebhooks() bool {
	return p.HasPermission(CanReadWebhooks)
}

// CanCreateWebhooks determines whether or not a user can create items.
func (p ServiceUserPermissions) CanCreateWebhooks() bool {
	return p.HasPermission(CanCreateWebhooks)
}

// CanUpdateWebhooks determines whether or not a user can update items.
func (p ServiceUserPermissions) CanUpdateWebhooks() bool {
	return p.HasPermission(CanUpdateWebhooks)
}

// CanArchiveWebhooks determines whether or not a user can archive items.
func (p ServiceUserPermissions) CanArchiveWebhooks() bool {
	return p.HasPermission(CanArchiveWebhooks)
}

// CanReadAPIClients determines whether or not a user can view items.
func (p ServiceUserPermissions) CanReadAPIClients() bool {
	return p.HasPermission(CanReadAPIClients)
}

// CanCreateAPIClients determines whether or not a user can create items.
func (p ServiceUserPermissions) CanCreateAPIClients() bool {
	return p.HasPermission(CanCreateAPIClients)
}

// CanUpdateAPIClients determines whether or not a user can update items.
func (p ServiceUserPermissions) CanUpdateAPIClients() bool {
	return p.HasPermission(CanUpdateAPIClients)
}

// CanArchiveAPIClients determines whether or not a user can archive items.
func (p ServiceUserPermissions) CanArchiveAPIClients() bool {
	return p.HasPermission(CanArchiveAPIClients)
}

// CanReadItems determines whether or not a user can view items.
func (p ServiceUserPermissions) CanReadItems() bool {
	return p.HasPermission(CanReadItems)
}

// CanCreateItems determines whether or not a user can create items.
func (p ServiceUserPermissions) CanCreateItems() bool {
	return p.HasPermission(CanCreateItems)
}

// CanUpdateItems determines whether or not a user can update items.
func (p ServiceUserPermissions) CanUpdateItems() bool {
	return p.HasPermission(CanUpdateItems)
}

// CanArchiveItems determines whether or not a user can archive items.
func (p ServiceUserPermissions) CanArchiveItems() bool {
	return p.HasPermission(CanArchiveItems)
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
