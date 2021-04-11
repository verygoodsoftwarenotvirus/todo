package permissions

import (
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
)

const (
	// cycleCookieSecretPermission signifies whether the admin in question can cycle cookie secrets.
	cycleCookieSecretPermission ServiceAdminPermission = 1 << iota
	banUserPermission
	canTerminateAccountsPermission
	canImpersonateAccountsPermission
	unusedServiceAdminPermission5
	unusedServiceAdminPermission6
	unusedServiceAdminPermission7
	unusedServiceAdminPermission8
	unusedServiceAdminPermission9
	unusedServiceAdminPermission10
	unusedServiceAdminPermission11
	unusedServiceAdminPermission12
	unusedServiceAdminPermission13
	unusedServiceAdminPermission14
	unusedServiceAdminPermission15
	unusedServiceAdminPermission16
	unusedServiceAdminPermission17
	unusedServiceAdminPermission18
	unusedServiceAdminPermission19
	unusedServiceAdminPermission20
	unusedServiceAdminPermission21
	unusedServiceAdminPermission22
	unusedServiceAdminPermission23
	unusedServiceAdminPermission24
	unusedServiceAdminPermission25
	unusedServiceAdminPermission26
	unusedServiceAdminPermission27
	unusedServiceAdminPermission28
	unusedServiceAdminPermission29
	unusedServiceAdminPermission30
	unusedServiceAdminPermission31
	unusedServiceAdminPermission32
	unusedServiceAdminPermission33
	unusedServiceAdminPermission34
	unusedServiceAdminPermission35
	unusedServiceAdminPermission36
	unusedServiceAdminPermission37
	unusedServiceAdminPermission38
	unusedServiceAdminPermission39
	unusedServiceAdminPermission40
	unusedServiceAdminPermission41
	unusedServiceAdminPermission42
	unusedServiceAdminPermission43
	unusedServiceAdminPermission44
	unusedServiceAdminPermission45
	unusedServiceAdminPermission46
	unusedServiceAdminPermission47
	unusedServiceAdminPermission48
	unusedServiceAdminPermission49
	unusedServiceAdminPermission50
	unusedServiceAdminPermission51
	unusedServiceAdminPermission52
	unusedServiceAdminPermission53
	unusedServiceAdminPermission54
	unusedServiceAdminPermission55
	unusedServiceAdminPermission56
	unusedServiceAdminPermission57
	unusedServiceAdminPermission58
	unusedServiceAdminPermission59
	unusedServiceAdminPermission61
	unusedServiceAdminPermission62
	unusedServiceAdminPermission63
	unusedServiceAdminPermission64
)

func init() {
	gob.Register(ServiceAdminPermission(0))
}

// ServiceAdminPermissionChecker returns whether a given permission applies to a user.
type ServiceAdminPermissionChecker interface {
	IsServiceAdmin() bool
	CanCycleCookieSecrets() bool
	CanBanUsers() bool
	CanTerminateAccounts() bool
}

// ServiceAdminPermissionsSummary summarizes a user's permissions.
type ServiceAdminPermissionsSummary struct {
	CanCycleCookieSecrets  bool `json:"canCycleCookieSecret"`
	CanBanUsers            bool `json:"canBanUsers"`
	CanTerminateAccounts   bool `json:"canTerminateAccounts"`
	CanImpersonateAccounts bool `json:"canImpersonateAccounts"`
}

// ServiceAdminPermission is a bitmask for keeping track of admin user permissions.
type ServiceAdminPermission int64

// NewServiceAdminPermissions builds a new ServiceAdminPermissionChecker.
func NewServiceAdminPermissions(x int64) ServiceAdminPermission {
	return ServiceAdminPermission(x)
}

// Value implements the driver.Valuer interface.
func (p ServiceAdminPermission) Value() (driver.Value, error) {
	return driver.Value(int64(p)), nil
}

// Scan implements the sql.Scanner interface.
func (p *ServiceAdminPermission) Scan(value interface{}) error {
	b, ok := value.(int32)
	if !ok {
		*p = ServiceAdminPermission(0)
	}

	*p = ServiceAdminPermission(b)

	return nil
}

var _ json.Marshaler = (*ServiceAdminPermission)(nil)

// MarshalJSON implements the json.Marshaler interface.
func (p *ServiceAdminPermission) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(*p))
}

var _ json.Unmarshaler = (*ServiceAdminPermission)(nil)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (p *ServiceAdminPermission) UnmarshalJSON(data []byte) error {
	var v int64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	*p = ServiceAdminPermission(v)

	return nil
}

// Summary produces a ServiceAdminPermissionsSummary.
func (p ServiceAdminPermission) Summary() *ServiceAdminPermissionsSummary {
	if p == 0 {
		return nil
	}

	return &ServiceAdminPermissionsSummary{
		CanCycleCookieSecrets:  p.CanCycleCookieSecrets(),
		CanBanUsers:            p.CanBanUsers(),
		CanTerminateAccounts:   p.CanTerminateAccounts(),
		CanImpersonateAccounts: p.CanImpersonateAccounts(),
	}
}

// HasPermission determines whether a user can view items.
func (p ServiceAdminPermission) HasPermission(perm ServiceAdminPermission) bool {
	return p&perm != 0
}

// IsServiceAdmin determines whether a user has service admin permissions.
func (p ServiceAdminPermission) IsServiceAdmin() bool {
	return p != 0
}

// CanCycleCookieSecrets determines whether a user can cycle cookie secrets.
func (p ServiceAdminPermission) CanCycleCookieSecrets() bool {
	return p.HasPermission(cycleCookieSecretPermission)
}

