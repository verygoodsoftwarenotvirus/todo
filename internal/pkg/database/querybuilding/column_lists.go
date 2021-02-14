package querybuilding

import (
	"fmt"
)

var (
	//
	// AccountSubscriptionPlans Table.
	//

	// AccountSubscriptionPlansTableColumns are the columns for the users table.
	AccountSubscriptionPlansTableColumns = []string{
		fmt.Sprintf("%s.%s", AccountSubscriptionPlansTableName, IDColumn),
		fmt.Sprintf("%s.%s", AccountSubscriptionPlansTableName, ExternalIDColumn),
		fmt.Sprintf("%s.%s", AccountSubscriptionPlansTableName, AccountSubscriptionPlansTableNameColumn),
		fmt.Sprintf("%s.%s", AccountSubscriptionPlansTableName, AccountSubscriptionPlansTableDescriptionColumn),
		fmt.Sprintf("%s.%s", AccountSubscriptionPlansTableName, AccountSubscriptionPlansTablePriceColumn),
		fmt.Sprintf("%s.%s", AccountSubscriptionPlansTableName, AccountSubscriptionPlansTablePeriodColumn),
		fmt.Sprintf("%s.%s", AccountSubscriptionPlansTableName, CreatedOnColumn),
		fmt.Sprintf("%s.%s", AccountSubscriptionPlansTableName, LastUpdatedOnColumn),
		fmt.Sprintf("%s.%s", AccountSubscriptionPlansTableName, ArchivedOnColumn),
	}

	//
	// Accounts Table.
	//

	// AccountsUserMembershipTableColumns are the columns for the items table.
	AccountsUserMembershipTableColumns = []string{
		fmt.Sprintf("%s.%s", AccountsUserMembershipTableName, IDColumn),
		fmt.Sprintf("%s.%s", AccountsUserMembershipTableName, userOwnershipColumn),
		fmt.Sprintf("%s.%s", AccountsUserMembershipTableName, accountOwnershipColumn),
		fmt.Sprintf("%s.%s", AccountsUserMembershipTableName, AccountsUserMembershipTableUserPermissionsColumn),
		fmt.Sprintf("%s.%s", AccountsUserMembershipTableName, AccountsUserMembershipTablePrimaryUserAccountColumn),
		fmt.Sprintf("%s.%s", AccountsUserMembershipTableName, CreatedOnColumn),
		fmt.Sprintf("%s.%s", AccountsUserMembershipTableName, ArchivedOnColumn),
	}

	//
	// Accounts Table.
	//

	// AccountsTableColumns are the columns for the items table.
	AccountsTableColumns = []string{
		fmt.Sprintf("%s.%s", AccountsTableName, IDColumn),
		fmt.Sprintf("%s.%s", AccountsTableName, ExternalIDColumn),
		fmt.Sprintf("%s.%s", AccountsTableName, AccountsTableNameColumn),
		fmt.Sprintf("%s.%s", AccountsTableName, AccountsTablePlanIDColumn),
		fmt.Sprintf("%s.%s", AccountsTableName, AccountsTablePersonalAccountColumn),
		fmt.Sprintf("%s.%s", AccountsTableName, CreatedOnColumn),
		fmt.Sprintf("%s.%s", AccountsTableName, LastUpdatedOnColumn),
		fmt.Sprintf("%s.%s", AccountsTableName, ArchivedOnColumn),
		fmt.Sprintf("%s.%s", AccountsTableName, AccountsTableUserOwnershipColumn),
	}

	//
	// Users Table.
	//

	// UsersTableColumns are the columns for the users table.
	UsersTableColumns = []string{
		fmt.Sprintf("%s.%s", UsersTableName, IDColumn),
		fmt.Sprintf("%s.%s", UsersTableName, ExternalIDColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableUsernameColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableAvatarColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableHashedPasswordColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableSaltColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableRequiresPasswordChangeColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTablePasswordLastChangedOnColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableTwoFactorSekretColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableTwoFactorVerifiedOnColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableAdminPermissionsColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableReputationColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableStatusExplanationColumn),
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
		fmt.Sprintf("%s.%s", AuditLogEntriesTableName, ExternalIDColumn),
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
		fmt.Sprintf("%s.%s", OAuth2ClientsTableName, ExternalIDColumn),
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
	// Delegated Clients Table.
	//

	// DelegatedClientsTableColumns are the columns for the oauth2 clients table.
	DelegatedClientsTableColumns = []string{
		fmt.Sprintf("%s.%s", DelegatedClientsTableName, IDColumn),
		fmt.Sprintf("%s.%s", DelegatedClientsTableName, ExternalIDColumn),
		fmt.Sprintf("%s.%s", DelegatedClientsTableName, DelegatedClientsTableNameColumn),
		fmt.Sprintf("%s.%s", DelegatedClientsTableName, DelegatedClientsTableClientIDColumn),
		fmt.Sprintf("%s.%s", DelegatedClientsTableName, DelegatedClientsTableClientSecretColumn),
		fmt.Sprintf("%s.%s", DelegatedClientsTableName, CreatedOnColumn),
		fmt.Sprintf("%s.%s", DelegatedClientsTableName, LastUpdatedOnColumn),
		fmt.Sprintf("%s.%s", DelegatedClientsTableName, ArchivedOnColumn),
		fmt.Sprintf("%s.%s", DelegatedClientsTableName, OAuth2ClientsTableOwnershipColumn),
	}

	//
	// Webhooks Table.
	//

	// WebhooksTableColumns are the columns for the webhooks table.
	WebhooksTableColumns = []string{
		fmt.Sprintf("%s.%s", WebhooksTableName, IDColumn),
		fmt.Sprintf("%s.%s", WebhooksTableName, ExternalIDColumn),
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
		fmt.Sprintf("%s.%s", ItemsTableName, ExternalIDColumn),
		fmt.Sprintf("%s.%s", ItemsTableName, ItemsTableNameColumn),
		fmt.Sprintf("%s.%s", ItemsTableName, ItemsTableDetailsColumn),
		fmt.Sprintf("%s.%s", ItemsTableName, CreatedOnColumn),
		fmt.Sprintf("%s.%s", ItemsTableName, LastUpdatedOnColumn),
		fmt.Sprintf("%s.%s", ItemsTableName, ArchivedOnColumn),
		fmt.Sprintf("%s.%s", ItemsTableName, ItemsTableUserOwnershipColumn),
		fmt.Sprintf("%s.%s", ItemsTableName, ItemsTableAccountOwnershipColumn),
	}
)
