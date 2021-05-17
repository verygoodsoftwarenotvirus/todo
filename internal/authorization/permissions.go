package authorization

import "gopkg.in/mikespook/gorbac.v2"

type permission string

const (
	CycleCookieSecretPermission            permission = "update.cookie_secret"
	ReadAccountAuditLogEntriesPermission   permission = "read.account_audit_log_entries"
	ReadAPIClientAuditLogEntriesPermission permission = "read.api_client_audit_log_entries"
	ReadUserAuditLogEntriesPermission      permission = "read.user_audit_log_entries"
	ReadWebhookAuditLogEntriesPermission   permission = "read.webhook_audit_log_entries"
	UpdateUserReputationPermission         permission = "update.user_reputation"
	ReadUserPermission                     permission = "read.user"
	SearchUserPermission                   permission = "search.user"

	UpdateAccountPermission            permission = "update.account"
	DeleteAccountPermission            permission = "delete.account"
	AddMemberAccountPermission         permission = "add_member.account"
	RemoveMemberAccountPermission      permission = "remove_member.account"
	TransferAccountPermission          permission = "transfer.account"
	CreateWebhooksPermission           permission = "create.webhooks"
	ReadWebhooksPermission             permission = "read.webhooks"
	UpdateWebhooksPermission           permission = "update.webhooks"
	DeleteWebhooksPermission           permission = "delete.webhooks"
	CreateAPIClientsPermission         permission = "create.api_clients"
	ReadAPIClientsPermission           permission = "read.api_clients"
	DeleteAPIClientsPermission         permission = "delete.api_clients"
	ReadItemsAuditLogEntriesPermission permission = "read.items_audit_log_entries"

	CreateItemsPermission permission = "create.items"
	ReadItemsPermission   permission = "read.items"
	SearchItemsPermission permission = "search.items"
	UpdateItemsPermission permission = "update.items"
	DeleteItemsPermission permission = "delete.items"
)

func (p permission) ID() string {
	return string(p)
}

func (p permission) Match(perm gorbac.Permission) bool {
	return p.ID() == perm.ID()
}

var (
	// service admin permissions
	serviceAdminPermissions = map[string]gorbac.Permission{
		ReadAccountAuditLogEntriesPermission.ID():   ReadAccountAuditLogEntriesPermission,
		ReadAPIClientAuditLogEntriesPermission.ID(): ReadAPIClientAuditLogEntriesPermission,
		ReadUserAuditLogEntriesPermission.ID():      ReadUserAuditLogEntriesPermission,
		ReadWebhookAuditLogEntriesPermission.ID():   ReadWebhookAuditLogEntriesPermission,
		UpdateUserReputationPermission.ID():         UpdateUserReputationPermission,
		ReadUserPermission.ID():                     ReadUserPermission,
		SearchUserPermission.ID():                   SearchUserPermission,
	}

	// account admin permissions
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

	// account member permissions
	accountMemberPermissions = map[string]gorbac.Permission{
		CreateItemsPermission.ID(): CreateItemsPermission,
		ReadItemsPermission.ID():   ReadItemsPermission,
		SearchItemsPermission.ID(): SearchItemsPermission,
		UpdateItemsPermission.ID(): UpdateItemsPermission,
		DeleteItemsPermission.ID(): DeleteItemsPermission,
	}
)

func init() {
	// assign service admin permissions
	for _, perm := range serviceAdminPermissions {
		must(serviceAdmin.Assign(perm))
	}

	// assign account admin permissions
	for _, perm := range accountAdminPermissions {
		must(accountAdmin.Assign(perm))
	}

	// assign account member permissions
	for _, perm := range accountMemberPermissions {
		must(accountMember.Assign(perm))
	}
}