// CanBanUsers determines whether a user can ban users.
func (p ServiceAdminPermission) CanBanUsers() bool {
	return p.HasPermission(banUserPermission)
}

// CanTerminateAccounts determines whether a user can terminate accounts.
func (p ServiceAdminPermission) CanTerminateAccounts() bool {
	return p.HasPermission(canTerminateAccountsPermission)
}

// CanImpersonateAccounts determines whether a user can impersonate accounts.
func (p ServiceAdminPermission) CanImpersonateAccounts() bool {
	return p.HasPermission(canImpersonateAccountsPermission)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission5() bool {
	return p.HasPermission(unusedServiceAdminPermission5)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission6() bool {
	return p.HasPermission(unusedServiceAdminPermission6)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission7() bool {
	return p.HasPermission(unusedServiceAdminPermission7)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission8() bool {
	return p.HasPermission(unusedServiceAdminPermission8)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission9() bool {
	return p.HasPermission(unusedServiceAdminPermission9)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission10() bool {
	return p.HasPermission(unusedServiceAdminPermission10)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission11() bool {
	return p.HasPermission(unusedServiceAdminPermission11)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission12() bool {
	return p.HasPermission(unusedServiceAdminPermission12)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission13() bool {
	return p.HasPermission(unusedServiceAdminPermission13)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission14() bool {
	return p.HasPermission(unusedServiceAdminPermission14)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission15() bool {
	return p.HasPermission(unusedServiceAdminPermission15)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission16() bool {
	return p.HasPermission(unusedServiceAdminPermission16)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission17() bool {
	return p.HasPermission(unusedServiceAdminPermission17)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission18() bool {
	return p.HasPermission(unusedServiceAdminPermission18)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission19() bool {
	return p.HasPermission(unusedServiceAdminPermission19)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission20() bool {
	return p.HasPermission(unusedServiceAdminPermission20)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission21() bool {
	return p.HasPermission(unusedServiceAdminPermission21)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission22() bool {
	return p.HasPermission(unusedServiceAdminPermission22)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission23() bool {
	return p.HasPermission(unusedServiceAdminPermission23)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission24() bool {
	return p.HasPermission(unusedServiceAdminPermission24)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission25() bool {
	return p.HasPermission(unusedServiceAdminPermission25)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission26() bool {
	return p.HasPermission(unusedServiceAdminPermission26)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission27() bool {
	return p.HasPermission(unusedServiceAdminPermission27)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission28() bool {
	return p.HasPermission(unusedServiceAdminPermission28)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission29() bool {
	return p.HasPermission(unusedServiceAdminPermission29)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission30() bool {
	return p.HasPermission(unusedServiceAdminPermission30)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission31() bool {
	return p.HasPermission(unusedServiceAdminPermission31)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission32() bool {
	return p.HasPermission(unusedServiceAdminPermission32)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission33() bool {
	return p.HasPermission(unusedServiceAdminPermission33)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission34() bool {
	return p.HasPermission(unusedServiceAdminPermission34)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission35() bool {
	return p.HasPermission(unusedServiceAdminPermission35)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission36() bool {
	return p.HasPermission(unusedServiceAdminPermission36)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission37() bool {
	return p.HasPermission(unusedServiceAdminPermission37)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission38() bool {
	return p.HasPermission(unusedServiceAdminPermission38)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission39() bool {
	return p.HasPermission(unusedServiceAdminPermission39)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission40() bool {
	return p.HasPermission(unusedServiceAdminPermission40)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission41() bool {
	return p.HasPermission(unusedServiceAdminPermission41)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission42() bool {
	return p.HasPermission(unusedServiceAdminPermission42)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission43() bool {
	return p.HasPermission(unusedServiceAdminPermission43)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission44() bool {
	return p.HasPermission(unusedServiceAdminPermission44)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission45() bool {
	return p.HasPermission(unusedServiceAdminPermission45)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission46() bool {
	return p.HasPermission(unusedServiceAdminPermission46)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission47() bool {
	return p.HasPermission(unusedServiceAdminPermission47)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission48() bool {
	return p.HasPermission(unusedServiceAdminPermission48)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission49() bool {
	return p.HasPermission(unusedServiceAdminPermission49)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission50() bool {
	return p.HasPermission(unusedServiceAdminPermission50)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission51() bool {
	return p.HasPermission(unusedServiceAdminPermission51)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission52() bool {
	return p.HasPermission(unusedServiceAdminPermission52)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission53() bool {
	return p.HasPermission(unusedServiceAdminPermission53)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission54() bool {
	return p.HasPermission(unusedServiceAdminPermission54)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission55() bool {
	return p.HasPermission(unusedServiceAdminPermission55)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission56() bool {
	return p.HasPermission(unusedServiceAdminPermission56)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission57() bool {
	return p.HasPermission(unusedServiceAdminPermission57)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission58() bool {
	return p.HasPermission(unusedServiceAdminPermission58)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission59() bool {
	return p.HasPermission(unusedServiceAdminPermission59)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission61() bool {
	return p.HasPermission(unusedServiceAdminPermission61)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission62() bool {
	return p.HasPermission(unusedServiceAdminPermission62)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission63() bool {
	return p.HasPermission(unusedServiceAdminPermission63)
}

func (p ServiceAdminPermission) hasUnusedServiceAdminPermission64() bool {
	return p.HasPermission(unusedServiceAdminPermission64)
}
