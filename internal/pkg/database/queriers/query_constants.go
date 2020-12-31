package queriers

import "fmt"

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
	// CreatedOnColumn is a common column name for the row creation time column.
	CreatedOnColumn = "created_on"
	// LastUpdatedOnColumn is a common column name for the latest row update column.
	LastUpdatedOnColumn = "last_updated_on"
	// ArchivedOnColumn is a common column name for the archive time column.
	ArchivedOnColumn    = "archived_on"
	commaSeparator      = ","
	userOwnershipColumn = "belongs_to_user"

	//
	// Plans Table.
	//

	// PlansTableName is what the users table calls the <> column.
	PlansTableName = "plans"
	// PlansTableNameColumn is what the users table calls the <> column.
	PlansTableNameColumn = "name"
	// PlansTableDescriptionColumn is what the users table calls the <> column.
	PlansTableDescriptionColumn = "description"
	// PlansTablePriceColumn is what the users table calls the <> column.
	PlansTablePriceColumn = "price"
	// PlansTablePeriodColumn is what the users table calls the <> column.
	PlansTablePeriodColumn = "period"

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
	// UsersTableTwoFactorColumn is what the users table calls the <> column.
	UsersTableTwoFactorColumn = "two_factor_secret"
	// UsersTableTwoFactorVerifiedOnColumn is what the users table calls the <> column.
	UsersTableTwoFactorVerifiedOnColumn = "two_factor_secret_verified_on"
	// UsersTableIsAdminColumn is what the users table calls the <> column.
	UsersTableIsAdminColumn = "is_admin"
	// UsersTableAdminPermissionsColumn is what the users table calls the <> column.
	UsersTableAdminPermissionsColumn = "admin_permissions"
	// UsersTableAccountStatusColumn is what the users table calls the <> column.
	UsersTableAccountStatusColumn = "account_status"
	// UsersTableStatusExplanationColumn is what the users table calls the <> column.
	UsersTableStatusExplanationColumn = "status_explanation"
	// UsersTablePlanIDColumn is what the users table calls the <> column.
	UsersTablePlanIDColumn = "plan_id"
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

