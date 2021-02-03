package querybuilding

const (
	//
	// Common Columns.
	//

	// ExistencePrefix goes before a sql query.
	ExistencePrefix = "SELECT EXISTS ("
	// ExistenceSuffix goes after a sql query.
	ExistenceSuffix = ")"

	// IDColumn is a common column name for the sequential ID column.
	IDColumn = "id"
	// ExternalIDColumn is a common column name for the external ID column.
	ExternalIDColumn = "external_id"
	// CreatedOnColumn is a common column name for the row creation time column.
	CreatedOnColumn = "created_on"
	// LastUpdatedOnColumn is a common column name for the latest row update column.
	LastUpdatedOnColumn = "last_updated_on"
	// ArchivedOnColumn is a common column name for the archive time column.
	ArchivedOnColumn       = "archived_on"
	commaSeparator         = ","
	userOwnershipColumn    = "belongs_to_user"
	accountOwnershipColumn = "belongs_to_account"

	//
	// Accounts Table.
	//

	// AccountsTableName is what the accounts table calls itself.
	AccountsTableName = "accounts"
	// AccountsTableNameColumn is what the accounts table calls the name column.
	AccountsTableNameColumn = "name"
	// AccountsTablePlanIDColumn is what the accounts table calls the <> column.
	AccountsTablePlanIDColumn = "plan_id"
	// AccountsTablePersonalAccountColumn is what the accounts table calls the user account permissions column.
	AccountsTablePersonalAccountColumn = "is_personal_account"
	// AccountsTableUserOwnershipColumn is what the accounts table calls the user ownership column.
	AccountsTableUserOwnershipColumn = userOwnershipColumn

	//
	// Accounts Membership Table.
	//

	// AccountsMembershipTableName is what the accounts membership table calls itself.
	AccountsMembershipTableName = "accounts_membership"
	// AccountsMembershipTablePrimaryUserAccountColumn is what the accounts membership table calls the column indicating primary status.
	AccountsMembershipTablePrimaryUserAccountColumn = "is_primary_user_account"
	// AccountsMembershipTableUserAccountPermissionsColumn is what the accounts membership table calls the user account permissions column.
	AccountsMembershipTableUserAccountPermissionsColumn = "user_account_permissions"
	// AccountsMembershipTableAccountOwnershipColumn is what the accounts membership table calls the user ownership column.
	AccountsMembershipTableAccountOwnershipColumn = accountOwnershipColumn
	// AccountsMembershipTableUserOwnershipColumn is what the accounts membership table calls the user ownership column.
	AccountsMembershipTableUserOwnershipColumn = userOwnershipColumn

	//
	// AccountSubscriptionPlans Table.
	//

	// AccountSubscriptionPlansTableName is what the users table calls the <> column.
	AccountSubscriptionPlansTableName = "account_subscription_plans"
	// AccountSubscriptionPlansTableNameColumn is what the users table calls the <> column.
	AccountSubscriptionPlansTableNameColumn = "name"
	// AccountSubscriptionPlansTableDescriptionColumn is what the users table calls the <> column.
	AccountSubscriptionPlansTableDescriptionColumn = "description"
	// AccountSubscriptionPlansTablePriceColumn is what the users table calls the <> column.
	AccountSubscriptionPlansTablePriceColumn = "price"
	// AccountSubscriptionPlansTablePeriodColumn is what the users table calls the <> column.
	AccountSubscriptionPlansTablePeriodColumn = "period"

	//
	// Users Table.
	//

	// UsersTableName is what the users table calls the <> column.
	UsersTableName = "users"
	// UsersTableUsernameColumn is what the users table calls the <> column.
	UsersTableUsernameColumn = "username"
	// UsersTableHashedPasswordColumn is what the users table calls the <> column.
	UsersTableHashedPasswordColumn = "hashed_password"
	// UsersTableSaltColumn is what the users table calls the <> column.
	UsersTableSaltColumn = "salt"
	// UsersTableRequiresPasswordChangeColumn is what the users table calls the <> column.
	UsersTableRequiresPasswordChangeColumn = "requires_password_change"
	// UsersTablePasswordLastChangedOnColumn is what the users table calls the <> column.
	UsersTablePasswordLastChangedOnColumn = "password_last_changed_on"
	// UsersTableTwoFactorSekretColumn is what the users table calls the <> column.
	UsersTableTwoFactorSekretColumn = "two_factor_secret"
	// UsersTableTwoFactorVerifiedOnColumn is what the users table calls the <> column.
	UsersTableTwoFactorVerifiedOnColumn = "two_factor_secret_verified_on"
	// UsersTableIsAdminColumn is what the users table calls the <> column.
	UsersTableIsAdminColumn = "is_site_admin"
	// UsersTableAdminPermissionsColumn is what the users table calls the <> column.
	UsersTableAdminPermissionsColumn = "site_admin_permissions"
	// UsersTableReputationColumn is what the users table calls the <> column.
	UsersTableReputationColumn = "reputation"
	// UsersTableStatusExplanationColumn is what the users table calls the <> column.
	UsersTableStatusExplanationColumn = "reputation_explanation"
	// UsersTableAvatarColumn is what the users table calls the <> column.
	UsersTableAvatarColumn = "avatar_src"

	//
	// Audit Log Entries Table.
	//

	// AuditLogEntriesTableName is what the audit log entries table calls itself.
	AuditLogEntriesTableName = "audit_log"
	// AuditLogEntriesTableEventTypeColumn is what the audit log entries table calls the event type column.
	AuditLogEntriesTableEventTypeColumn = "event_type"
	// AuditLogEntriesTableContextColumn is what the audit log entries table calls the context column.
	AuditLogEntriesTableContextColumn = "context"

	//
	// Delegated Clients.
	//

	// DelegatedClientsTableScopeSeparator is what the oauth2 clients table calls the <> column.
	DelegatedClientsTableScopeSeparator = commaSeparator
	// DelegatedClientsTableName is what the oauth2 clients table calls the <> column.
	DelegatedClientsTableName = "delegated_clients"
	// DelegatedClientsTableNameColumn is what the oauth2 clients table calls the <> column.
	DelegatedClientsTableNameColumn = "name"
	// DelegatedClientsTableClientIDColumn is what the oauth2 clients table calls the <> column.
	DelegatedClientsTableClientIDColumn = "client_id"
	// DelegatedClientsTableClientSecretColumn is what the oauth2 clients table calls the <> column.
	DelegatedClientsTableClientSecretColumn = "client_secret"
	// DelegatedClientsTableOwnershipColumn is what the oauth2 clients table calls the <> column.
	DelegatedClientsTableOwnershipColumn = userOwnershipColumn

	//
	// OAuth2 Clients.
	//

	// OAuth2ClientsTableScopeSeparator is what the oauth2 clients table calls the <> column.
	OAuth2ClientsTableScopeSeparator = commaSeparator
	// OAuth2ClientsTableName is what the oauth2 clients table calls the <> column.
	OAuth2ClientsTableName = "oauth2_clients"
	// OAuth2ClientsTableNameColumn is what the oauth2 clients table calls the <> column.
	OAuth2ClientsTableNameColumn = "name"
	// OAuth2ClientsTableClientIDColumn is what the oauth2 clients table calls the <> column.
	OAuth2ClientsTableClientIDColumn = "client_id"
	// OAuth2ClientsTableScopesColumn is what the oauth2 clients table calls the <> column.
	OAuth2ClientsTableScopesColumn = "scopes"
	// OAuth2ClientsTableRedirectURIColumn is what the oauth2 clients table calls the <> column.
	OAuth2ClientsTableRedirectURIColumn = "redirect_uri"
	// OAuth2ClientsTableClientSecretColumn is what the oauth2 clients table calls the <> column.
	OAuth2ClientsTableClientSecretColumn = "client_secret"
	// OAuth2ClientsTableOwnershipColumn is what the oauth2 clients table calls the <> column.
	OAuth2ClientsTableOwnershipColumn = userOwnershipColumn

	//
	// Webhooks Table.
	//

	// WebhooksTableName is what the webhooks table calls the <> column.
	WebhooksTableName = "webhooks"
	// WebhooksTableNameColumn is what the webhooks table calls the <> column.
	WebhooksTableNameColumn = "name"
	// WebhooksTableContentTypeColumn is what the webhooks table calls the <> column.
	WebhooksTableContentTypeColumn = "content_type"
	// WebhooksTableURLColumn is what the webhooks table calls the <> column.
	WebhooksTableURLColumn = "url"
	// WebhooksTableMethodColumn is what the webhooks table calls the <> column.
	WebhooksTableMethodColumn = "method"
	// WebhooksTableEventsColumn is what the webhooks table calls the <> column.
	WebhooksTableEventsColumn = "events"
	// WebhooksTableEventsSeparator is what the webhooks table calls the <> column.
	WebhooksTableEventsSeparator = commaSeparator
	// WebhooksTableDataTypesColumn is what the webhooks table calls the <> column.
	WebhooksTableDataTypesColumn = "data_types"
	// WebhooksTableDataTypesSeparator is what the webhooks table calls the <> column.
	WebhooksTableDataTypesSeparator = commaSeparator
	// WebhooksTableTopicsColumn is what the webhooks table calls the <> column.
	WebhooksTableTopicsColumn = "topics"
	// WebhooksTableTopicsSeparator is what the webhooks table calls the <> column.
	WebhooksTableTopicsSeparator = commaSeparator
	// WebhooksTableOwnershipColumn is what the webhooks table calls the <> column.
	WebhooksTableOwnershipColumn = "belongs_to_user"

	//
	// Items Table.
	//

	// ItemsTableName is what the items table calls itself.
	ItemsTableName = "items"
	// ItemsTableNameColumn is what the items table calls the name column.
	ItemsTableNameColumn = "name"
	// ItemsTableDetailsColumn is what the items table calls the details column.
	ItemsTableDetailsColumn = "details"
	// ItemsTableUserOwnershipColumn is what the items table calls the user ownership column.
	ItemsTableUserOwnershipColumn = userOwnershipColumn
)
