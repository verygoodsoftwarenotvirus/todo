package querybuilding

import (
	"context"
	"database/sql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

type (
	// AccountSQLQueryBuilder describes a structure capable of generating query/arg pairs for certain situations.
	AccountSQLQueryBuilder interface {
		BuildGetAccountQuery(ctx context.Context, accountID, userID uint64) (query string, args []interface{})
		BuildGetAllAccountsCountQuery(ctx context.Context) string
		BuildGetBatchOfAccountsQuery(ctx context.Context, beginID, endID uint64) (query string, args []interface{})
		BuildGetAccountsQuery(ctx context.Context, userID uint64, forAdmin bool, filter *types.QueryFilter) (query string, args []interface{})
		BuildAccountCreationQuery(ctx context.Context, input *types.AccountCreationInput) (query string, args []interface{})
		BuildUpdateAccountQuery(ctx context.Context, input *types.Account) (query string, args []interface{})
		BuildArchiveAccountQuery(ctx context.Context, accountID, userID uint64) (query string, args []interface{})
		BuildTransferAccountOwnershipQuery(ctx context.Context, currentOwnerID, newOwnerID, accountID uint64) (query string, args []interface{})
		BuildGetAuditLogEntriesForAccountQuery(ctx context.Context, accountID uint64) (query string, args []interface{})
	}

	// AccountUserMembershipSQLQueryBuilder describes a structure capable of generating query/arg pairs for certain situations.
	AccountUserMembershipSQLQueryBuilder interface {
		BuildGetDefaultAccountIDForUserQuery(ctx context.Context, userID uint64) (query string, args []interface{})
		BuildArchiveAccountMembershipsForUserQuery(ctx context.Context, userID uint64) (query string, args []interface{})
		BuildGetAccountMembershipsForUserQuery(ctx context.Context, userID uint64) (query string, args []interface{})
		BuildMarkAccountAsUserDefaultQuery(ctx context.Context, userID, accountID uint64) (query string, args []interface{})
		BuildModifyUserPermissionsQuery(ctx context.Context, userID, accountID uint64, newRoles []string) (query string, args []interface{})
		BuildTransferAccountMembershipsQuery(ctx context.Context, currentOwnerID, newOwnerID, accountID uint64) (query string, args []interface{})
		BuildUserIsMemberOfAccountQuery(ctx context.Context, userID, accountID uint64) (query string, args []interface{})
		BuildCreateMembershipForNewUserQuery(ctx context.Context, userID, accountID uint64) (query string, args []interface{})
		BuildAddUserToAccountQuery(ctx context.Context, input *types.AddUserToAccountInput) (query string, args []interface{})
		BuildRemoveUserFromAccountQuery(ctx context.Context, userID, accountID uint64) (query string, args []interface{})
	}

	// APIClientSQLQueryBuilder describes a structure capable of generating query/arg pairs for certain situations.
	APIClientSQLQueryBuilder interface {
		BuildGetBatchOfAPIClientsQuery(ctx context.Context, beginID, endID uint64) (query string, args []interface{})
		BuildGetAPIClientByClientIDQuery(ctx context.Context, clientID string) (query string, args []interface{})
		BuildGetAPIClientByDatabaseIDQuery(ctx context.Context, clientID, userID uint64) (query string, args []interface{})
		BuildGetAllAPIClientsCountQuery(ctx context.Context) string
		BuildGetAPIClientsQuery(ctx context.Context, userID uint64, filter *types.QueryFilter) (query string, args []interface{})
		BuildCreateAPIClientQuery(ctx context.Context, input *types.APIClientCreationInput) (query string, args []interface{})
		BuildUpdateAPIClientQuery(ctx context.Context, input *types.APIClient) (query string, args []interface{})
		BuildArchiveAPIClientQuery(ctx context.Context, clientID, userID uint64) (query string, args []interface{})
		BuildGetAuditLogEntriesForAPIClientQuery(ctx context.Context, clientID uint64) (query string, args []interface{})
	}

	// AuditLogEntrySQLQueryBuilder describes a structure capable of generating query/arg pairs for certain situations.
	AuditLogEntrySQLQueryBuilder interface {
		BuildGetAuditLogEntryQuery(ctx context.Context, entryID uint64) (query string, args []interface{})
		BuildGetAllAuditLogEntriesCountQuery(ctx context.Context) string
		BuildGetBatchOfAuditLogEntriesQuery(ctx context.Context, beginID, endID uint64) (query string, args []interface{})
		BuildGetAuditLogEntriesQuery(ctx context.Context, filter *types.QueryFilter) (query string, args []interface{})
		BuildCreateAuditLogEntryQuery(ctx context.Context, input *types.AuditLogEntryCreationInput) (query string, args []interface{})
	}

	// UserSQLQueryBuilder describes a structure capable of generating query/arg pairs for certain situations.
	UserSQLQueryBuilder interface {
		BuildUserHasStatusQuery(ctx context.Context, userID uint64, statuses ...string) (query string, args []interface{})
		BuildGetUserQuery(ctx context.Context, userID uint64) (query string, args []interface{})
		BuildGetUsersQuery(ctx context.Context, filter *types.QueryFilter) (query string, args []interface{})
		BuildGetUserWithUnverifiedTwoFactorSecretQuery(ctx context.Context, userID uint64) (query string, args []interface{})
		BuildGetUserByUsernameQuery(ctx context.Context, username string) (query string, args []interface{})
		BuildSearchForUserByUsernameQuery(ctx context.Context, usernameQuery string) (query string, args []interface{})
		BuildGetAllUsersCountQuery(ctx context.Context) (query string)
		BuildCreateUserQuery(ctx context.Context, input *types.UserDataStoreCreationInput) (query string, args []interface{})
		BuildUpdateUserQuery(ctx context.Context, input *types.User) (query string, args []interface{})
		BuildUpdateUserPasswordQuery(ctx context.Context, userID uint64, newHash string) (query string, args []interface{})
		BuildUpdateUserTwoFactorSecretQuery(ctx context.Context, userID uint64, newSecret string) (query string, args []interface{})
		BuildVerifyUserTwoFactorSecretQuery(ctx context.Context, userID uint64) (query string, args []interface{})
		BuildArchiveUserQuery(ctx context.Context, userID uint64) (query string, args []interface{})
		BuildGetAuditLogEntriesForUserQuery(ctx context.Context, userID uint64) (query string, args []interface{})
		BuildSetUserStatusQuery(ctx context.Context, input *types.UserReputationUpdateInput) (query string, args []interface{})
	}

	// WebhookSQLQueryBuilder describes a structure capable of generating query/arg pairs for certain situations.
	WebhookSQLQueryBuilder interface {
		BuildGetWebhookQuery(ctx context.Context, webhookID, accountID uint64) (query string, args []interface{})
		BuildGetAllWebhooksCountQuery(ctx context.Context) string
		BuildGetBatchOfWebhooksQuery(ctx context.Context, beginID, endID uint64) (query string, args []interface{})
		BuildGetWebhooksQuery(ctx context.Context, accountID uint64, filter *types.QueryFilter) (query string, args []interface{})
		BuildCreateWebhookQuery(ctx context.Context, x *types.WebhookCreationInput) (query string, args []interface{})
		BuildUpdateWebhookQuery(ctx context.Context, input *types.Webhook) (query string, args []interface{})
		BuildArchiveWebhookQuery(ctx context.Context, webhookID, accountID uint64) (query string, args []interface{})
		BuildGetAuditLogEntriesForWebhookQuery(ctx context.Context, webhookID uint64) (query string, args []interface{})
	}

	// ItemSQLQueryBuilder describes a structure capable of generating query/arg pairs for certain situations.
	ItemSQLQueryBuilder interface {
		BuildItemExistsQuery(ctx context.Context, itemID, accountID uint64) (query string, args []interface{})
		BuildGetItemQuery(ctx context.Context, itemID, accountID uint64) (query string, args []interface{})
		BuildGetAllItemsCountQuery(ctx context.Context) string
		BuildGetBatchOfItemsQuery(ctx context.Context, beginID, endID uint64) (query string, args []interface{})
		BuildGetItemsQuery(ctx context.Context, accountID uint64, forAdmin bool, filter *types.QueryFilter) (query string, args []interface{})
		BuildGetItemsWithIDsQuery(ctx context.Context, accountID uint64, limit uint8, ids []uint64, forAdmin bool) (query string, args []interface{})
		BuildCreateItemQuery(ctx context.Context, input *types.ItemCreationInput) (query string, args []interface{})
		BuildUpdateItemQuery(ctx context.Context, input *types.Item) (query string, args []interface{})
		BuildArchiveItemQuery(ctx context.Context, itemID, accountID uint64) (query string, args []interface{})
		BuildGetAuditLogEntriesForItemQuery(ctx context.Context, itemID uint64) (query string, args []interface{})
	}

	// SQLQueryBuilder describes anything that builds SQL queries to manage our data.
	SQLQueryBuilder interface {
		BuildMigrationFunc(db *sql.DB) func()
		BuildTestUserCreationQuery(ctx context.Context, testUserConfig *types.TestUserCreationConfig) (query string, args []interface{})

		AccountSQLQueryBuilder
		AccountUserMembershipSQLQueryBuilder
		UserSQLQueryBuilder
		AuditLogEntrySQLQueryBuilder
		APIClientSQLQueryBuilder
		WebhookSQLQueryBuilder
		ItemSQLQueryBuilder
	}
)
