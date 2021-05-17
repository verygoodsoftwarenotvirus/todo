package authorization

import (
	"strings"

	"gopkg.in/mikespook/gorbac.v2"
)

type (
	role int
)

const (
	invalidRole role = iota
	ServiceAdminRole
	AccountAdminRole
	AccountMemberRole
)

const (
	ServiceAdminRoleName  = "service_admin"
	AccountAdminRoleName  = "account_admin"
	AccountMemberRoleName = "account_member"
)

func roleIsValid(s string) bool {
	x := strings.TrimSpace(strings.ToLower(s))

	return x == ServiceAdminRoleName ||
		x == AccountAdminRoleName ||
		x == AccountMemberRoleName
}

func (r role) Name() string {
	switch r {
	case ServiceAdminRole:
		return ServiceAdminRoleName
	case AccountAdminRole:
		return AccountAdminRoleName
	case AccountMemberRole:
		return AccountMemberRoleName
	default:
		return ""
	}
}

var (
	serviceAdmin  = gorbac.NewStdRole(ServiceAdminRole.Name())
	accountAdmin  = gorbac.NewStdRole(AccountAdminRole.Name())
	accountMember = gorbac.NewStdRole(AccountMemberRole.Name())
)
