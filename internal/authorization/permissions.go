package authorization

import (
	"gopkg.in/mikespook/gorbac.v2"
)

type (
	permission string
)

const (
	// CycleCookieSecretPermission is a service admin permission.
	CycleCookieSecretPermission permission = "update.cookie_secret"
	// ReadAccountAuditLogEntriesPermission is a service admin permission.
	ReadAccountAuditLogEntriesPermission permission = "read.account_audit_log_entries"
	// ReadAPIClientAuditLogEntriesPermission is a service admin permission.
	ReadAPIClientAuditLogEntriesPermission permission = "read.api_client_audit_log_entries"
	// ReadUserAuditLogEntriesPermission is a service admin permission.
	ReadUserAuditLogEntriesPermission permission = "read.user_audit_log_entries"
	// ReadWebhookAuditLogEntriesPermission is a service admin permission.
	ReadWebhookAuditLogEntriesPermission permission = "read.webhook_audit_log_entries"
	// UpdateUserReputationPermission is a service admin permission.
	UpdateUserReputationPermission permission = "update.user_reputation"
	// ReadUserPermission is a service admin permission.
	ReadUserPermission permission = "read.user"
	// SearchUserPermission is a service admin permission.
	SearchUserPermission permission = "search.user"

	// UpdateAccountPermission is an account admin permission.
	UpdateAccountPermission permission = "update.account"
	// DeleteAccountPermission is an account admin permission.
	DeleteAccountPermission permission = "delete.account"
	// AddMemberAccountPermission is an account admin permission.
	AddMemberAccountPermission permission = "add_member.account"
	// RemoveMemberAccountPermission is an account admin permission.
	RemoveMemberAccountPermission permission = "remove_member.account"
	// TransferAccountPermission is an account admin permission.
	TransferAccountPermission permission = "transfer.account"
	// CreateWebhooksPermission is an account admin permission.
	CreateWebhooksPermission permission = "create.webhooks"
	// ReadWebhooksPermission is an account admin permission.
	ReadWebhooksPermission permission = "read.webhooks"
	// UpdateWebhooksPermission is an account admin permission.
	UpdateWebhooksPermission permission = "update.webhooks"
	// DeleteWebhooksPermission is an account admin permission.
	DeleteWebhooksPermission permission = "delete.webhooks"
	// CreateAPIClientsPermission is an account admin permission.
	CreateAPIClientsPermission permission = "create.api_clients"
	// ReadAPIClientsPermission is an account admin permission.
	ReadAPIClientsPermission permission = "read.api_clients"
	// DeleteAPIClientsPermission is an account admin permission.
	DeleteAPIClientsPermission permission = "delete.api_clients"
	// ReadItemsAuditLogEntriesPermission is an account admin permission.
	ReadItemsAuditLogEntriesPermission permission = "read.items_audit_log_entries"

	// CreateItemsPermission is an account user permission.
	CreateItemsPermission permission = "create.items"
	// ReadItemsPermission is an account user permission.
	ReadItemsPermission permission = "read.items"
	// SearchItemsPermission is an account user permission.
	SearchItemsPermission permission = "search.items"
	// UpdateItemsPermission is an account user permission.
	UpdateItemsPermission permission = "update.items"
	// DeleteItemsPermission is an account user permission.
	DeleteItemsPermission permission = "delete.items"
)

func (p permission) ID() string {
	return string(p)
}

func (p permission) Match(perm gorbac.Permission) bool {
	return p.ID() == perm.ID()
}

var (
	// service admin permissions.
	serviceAdminPermissions = map[string]gorbac.Permission{
		ReadAccountAuditLogEntriesPermission.ID():   ReadAccountAuditLogEntriesPermission,
		ReadAPIClientAuditLogEntriesPermission.ID(): ReadAPIClientAuditLogEntriesPermission,
		ReadUserAuditLogEntriesPermission.ID():      ReadUserAuditLogEntriesPermission,
		ReadWebhookAuditLogEntriesPermission.ID():   ReadWebhookAuditLogEntriesPermission,
		UpdateUserReputationPermission.ID():         UpdateUserReputationPermission,
		ReadUserPermission.ID():                     ReadUserPermission,
		SearchUserPermission.ID():                   SearchUserPermission,
	}

	// account admin permissions.
	accountAdminPermissions = map[string]gorbac.Permission{
		UpdateAccountPermission.ID():            UpdateAccountPermission,
		DeleteAccountPermission.ID():            DeleteAccountPermission,
		AddMemberAccountPermission.ID():         AddMemberAccountPermission,
		RemoveMemberAccountPermission.ID():      RemoveMemberAccountPermission,
		TransferAccountPermission.ID():          TransferAccountPermission,
		CreateWebhooksPermission.ID():           CreateWebhooksPermission,
		ReadWebhooksPermission.ID():             ReadWebhooksPermission,
		UpdateWebhooksPermission.ID():           UpdateWebhooksPermission,
		DeleteWebhooksPermission.ID():           DeleteWebhooksPermission,
		CreateAPIClientsPermission.ID():         CreateAPIClientsPermission,
		ReadAPIClientsPermission.ID():           ReadAPIClientsPermission,
		DeleteAPIClientsPermission.ID():         DeleteAPIClientsPermission,
		ReadItemsAuditLogEntriesPermission.ID(): ReadItemsAuditLogEntriesPermission,
	}

	// account member permissions.
	accountMemberPermissions = map[string]gorbac.Permission{
		CreateItemsPermission.ID(): CreateItemsPermission,
		ReadItemsPermission.ID():   ReadItemsPermission,
		SearchItemsPermission.ID(): SearchItemsPermission,
		UpdateItemsPermission.ID(): UpdateItemsPermission,
		DeleteItemsPermission.ID(): DeleteItemsPermission,
	}
)

func init() {
	// assign service admin permissions.
	for _, perm := range serviceAdminPermissions {
		must(serviceAdmin.Assign(perm))
	}

	// assign account admin permissions.
	for _, perm := range accountAdminPermissions {
		must(accountAdmin.Assign(perm))
	}

	// assign account member permissions.
	for _, perm := range accountMemberPermissions {
		must(accountMember.Assign(perm))
	}
}
