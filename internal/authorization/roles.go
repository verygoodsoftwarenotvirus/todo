package authorization

import (
	"fmt"
	"gopkg.in/mikespook/gorbac.v2"
)

type (
	role int
	// ServiceRole describes a role a user has for the Service context.
	ServiceRole role
	// AccountRole describes a role a user has for an Account context.
	AccountRole role
)

const (
	// invalidServiceRole is a service role to apply for non-admin users to have one.
	invalidServiceRole ServiceRole = iota
	// ServiceUserRole is a service role to apply for non-admin users to have one.
	ServiceUserRole ServiceRole = iota
	// ServiceAdminRole is a role that allows a user to do basically anything.
	ServiceAdminRole ServiceRole = iota
	// AccountMemberRole is a role for a plain account participant.
	AccountMemberRole AccountRole = iota
	// AccountAdminRole is a role for someone who can manipulate the details of an account.
	AccountAdminRole AccountRole = iota
)

const (
	serviceAdminRoleName  = "service_admin"
	serviceUserRoleName   = "service_user"
	accountAdminRoleName  = "account_admin"
	accountMemberRoleName = "account_member"
)

func (r ServiceRole) String() string {
	switch r {
	case ServiceUserRole:
		return serviceUserRoleName
	case ServiceAdminRole:
		return serviceAdminRoleName
	default:
		return ""
	}
}

func (r AccountRole) String() string {
	switch r {
	case AccountMemberRole:
		return accountMemberRoleName
	case AccountAdminRole:
		return accountAdminRoleName
	default:
		return ""
	}
}

func ServiceRoleFromString(s string) (ServiceRole, error) {
	switch s {
	case serviceAdminRoleName:
		return ServiceAdminRole, nil
	case serviceUserRoleName:
		return ServiceUserRole, nil
	default:
		return invalidServiceRole, fmt.Errorf("invalid service role %q", s)
	}
}

var (
	serviceUser   = gorbac.NewStdRole(serviceUserRoleName)
	serviceAdmin  = gorbac.NewStdRole(serviceAdminRoleName)
	accountAdmin  = gorbac.NewStdRole(accountAdminRoleName)
	accountMember = gorbac.NewStdRole(accountMemberRoleName)
)
