package authorization

import (
	"encoding/gob"

	"gopkg.in/mikespook/gorbac.v2"
)

type (
	// AccountRole describes a role a user has for an Account context.
	AccountRole role

	// AccountRolePermissionsChecker checks permissions for one or more account Roles.
	AccountRolePermissionsChecker interface {
		CanUpdateAccounts() bool
		CanDeleteAccounts() bool
		CanAddMemberToAccounts() bool
		CanRemoveMemberFromAccounts() bool
		CanTransferAccountToNewOwner() bool
		CanCreateWebhooks() bool
		CanSeeWebhooks() bool
		CanUpdateWebhooks() bool
		CanDeleteWebhooks() bool
		CanCreateAPIClients() bool
		CanSeeAPIClients() bool
		CanDeleteAPIClients() bool
		CanSeeItemsAuditLogEntries() bool
	}
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

type accountRoleCollection struct {
	Roles []string
}

func init() {
	gob.Register(accountRoleCollection{})
}

// NewAccountRolePermissionChecker returns a new checker for a set of Roles.
func NewAccountRolePermissionChecker(roles ...string) AccountRolePermissionsChecker {
	return &accountRoleCollection{
		Roles: roles,
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

// CanUpdateAccounts returns whether a user can update accounts or not.
func (r accountRoleCollection) CanUpdateAccounts() bool {
	return hasPermission(UpdateAccountPermission, r.Roles...)
}

// CanDeleteAccounts returns whether a user can delete accounts or not.
func (r accountRoleCollection) CanDeleteAccounts() bool {
	return hasPermission(DeleteAccountPermission, r.Roles...)
}

// CanAddMemberToAccounts returns whether a user can add members to accounts or not.
func (r accountRoleCollection) CanAddMemberToAccounts() bool {
	return hasPermission(AddMemberAccountPermission, r.Roles...)
}

// CanRemoveMemberFromAccounts returns whether a user can remove members from accounts or not.
func (r accountRoleCollection) CanRemoveMemberFromAccounts() bool {
	return hasPermission(RemoveMemberAccountPermission, r.Roles...)
}

// CanTransferAccountToNewOwner returns whether a user can transfer an account to a new owner or not.
func (r accountRoleCollection) CanTransferAccountToNewOwner() bool {
	return hasPermission(TransferAccountPermission, r.Roles...)
}

// CanCreateWebhooks returns whether a user can create webhooks or not.
func (r accountRoleCollection) CanCreateWebhooks() bool {
	return hasPermission(CreateWebhooksPermission, r.Roles...)
}

// CanSeeWebhooks returns whether a user can view webhooks or not.
func (r accountRoleCollection) CanSeeWebhooks() bool {
	return hasPermission(ReadWebhooksPermission, r.Roles...)
}

// CanUpdateWebhooks returns whether a user can update webhooks or not.
func (r accountRoleCollection) CanUpdateWebhooks() bool {
	return hasPermission(UpdateWebhooksPermission, r.Roles...)
}

// CanDeleteWebhooks returns whether a user can delete webhooks or not.
func (r accountRoleCollection) CanDeleteWebhooks() bool {
	return hasPermission(DeleteWebhooksPermission, r.Roles...)
}

// CanCreateAPIClients returns whether a user can create API clients or not.
func (r accountRoleCollection) CanCreateAPIClients() bool {
	return hasPermission(CreateAPIClientsPermission, r.Roles...)
}

// CanSeeAPIClients returns whether a user can view API clients or not.
func (r accountRoleCollection) CanSeeAPIClients() bool {
	return hasPermission(ReadAPIClientsPermission, r.Roles...)
}

// CanDeleteAPIClients returns whether a user can delete API clients or not.
func (r accountRoleCollection) CanDeleteAPIClients() bool {
	return hasPermission(DeleteAPIClientsPermission, r.Roles...)
}

// CanSeeItemsAuditLogEntries returns whether a user can view item audit log entries or not.
func (r accountRoleCollection) CanSeeItemsAuditLogEntries() bool {
	return hasPermission(ReadItemsAuditLogEntriesPermission, r.Roles...)
}
