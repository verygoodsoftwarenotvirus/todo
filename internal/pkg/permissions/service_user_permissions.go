package permissions

import (
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
)

const (
	// CanManageWebhooks represents a service user's ability to create webhooks.
	CanManageWebhooks ServiceUserPermission = 1 << iota
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
	unusedAccountUserPermission33
	unusedAccountUserPermission34
	unusedAccountUserPermission35
	unusedAccountUserPermission36
	unusedAccountUserPermission37
	unusedAccountUserPermission38
	unusedAccountUserPermission39
	unusedAccountUserPermission40
	unusedAccountUserPermission41
	unusedAccountUserPermission42
	unusedAccountUserPermission43
	unusedAccountUserPermission44
	unusedAccountUserPermission45
	unusedAccountUserPermission46
	unusedAccountUserPermission47
	unusedAccountUserPermission48
	unusedAccountUserPermission49
	unusedAccountUserPermission50
	unusedAccountUserPermission51
	unusedAccountUserPermission52
	unusedAccountUserPermission53
	unusedAccountUserPermission54
	unusedAccountUserPermission55
	unusedAccountUserPermission56
	unusedAccountUserPermission57
	unusedAccountUserPermission58
	unusedAccountUserPermission59
	unusedAccountUserPermission61
	unusedAccountUserPermission62
	unusedAccountUserPermission63
	unusedAccountUserPermission64
)

func init() {
	gob.Register(ServiceUserPermission(0))
}

// ServiceUserPermissionChecker returns whether a given permission applies to a user.
type ServiceUserPermissionChecker interface {
	// CanManageAPIClients should return whether a user is authorized to manage API clients.
	CanManageAPIClients() bool

	// CanManageWebhooks should return whether a user is authorized to manage webhooks.
	CanManageWebhooks() bool
}

// ServiceUserPermissionsSummary summarizes a user's permissions.
type ServiceUserPermissionsSummary struct {
	CanManageWebhooks   bool `json:"canManageWebhooks"`
	CanManageAPIClients bool `json:"canManageAPIClients"`
}

// ServiceUserPermission is a bitmask for keeping track of admin user permissions.
type ServiceUserPermission int64

// NewServiceUserPermissions builds a new ServiceUserPermission.
func NewServiceUserPermissions(x int64) ServiceUserPermission {
	return ServiceUserPermission(x)
}

// NewServiceUserPermissionChecker builds a new ServiceUserPermissionChecker.
func NewServiceUserPermissionChecker(x int64) ServiceUserPermissionChecker {
	return NewServiceUserPermissions(x)
}

// Summary implements the driver.Valuer interface.
func (p ServiceUserPermission) Summary() ServiceUserPermissionsSummary {
	return ServiceUserPermissionsSummary{
		CanManageWebhooks:   p.CanManageWebhooks(),
		CanManageAPIClients: p.CanManageAPIClients(),
	}
}

// Value implements the driver.Valuer interface.
func (p ServiceUserPermission) Value() (driver.Value, error) {
	return driver.Value(int64(p)), nil
}

// Scan implements the sql.Scanner interface.
func (p *ServiceUserPermission) Scan(value interface{}) error {
	b, ok := value.(int32)
	if !ok {
		*p = ServiceUserPermission(0)
	}

	*p = ServiceUserPermission(b)

	return nil
}

var _ json.Marshaler = (*ServiceUserPermission)(nil)

// MarshalJSON implements the json.Marshaler interface.
func (p *ServiceUserPermission) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(*p))
}

var _ json.Unmarshaler = (*ServiceUserPermission)(nil)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (p *ServiceUserPermission) UnmarshalJSON(data []byte) error {
	var v int64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	*p = ServiceUserPermission(v)

	return nil
}

// HasPermission determines whether a user can view items.
func (p ServiceUserPermission) HasPermission(perm ServiceUserPermission) bool {
	return p&perm != 0
}

// CanManageWebhooks determines whether a user can create items.
func (p ServiceUserPermission) CanManageWebhooks() bool {
	return p.HasPermission(CanManageWebhooks)
}

