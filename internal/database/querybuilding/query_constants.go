package querybuilding

var (
	// DefaultTestUserSalt is the default Salt we give to test users when we initialize them.
	DefaultTestUserSalt = []byte("aaaaaaaaaaaaaaaa")
)

const (
	// DefaultTestUserTwoFactorSecret is the default TwoFactorSecret we give to test users when we initialize them.
	// `otpauth://totp/todo:username?secret=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=&issuer=todo`
	DefaultTestUserTwoFactorSecret = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

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
	// AccountsTableDefaultUserPermissionsColumn is what the accounts table calls the <> column.
	AccountsTableDefaultUserPermissionsColumn = "default_user_permissions"
	// AccountsTableUserOwnershipColumn is what the accounts table calls the user ownership column.
	AccountsTableUserOwnershipColumn = userOwnershipColumn

	//
	// Accounts Membership Table.
	//

	// AccountsUserMembershipTableName is what the accounts membership table calls itself.
	AccountsUserMembershipTableName = "account_user_memberships"
	// AccountsUserMembershipTableUserPermissionsColumn is what the accounts membership table calls the column indicating user permissions.
	AccountsUserMembershipTableUserPermissionsColumn = "user_account_permissions"
	// AccountsUserMembershipTableAccountRoleColumn is what the accounts membership table calls the column indicating account role.
	AccountsUserMembershipTableAccountRoleColumn = "account_role"
	// AccountsUserMembershipTableAccountOwnershipColumn is what the accounts membership table calls the user ownership column.
	AccountsUserMembershipTableAccountOwnershipColumn = accountOwnershipColumn
	// AccountsUserMembershipTableUserOwnershipColumn is what the accounts membership table calls the user ownership column.
	AccountsUserMembershipTableUserOwnershipColumn = userOwnershipColumn
	// AccountsUserMembershipTableDefaultUserAccountColumn is what the accounts membership table calls the .
	AccountsUserMembershipTableDefaultUserAccountColumn = "default_account"

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
	// UsersTableRequiresPasswordChangeColumn is what the users table calls the <> column.
	UsersTableRequiresPasswordChangeColumn = "requires_password_change"
	// UsersTablePasswordLastChangedOnColumn is what the users table calls the <> column.
	UsersTablePasswordLastChangedOnColumn = "password_last_changed_on"
	// UsersTableTwoFactorSekretColumn is what the users table calls the <> column.
	UsersTableTwoFactorSekretColumn = "two_factor_secret"
	// UsersTableTwoFactorVerifiedOnColumn is what the users table calls the <> column.
	UsersTableTwoFactorVerifiedOnColumn = "two_factor_secret_verified_on"
	// UsersTableServiceRoleColumn is what the users table calls the <> column.
	UsersTableServiceRoleColumn = "service_role"
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
	// API Clients.
	//

	// APIClientsTableName is what the API clients table calls the <> column.
	APIClientsTableName = "api_clients"
	// APIClientsTableNameColumn is what the API clients table calls the <> column.
	APIClientsTableNameColumn = "name"
	// APIClientsTableClientIDColumn is what the API clients table calls the <> column.
	APIClientsTableClientIDColumn = "client_id"
	// APIClientsTableSecretKeyColumn is what the API clients table calls the <> column.
	APIClientsTableSecretKeyColumn = "secret_key"
	// APIClientsTableOwnershipColumn is what the API clients table calls the <> column.
	APIClientsTableOwnershipColumn = userOwnershipColumn

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
	WebhooksTableOwnershipColumn = accountOwnershipColumn

	//
	// Items Table.
	//

	// ItemsTableName is what the items table calls itself.
	ItemsTableName = "items"
	// ItemsTableNameColumn is what the items table calls the name column.
	ItemsTableNameColumn = "name"
	// ItemsTableDetailsColumn is what the items table calls the details column.
	ItemsTableDetailsColumn = "details"
	// ItemsTableAccountOwnershipColumn is what the items table calls the user ownership column.
	ItemsTableAccountOwnershipColumn = accountOwnershipColumn
)
