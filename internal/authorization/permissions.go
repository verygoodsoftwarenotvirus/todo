package authorization

import (
	"gopkg.in/mikespook/gorbac.v2"
)

type (
	// Permission is a simple string alias.
	Permission string
)

const (
	// CycleCookieSecretPermission is a service admin permission.
	CycleCookieSecretPermission Permission = "update.cookie_secret"
	// ReadAllAuditLogEntriesPermission is a service admin permission.
	ReadAllAuditLogEntriesPermission Permission = "read.audit_log_entries.all"
	// ReadAccountAuditLogEntriesPermission is a service admin permission.
	ReadAccountAuditLogEntriesPermission Permission = "read.audit_log_entries.account"
	// ReadAPIClientAuditLogEntriesPermission is a service admin permission.
	ReadAPIClientAuditLogEntriesPermission Permission = "read.audit_log_entries.api_client"
	// ReadUserAuditLogEntriesPermission is a service admin permission.
	ReadUserAuditLogEntriesPermission Permission = "read.audit_log_entries.user"
	// ReadWebhookAuditLogEntriesPermission is a service admin permission.
	ReadWebhookAuditLogEntriesPermission Permission = "read.audit_log_entries.webhook"
	// UpdateUserStatusPermission is a service admin permission.
	UpdateUserStatusPermission Permission = "update.user_status"
	// ReadUserPermission is a service admin permission.
	ReadUserPermission Permission = "read.user"
	// SearchUserPermission is a service admin permission.
	SearchUserPermission Permission = "search.user"

	// UpdateAccountPermission is an account admin permission.
	UpdateAccountPermission Permission = "update.account"
	// ArchiveAccountPermission is an account admin permission.
	ArchiveAccountPermission Permission = "archive.account"
	// AddMemberAccountPermission is an account admin permission.
	AddMemberAccountPermission Permission = "account.add.member"
	// ModifyMemberPermissionsForAccountPermission is an account admin permission.
	ModifyMemberPermissionsForAccountPermission Permission = "account.membership.modify"
	// RemoveMemberAccountPermission is an account admin permission.
	RemoveMemberAccountPermission Permission = "remove_member.account"
	// TransferAccountPermission is an account admin permission.
	TransferAccountPermission Permission = "transfer.account"
	// CreateWebhooksPermission is an account admin permission.
	CreateWebhooksPermission Permission = "create.webhooks"
	// ReadWebhooksPermission is an account admin permission.
	ReadWebhooksPermission Permission = "read.webhooks"
	// UpdateWebhooksPermission is an account admin permission.
	UpdateWebhooksPermission Permission = "update.webhooks"
	// ArchiveWebhooksPermission is an account admin permission.
	ArchiveWebhooksPermission Permission = "archive.webhooks"
	// CreateAPIClientsPermission is an account admin permission.
	CreateAPIClientsPermission Permission = "create.api_clients"
	// ReadAPIClientsPermission is an account admin permission.
	ReadAPIClientsPermission Permission = "read.api_clients"
	// ArchiveAPIClientsPermission is an account admin permission.
	ArchiveAPIClientsPermission Permission = "archive.api_clients"
	// ReadItemsAuditLogEntriesPermission is an account admin permission.
	ReadItemsAuditLogEntriesPermission Permission = "read.audit_log_entries.items"
	// ReadWebhooksAuditLogEntriesPermission is an account admin permission.
	ReadWebhooksAuditLogEntriesPermission Permission = "read.audit_log_entries.webhooks"

	// CreateItemsPermission is an account user permission.
	CreateItemsPermission Permission = "create.items"
	// ReadItemsPermission is an account user permission.
	ReadItemsPermission Permission = "read.items"
	// SearchItemsPermission is an account user permission.
	SearchItemsPermission Permission = "search.items"
	// UpdateItemsPermission is an account user permission.
	UpdateItemsPermission Permission = "update.items"
	// ArchiveItemsPermission is an account user permission.
	ArchiveItemsPermission Permission = "archive.items"
)

// ID implements the gorbac Permission interface.
func (p Permission) ID() string {
	return string(p)
}

// Match implements the gorbac Permission interface.
func (p Permission) Match(perm gorbac.Permission) bool {
	return p.ID() == perm.ID()
}

var (
	// service admin permissions.
	serviceAdminPermissions = map[string]gorbac.Permission{
		CycleCookieSecretPermission.ID():            CycleCookieSecretPermission,
		ReadAllAuditLogEntriesPermission.ID():       ReadAllAuditLogEntriesPermission,
		ReadAccountAuditLogEntriesPermission.ID():   ReadAccountAuditLogEntriesPermission,
		ReadAPIClientAuditLogEntriesPermission.ID(): ReadAPIClientAuditLogEntriesPermission,
		ReadUserAuditLogEntriesPermission.ID():      ReadUserAuditLogEntriesPermission,
		ReadWebhookAuditLogEntriesPermission.ID():   ReadWebhookAuditLogEntriesPermission,
		UpdateUserStatusPermission.ID():             UpdateUserStatusPermission,
		ReadUserPermission.ID():                     ReadUserPermission,
		SearchUserPermission.ID():                   SearchUserPermission,
	}

	// account admin permissions.
	accountAdminPermissions = map[string]gorbac.Permission{
		UpdateAccountPermission.ID():                     UpdateAccountPermission,
		ArchiveAccountPermission.ID():                    ArchiveAccountPermission,
		AddMemberAccountPermission.ID():                  AddMemberAccountPermission,
		ModifyMemberPermissionsForAccountPermission.ID(): ModifyMemberPermissionsForAccountPermission,
		RemoveMemberAccountPermission.ID():               RemoveMemberAccountPermission,
		TransferAccountPermission.ID():                   TransferAccountPermission,
		CreateWebhooksPermission.ID():                    CreateWebhooksPermission,
		ReadWebhooksPermission.ID():                      ReadWebhooksPermission,
		UpdateWebhooksPermission.ID():                    UpdateWebhooksPermission,
		ArchiveWebhooksPermission.ID():                   ArchiveWebhooksPermission,
		CreateAPIClientsPermission.ID():                  CreateAPIClientsPermission,
		ReadAPIClientsPermission.ID():                    ReadAPIClientsPermission,
		ArchiveAPIClientsPermission.ID():                 ArchiveAPIClientsPermission,
		ReadItemsAuditLogEntriesPermission.ID():          ReadItemsAuditLogEntriesPermission,
		ReadWebhooksAuditLogEntriesPermission.ID():       ReadWebhooksAuditLogEntriesPermission,
	}

	// account member permissions.
	accountMemberPermissions = map[string]gorbac.Permission{
		CreateItemsPermission.ID():  CreateItemsPermission,
		ReadItemsPermission.ID():    ReadItemsPermission,
		SearchItemsPermission.ID():  SearchItemsPermission,
		UpdateItemsPermission.ID():  UpdateItemsPermission,
		ArchiveItemsPermission.ID(): ArchiveItemsPermission,
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
