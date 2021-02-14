package bitmask

import (
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
)

const (
	// cycleCookieSecretPermission signifies whether or not the admin in question can cycle cookie secrets.
	cycleCookieSecretPermission ServiceAdminPermissions = 1 << iota
	banUserPermission
	canTerminateAccountsPermission
	unusedServiceAdminPermission4
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
)

func init() {
	gob.Register(ServiceAdminPermissions(0))
}

// ServiceAdminPermissions is a bitmask for keeping track of admin user permissions.
type ServiceAdminPermissions uint32

// NewServiceAdminPermissions builds a new ServiceAdminPermissionChecker.
func NewServiceAdminPermissions(x uint32) ServiceAdminPermissions {
	return ServiceAdminPermissions(x)
}

// ServiceAdminPermissionsSummary produces a ServiceAdminPermissionsSummary.
func (p ServiceAdminPermissions) ServiceAdminPermissionsSummary() *permissions.ServiceAdminPermissionsSummary {
	if p == 0 {
		return nil
	}

	return &permissions.ServiceAdminPermissionsSummary{
		CanCycleCookieSecrets: p.CanCycleCookieSecrets(),
		CanBanUsers:           p.CanBanUsers(),
	}
}

// Value implements the driver.Valuer interface.
func (p ServiceAdminPermissions) Value() (driver.Value, error) {
	return driver.Value(int64(p)), nil
}

// Scan implements the sql.Scanner interface.
func (p *ServiceAdminPermissions) Scan(value interface{}) error {
	b, ok := value.(int32)
	if !ok {
		*p = ServiceAdminPermissions(0)
	}

	*p = ServiceAdminPermissions(b)

	return nil
}

var _ json.Marshaler = (*ServiceAdminPermissions)(nil)

// MarshalJSON implements the json.Marshaler interface.
func (p *ServiceAdminPermissions) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint32(*p))
}

var _ json.Unmarshaler = (*ServiceAdminPermissions)(nil)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (p *ServiceAdminPermissions) UnmarshalJSON(data []byte) error {
	var v uint32
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	*p = ServiceAdminPermissions(v)

	return nil
}

// IsServiceAdmin determines whether or not a user has service admin permissions.
func (p ServiceAdminPermissions) IsServiceAdmin() bool {
	return p != 0
}

// CanCycleCookieSecrets determines whether or not a user can cycle cookie secrets.
func (p ServiceAdminPermissions) CanCycleCookieSecrets() bool {
	return p&cycleCookieSecretPermission != 0
}

// CanBanUsers determines whether or not a user can ban users.
func (p ServiceAdminPermissions) CanBanUsers() bool {
	return p&banUserPermission != 0
}

// CanTerminateAccounts determines whether or not a user can terminate accounts.
func (p ServiceAdminPermissions) CanTerminateAccounts() bool {
	return p&canTerminateAccountsPermission != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission4() bool {
	return p&unusedServiceAdminPermission4 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission5() bool {
	return p&unusedServiceAdminPermission5 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission6() bool {
	return p&unusedServiceAdminPermission6 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission7() bool {
	return p&unusedServiceAdminPermission7 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission8() bool {
	return p&unusedServiceAdminPermission8 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission9() bool {
	return p&unusedServiceAdminPermission9 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission10() bool {
	return p&unusedServiceAdminPermission10 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission11() bool {
	return p&unusedServiceAdminPermission11 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission12() bool {
	return p&unusedServiceAdminPermission12 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission13() bool {
	return p&unusedServiceAdminPermission13 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission14() bool {
	return p&unusedServiceAdminPermission14 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission15() bool {
	return p&unusedServiceAdminPermission15 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission16() bool {
	return p&unusedServiceAdminPermission16 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission17() bool {
	return p&unusedServiceAdminPermission17 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission18() bool {
	return p&unusedServiceAdminPermission18 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission19() bool {
	return p&unusedServiceAdminPermission19 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission20() bool {
	return p&unusedServiceAdminPermission20 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission21() bool {
	return p&unusedServiceAdminPermission21 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission22() bool {
	return p&unusedServiceAdminPermission22 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission23() bool {
	return p&unusedServiceAdminPermission23 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission24() bool {
	return p&unusedServiceAdminPermission24 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission25() bool {
	return p&unusedServiceAdminPermission25 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission26() bool {
	return p&unusedServiceAdminPermission26 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission27() bool {
	return p&unusedServiceAdminPermission27 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission28() bool {
	return p&unusedServiceAdminPermission28 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission29() bool {
	return p&unusedServiceAdminPermission29 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission30() bool {
	return p&unusedServiceAdminPermission30 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission31() bool {
	return p&unusedServiceAdminPermission31 != 0
}

func (p ServiceAdminPermissions) hasReservedUnusedPermission32() bool {
	return p&unusedServiceAdminPermission32 != 0
}
