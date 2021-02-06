package bitmask

import (
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
)

const (
	// cycleCookieSecretPermission signifies whether or not the admin in question can cycle cookie secrets.
	cycleCookieSecretPermission SiteAdminPermissions = 1 << iota
	banUserPermission
	canTerminateAccountsPermission
	unusedSiteAdminPermission4
	unusedSiteAdminPermission5
	unusedSiteAdminPermission6
	unusedSiteAdminPermission7
	unusedSiteAdminPermission8
	unusedSiteAdminPermission9
	unusedSiteAdminPermission10
	unusedSiteAdminPermission11
	unusedSiteAdminPermission12
	unusedSiteAdminPermission13
	unusedSiteAdminPermission14
	unusedSiteAdminPermission15
	unusedSiteAdminPermission16
	unusedSiteAdminPermission17
	unusedSiteAdminPermission18
	unusedSiteAdminPermission19
	unusedSiteAdminPermission20
	unusedSiteAdminPermission21
	unusedSiteAdminPermission22
	unusedSiteAdminPermission23
	unusedSiteAdminPermission24
	unusedSiteAdminPermission25
	unusedSiteAdminPermission26
	unusedSiteAdminPermission27
	unusedSiteAdminPermission28
	unusedSiteAdminPermission29
	unusedSiteAdminPermission30
	unusedSiteAdminPermission31
	unusedSiteAdminPermission32
)

func init() {
	gob.Register(SiteAdminPermissions(0))
}

// SiteAdminPermissions is a bitmask for keeping track of admin user permissions.
type SiteAdminPermissions uint32

// NewSiteAdminPermissions builds a new SiteAdminPermissionChecker.
func NewSiteAdminPermissions(x uint32) SiteAdminPermissions {
	return SiteAdminPermissions(x)
}

// SiteAdminPermissionsSummary produces a SiteAdminPermissionsSummary.
func (p SiteAdminPermissions) SiteAdminPermissionsSummary() *permissions.SiteAdminPermissionsSummary {
	if p == 0 {
		return nil
	}

	return &permissions.SiteAdminPermissionsSummary{
		CanCycleCookieSecrets: p.CanCycleCookieSecrets(),
		CanBanUsers:           p.CanBanUsers(),
	}
}

// Value implements the driver.Valuer interface.
func (p SiteAdminPermissions) Value() (driver.Value, error) {
	return driver.Value(int64(p)), nil
}

// Scan implements the sql.Scanner interface.
func (p *SiteAdminPermissions) Scan(value interface{}) error {
	b, ok := value.(int32)
	if !ok {
		*p = SiteAdminPermissions(0)
	}

	*p = SiteAdminPermissions(b)

	return nil
}

var _ json.Marshaler = (*SiteAdminPermissions)(nil)

// MarshalJSON implements the json.Marshaler interface.
func (p *SiteAdminPermissions) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint32(*p))
}

var _ json.Unmarshaler = (*SiteAdminPermissions)(nil)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (p *SiteAdminPermissions) UnmarshalJSON(data []byte) error {
	var v uint32
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	*p = SiteAdminPermissions(v)

	return nil
}

// CanCycleCookieSecrets determines whether or not a user can cycle cookie secrets.
func (p SiteAdminPermissions) CanCycleCookieSecrets() bool {
	return p&cycleCookieSecretPermission != 0
}

// CanBanUsers determines whether or not a user can ban users.
func (p SiteAdminPermissions) CanBanUsers() bool {
	return p&banUserPermission != 0
}

// CanTerminateAccounts determines whether or not a user can terminate accounts.
func (p SiteAdminPermissions) CanTerminateAccounts() bool {
	return p&canTerminateAccountsPermission != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission4() bool {
	return p&unusedSiteAdminPermission4 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission5() bool {
	return p&unusedSiteAdminPermission5 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission6() bool {
	return p&unusedSiteAdminPermission6 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission7() bool {
	return p&unusedSiteAdminPermission7 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission8() bool {
	return p&unusedSiteAdminPermission8 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission9() bool {
	return p&unusedSiteAdminPermission9 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission10() bool {
	return p&unusedSiteAdminPermission10 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission11() bool {
	return p&unusedSiteAdminPermission11 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission12() bool {
	return p&unusedSiteAdminPermission12 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission13() bool {
	return p&unusedSiteAdminPermission13 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission14() bool {
	return p&unusedSiteAdminPermission14 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission15() bool {
	return p&unusedSiteAdminPermission15 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission16() bool {
	return p&unusedSiteAdminPermission16 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission17() bool {
	return p&unusedSiteAdminPermission17 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission18() bool {
	return p&unusedSiteAdminPermission18 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission19() bool {
	return p&unusedSiteAdminPermission19 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission20() bool {
	return p&unusedSiteAdminPermission20 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission21() bool {
	return p&unusedSiteAdminPermission21 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission22() bool {
	return p&unusedSiteAdminPermission22 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission23() bool {
	return p&unusedSiteAdminPermission23 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission24() bool {
	return p&unusedSiteAdminPermission24 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission25() bool {
	return p&unusedSiteAdminPermission25 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission26() bool {
	return p&unusedSiteAdminPermission26 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission27() bool {
	return p&unusedSiteAdminPermission27 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission28() bool {
	return p&unusedSiteAdminPermission28 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission29() bool {
	return p&unusedSiteAdminPermission29 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission30() bool {
	return p&unusedSiteAdminPermission30 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission31() bool {
	return p&unusedSiteAdminPermission31 != 0
}

func (p SiteAdminPermissions) hasReservedUnusedPermission32() bool {
	return p&unusedSiteAdminPermission32 != 0
}
