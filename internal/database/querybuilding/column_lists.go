package querybuilding

import (
	"fmt"
)

var (
	//
	// Accounts Table.
	//

	// AccountsUserMembershipTableColumns are the columns for the items table.
	AccountsUserMembershipTableColumns = []string{
		fmt.Sprintf("%s.%s", AccountsUserMembershipTableName, IDColumn),
		fmt.Sprintf("%s.%s", AccountsUserMembershipTableName, userOwnershipColumn),
		fmt.Sprintf("%s.%s", AccountsUserMembershipTableName, accountOwnershipColumn),
		fmt.Sprintf("%s.%s", AccountsUserMembershipTableName, AccountsUserMembershipTableAccountRolesColumn),
		fmt.Sprintf("%s.%s", AccountsUserMembershipTableName, AccountsUserMembershipTableDefaultUserAccountColumn),
		fmt.Sprintf("%s.%s", AccountsUserMembershipTableName, CreatedOnColumn),
		fmt.Sprintf("%s.%s", AccountsUserMembershipTableName, LastUpdatedOnColumn),
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
		fmt.Sprintf("%s.%s", AccountsTableName, AccountsTableBillingStatusColumn),
		fmt.Sprintf("%s.%s", AccountsTableName, AccountsTableContactEmailColumn),
		fmt.Sprintf("%s.%s", AccountsTableName, AccountsTableContactPhoneColumn),
		fmt.Sprintf("%s.%s", AccountsTableName, AccountsTablePaymentProcessorCustomerIDColumn),
		fmt.Sprintf("%s.%s", AccountsTableName, AccountsTableSubscriptionPlanIDColumn),
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
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableRequiresPasswordChangeColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTablePasswordLastChangedOnColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableTwoFactorSekretColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableTwoFactorVerifiedOnColumn),
		fmt.Sprintf("%s.%s", UsersTableName, UsersTableServiceRolesColumn),
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
	// API Clients Table.
	//

	// APIClientsTableColumns are the columns for the API clients table.
	APIClientsTableColumns = []string{
		fmt.Sprintf("%s.%s", APIClientsTableName, IDColumn),
		fmt.Sprintf("%s.%s", APIClientsTableName, ExternalIDColumn),
		fmt.Sprintf("%s.%s", APIClientsTableName, APIClientsTableNameColumn),
		fmt.Sprintf("%s.%s", APIClientsTableName, APIClientsTableClientIDColumn),
		fmt.Sprintf("%s.%s", APIClientsTableName, APIClientsTableSecretKeyColumn),
		fmt.Sprintf("%s.%s", APIClientsTableName, CreatedOnColumn),
		fmt.Sprintf("%s.%s", APIClientsTableName, LastUpdatedOnColumn),
		fmt.Sprintf("%s.%s", APIClientsTableName, ArchivedOnColumn),
		fmt.Sprintf("%s.%s", APIClientsTableName, APIClientsTableOwnershipColumn),
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
		fmt.Sprintf("%s.%s", ItemsTableName, ItemsTableAccountOwnershipColumn),
	}
)
