package mariadb

import (
	"fmt"

	"github.com/Masterminds/squirrel"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions/bitmask"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.AccountUserMembershipSQLQueryBuilder = (*MariaDB)(nil)

// BuildMarkAccountAsUserDefaultQuery does .
func (q *MariaDB) BuildMarkAccountAsUserDefaultQuery(userID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.AccountsUserMembershipTableName).
		Where(squirrel.Eq{
			"FILL ME OUT PLEASE": true,
			"oldOwnerID":         userID,
			"accountID":          accountID,
		}),
	)
}

// BuildTransferAccountOwnershipQuery does .
func (q *MariaDB) BuildTransferAccountOwnershipQuery(oldOwnerID, newOwnerID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.AccountsUserMembershipTableName).
		Where(squirrel.Eq{
			"FILL ME OUT PLEASE": true,
			"oldOwnerID":         oldOwnerID,
			"newOwnerID":         newOwnerID,
			"accountID":          accountID,
		}),
	)
}

// BuildModifyUserPermissionsQuery builds.
func (q *MariaDB) BuildModifyUserPermissionsQuery(userID, accountID uint64, permissions bitmask.ServiceUserPermissions) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.AccountsUserMembershipTableName).
		Where(squirrel.Eq{
			"FILL ME OUT PLEASE": true,
			"userID":             userID,
			"permissions":        permissions,
			"accountID":          accountID,
		}),
	)
}

// BuildArchiveAccountMembershipsForUserQuery does .
func (q *MariaDB) BuildArchiveAccountMembershipsForUserQuery(userID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.AccountsUserMembershipTableName).
		Set(querybuilding.ArchivedOnColumn, currentUnixTimeQuery).
		Where(squirrel.Eq{
			querybuilding.AccountsUserMembershipTableUserOwnershipColumn: userID,
			querybuilding.ArchivedOnColumn:                               nil,
		}),
	)
}

// BuildGetAccountMembershipsForUserQuery does .
func (q *MariaDB) BuildGetAccountMembershipsForUserQuery(userID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.AccountsUserMembershipTableColumns...).
		From(querybuilding.AccountsUserMembershipTableName).
		Where(squirrel.Eq{
			querybuilding.ArchivedOnColumn:                               nil,
			querybuilding.AccountsUserMembershipTableUserOwnershipColumn: userID,
		}),
	)
}

// BuildCreateMembershipForNewUserQuery builds a query that .
func (q *MariaDB) BuildCreateMembershipForNewUserQuery(userID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Insert(querybuilding.AccountsUserMembershipTableName).
		Columns(
			querybuilding.AccountsUserMembershipTableUserOwnershipColumn,
			querybuilding.AccountsUserMembershipTableAccountOwnershipColumn,
			querybuilding.AccountsUserMembershipTablePrimaryUserAccountColumn,
		).
		Values(
			userID,
			accountID,
			true,
		),
	)
}

// BuildGetAccountUserMembershipQuery does .
func (q *MariaDB) BuildGetAccountUserMembershipQuery(accountUserMembershipID, userID uint64) (query string, args []interface{}) {
	panic("implement me")
}

// BuildGetAuditLogEntriesForAccountUserMembershipQuery does .
func (q *MariaDB) BuildGetAuditLogEntriesForAccountUserMembershipQuery(accountUserMembershipID uint64) (query string, args []interface{}) {
	panic("implement me")
}

// BuildMarkAccountAsUserPrimaryQuery builds a query that marks a user's account as their primary.
func (q *MariaDB) BuildMarkAccountAsUserPrimaryQuery(userID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.AccountsUserMembershipTableName).
		Set(querybuilding.AccountsUserMembershipTablePrimaryUserAccountColumn, squirrel.And{
			squirrel.Eq{querybuilding.AccountsUserMembershipTableUserOwnershipColumn: userID},
			squirrel.Eq{querybuilding.AccountsUserMembershipTableAccountOwnershipColumn: accountID},
		}).
		Where(squirrel.Eq{
			querybuilding.AccountsUserMembershipTableUserOwnershipColumn: userID,
			querybuilding.ArchivedOnColumn:                               nil,
		}),
	)
}

// BuildUserIsMemberOfAccountQuery builds a query that checks to see if the user is the member of a given account.
func (q *MariaDB) BuildUserIsMemberOfAccountQuery(userID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.IDColumn)).
		Prefix(querybuilding.ExistencePrefix).
		From(querybuilding.AccountsUserMembershipTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.AccountsUserMembershipTableUserOwnershipColumn): accountID,
			fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.AccountsUserMembershipTableUserOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.ArchivedOnColumn):                               nil,
		}).
		Suffix(querybuilding.ExistenceSuffix),
	)
}

// BuildAddUserToAccountQuery builds a query that adds a user to an account.
func (q *MariaDB) BuildAddUserToAccountQuery(userID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Insert(querybuilding.AccountsUserMembershipTableName).
		Columns(
			querybuilding.AccountsUserMembershipTableUserOwnershipColumn,
			querybuilding.AccountsUserMembershipTableAccountOwnershipColumn,
		).
		Values(
			userID,
			accountID,
		),
	)
}

// BuildRemoveUserFromAccountQuery builds a query that removes a user from an account.
func (q *MariaDB) BuildRemoveUserFromAccountQuery(userID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Delete(querybuilding.AccountsUserMembershipTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.AccountsUserMembershipTableAccountOwnershipColumn): accountID,
			fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.AccountsUserMembershipTableUserOwnershipColumn):    userID,
			fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.ArchivedOnColumn):                                  nil,
		}),
	)
}