// CanManageAPIClients determines whether a user can create items.
func (p ServiceUserPermission) CanManageAPIClients() bool {
	return p.HasPermission(CanManageAPIClients)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission3() bool {
	return p.HasPermission(unusedAccountUserPermission3)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission4() bool {
	return p.HasPermission(unusedAccountUserPermission4)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission5() bool {
	return p.HasPermission(unusedAccountUserPermission5)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission6() bool {
	return p.HasPermission(unusedAccountUserPermission6)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission7() bool {
	return p.HasPermission(unusedAccountUserPermission7)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission8() bool {
	return p.HasPermission(unusedAccountUserPermission8)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission9() bool {
	return p.HasPermission(unusedAccountUserPermission9)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission10() bool {
	return p.HasPermission(unusedAccountUserPermission10)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission11() bool {
	return p.HasPermission(unusedAccountUserPermission11)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission12() bool {
	return p.HasPermission(unusedAccountUserPermission12)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission13() bool {
	return p.HasPermission(unusedAccountUserPermission13)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission14() bool {
	return p.HasPermission(unusedAccountUserPermission14)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission15() bool {
	return p.HasPermission(unusedAccountUserPermission15)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission16() bool {
	return p.HasPermission(unusedAccountUserPermission16)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission17() bool {
	return p.HasPermission(unusedAccountUserPermission17)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission18() bool {
	return p.HasPermission(unusedAccountUserPermission18)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission19() bool {
	return p.HasPermission(unusedAccountUserPermission19)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission20() bool {
	return p.HasPermission(unusedAccountUserPermission20)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission21() bool {
	return p.HasPermission(unusedAccountUserPermission21)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission22() bool {
	return p.HasPermission(unusedAccountUserPermission22)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission23() bool {
	return p.HasPermission(unusedAccountUserPermission23)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission24() bool {
	return p.HasPermission(unusedAccountUserPermission24)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission25() bool {
	return p.HasPermission(unusedAccountUserPermission25)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission26() bool {
	return p.HasPermission(unusedAccountUserPermission26)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission27() bool {
	return p.HasPermission(unusedAccountUserPermission27)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission28() bool {
	return p.HasPermission(unusedAccountUserPermission28)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission29() bool {
	return p.HasPermission(unusedAccountUserPermission29)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission30() bool {
	return p.HasPermission(unusedAccountUserPermission30)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission31() bool {
	return p.HasPermission(unusedAccountUserPermission31)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission32() bool {
	return p.HasPermission(unusedAccountUserPermission32)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission33() bool {
	return p.HasPermission(unusedAccountUserPermission33)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission34() bool {
	return p.HasPermission(unusedAccountUserPermission34)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission35() bool {
	return p.HasPermission(unusedAccountUserPermission35)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission36() bool {
	return p.HasPermission(unusedAccountUserPermission36)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission37() bool {
	return p.HasPermission(unusedAccountUserPermission37)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission38() bool {
	return p.HasPermission(unusedAccountUserPermission38)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission39() bool {
	return p.HasPermission(unusedAccountUserPermission39)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission40() bool {
	return p.HasPermission(unusedAccountUserPermission40)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission41() bool {
	return p.HasPermission(unusedAccountUserPermission41)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission42() bool {
	return p.HasPermission(unusedAccountUserPermission42)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission43() bool {
	return p.HasPermission(unusedAccountUserPermission43)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission44() bool {
	return p.HasPermission(unusedAccountUserPermission44)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission45() bool {
	return p.HasPermission(unusedAccountUserPermission45)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission46() bool {
	return p.HasPermission(unusedAccountUserPermission46)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission47() bool {
	return p.HasPermission(unusedAccountUserPermission47)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission48() bool {
	return p.HasPermission(unusedAccountUserPermission48)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission49() bool {
	return p.HasPermission(unusedAccountUserPermission49)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission50() bool {
	return p.HasPermission(unusedAccountUserPermission50)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission51() bool {
	return p.HasPermission(unusedAccountUserPermission51)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission52() bool {
	return p.HasPermission(unusedAccountUserPermission52)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission53() bool {
	return p.HasPermission(unusedAccountUserPermission53)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission54() bool {
	return p.HasPermission(unusedAccountUserPermission54)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission55() bool {
	return p.HasPermission(unusedAccountUserPermission55)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission56() bool {
	return p.HasPermission(unusedAccountUserPermission56)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission57() bool {
	return p.HasPermission(unusedAccountUserPermission57)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission58() bool {
	return p.HasPermission(unusedAccountUserPermission58)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission59() bool {
	return p.HasPermission(unusedAccountUserPermission59)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission61() bool {
	return p.HasPermission(unusedAccountUserPermission61)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission62() bool {
	return p.HasPermission(unusedAccountUserPermission62)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission63() bool {
	return p.HasPermission(unusedAccountUserPermission63)
}

func (p ServiceUserPermission) hasUnusedAccountUserPermission64() bool {
	return p.HasPermission(unusedAccountUserPermission64)
}