var (
	//
	// Plans Table.
	//

	// PlansTableColumns are the columns for the users table.
	PlansTableColumns = []string{
		fmt.Sprintf("%s.%s", PlansTableName, IDColumn),
		fmt.Sprintf("%s.%s", PlansTableName, PlansTableNameColumn),
		fmt.Sprintf("%s.%s", PlansTableName, PlansTableDescriptionColumn),
		fmt.Sprintf("%s.%s", PlansTableName, PlansTablePriceColumn),
		fmt.Sprintf("%s.%s", PlansTableName, PlansTablePeriodColumn),
		fmt.Sprintf("%s.%s", PlansTableName, CreatedOnColumn),
		fmt.Sprintf("%s.%s", PlansTableName, LastUpdatedOnColumn),
		fmt.Sprintf("%s.%s", PlansTableName, ArchivedOnColumn),
	}
	//
	// Users Table.
	//

	// UsersTableColumns are the columns for the users table.
	UsersTableColumns = []string{
		fmt.Sprintf("%s.%s", UsersTableName, IDColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableUsernameColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableAvatarColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableHashedPasswordColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableSaltColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableRequiresPasswordChangeColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTablePasswordLastChangedOnColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableTwoFactorColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableTwoFactorVerifiedOnColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableIsAdminColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableAdminPermissionsColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableAccountStatusColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableStatusExplanationColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTablePlanIDColumn),
		fmt.Sprintf("%s.%s", UsersTableName, CreatedOnColumn),
		fmt.Sprintf("%s.%s", UsersTableName, LastUpdatedOnColumn),
		fmt.Sprintf("%s.%s", UsersTableName, ArchivedOnColumn),
	}

	//
	// Audit Log Entries Table.
	//

	// AuditLogEntriesTableColumns are the columns for the audit log entries table.
	AuditLogEntriesTableColumns = []string{
		fmt.Sprintf("%s.%s", AuditLogEntriesTableName, IDColumn),
		fmt.Sprintf("%s.%s", AuditLogEntriesTableName, AuditLogEntriesTableEventTypeColumn),
		fmt.Sprintf("%s.%s", AuditLogEntriesTableName, AuditLogEntriesTableContextColumn),
		fmt.Sprintf("%s.%s", AuditLogEntriesTableName, CreatedOnColumn),
	}

	//
	// OAuth2 Clients Table.
	//

	// OAuth2ClientsTableColumns are the columns for the oauth2 clients table.
	OAuth2ClientsTableColumns = []string{
		fmt.Sprintf("%s.%s", OAuth2ClientsTableName, IDColumn),
		fmt.Sprintf("%s.%s", OAuth2ClientsTableName, OAuth2ClientsTableNameColumn),
		fmt.Sprintf("%s.%s", OAuth2ClientsTableName, OAuth2ClientsTableClientIDColumn),
		fmt.Sprintf("%s.%s", OAuth2ClientsTableName, OAuth2ClientsTableScopesColumn),
		fmt.Sprintf("%s.%s", OAuth2ClientsTableName, OAuth2ClientsTableRedirectURIColumn),
		fmt.Sprintf("%s.%s", OAuth2ClientsTableName, OAuth2ClientsTableClientSecretColumn),
		fmt.Sprintf("%s.%s", OAuth2ClientsTableName, CreatedOnColumn),
		fmt.Sprintf("%s.%s", OAuth2ClientsTableName, LastUpdatedOnColumn),
		fmt.Sprintf("%s.%s", OAuth2ClientsTableName, ArchivedOnColumn),
		fmt.Sprintf("%s.%s", OAuth2ClientsTableName, OAuth2ClientsTableOwnershipColumn),
	}

	//
	// Webhooks Table.
	//

	// WebhooksTableColumns are the columns for the webhooks table.
	WebhooksTableColumns = []string{
		fmt.Sprintf("%s.%s", WebhooksTableName, IDColumn),
		fmt.Sprintf("%s.%s", WebhooksTableName, WebhooksTableNameColumn),
		fmt.Sprintf("%s.%s", WebhooksTableName, WebhooksTableContentTypeColumn),
		fmt.Sprintf("%s.%s", WebhooksTableName, WebhooksTableURLColumn),
		fmt.Sprintf("%s.%s", WebhooksTableName, WebhooksTableMethodColumn),
		fmt.Sprintf("%s.%s", WebhooksTableName, WebhooksTableEventsColumn),
		fmt.Sprintf("%s.%s", WebhooksTableName, WebhooksTableDataTypesColumn),
		fmt.Sprintf("%s.%s", WebhooksTableName, WebhooksTableTopicsColumn),
		fmt.Sprintf("%s.%s", WebhooksTableName, CreatedOnColumn),
		fmt.Sprintf("%s.%s", WebhooksTableName, LastUpdatedOnColumn),
		fmt.Sprintf("%s.%s", WebhooksTableName, ArchivedOnColumn),
		fmt.Sprintf("%s.%s", WebhooksTableName, WebhooksTableOwnershipColumn),
	}

	//
	// Items Table.
	//

	// ItemsTableColumns are the columns for the items table.
	ItemsTableColumns = []string{
		fmt.Sprintf("%s.%s", ItemsTableName, IDColumn),
		fmt.Sprintf("%s.%s", ItemsTableName, ItemsTableNameColumn),
		fmt.Sprintf("%s.%s", ItemsTableName, ItemsTableDetailsColumn),
		fmt.Sprintf("%s.%s", ItemsTableName, CreatedOnColumn),
		fmt.Sprintf("%s.%s", ItemsTableName, LastUpdatedOnColumn),
		fmt.Sprintf("%s.%s", ItemsTableName, ArchivedOnColumn),
		fmt.Sprintf("%s.%s", ItemsTableName, ItemsTableUserOwnershipColumn),
	}
)
