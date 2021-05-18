package authorization

import "gopkg.in/mikespook/gorbac.v2"

type (
	// AccountRole describes a role a user has for an Account context.
	AccountRole role
)

const (
	// AccountMemberRole is a role for a plain account participant.
	AccountMemberRole AccountRole = iota
	// AccountAdminRole is a role for someone who can manipulate the details of an account.
	AccountAdminRole AccountRole = iota

	accountAdminRoleName  = "account_admin"
	accountMemberRoleName = "account_member"
)

var (
	accountAdmin  = gorbac.NewStdRole(accountAdminRoleName)
	accountMember = gorbac.NewStdRole(accountMemberRoleName)
)

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
