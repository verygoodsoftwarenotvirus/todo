package sqlite

import (
	"fmt"
	"math"

	"github.com/Masterminds/squirrel"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.AccountUserMembershipSQLQueryBuilder = (*BaseQueryBuilder)(nil)

// BuildArchiveAccountMembershipsForUserQuery does .
func (q *BaseQueryBuilder) BuildArchiveAccountMembershipsForUserQuery(userID uint64) (query string, args []interface{}) {
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
func (q *BaseQueryBuilder) BuildGetAccountMembershipsForUserQuery(userID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(querybuilding.AccountsUserMembershipTableColumns...).
		From(querybuilding.AccountsUserMembershipTableName).
		Where(squirrel.Eq{
			querybuilding.ArchivedOnColumn:                               nil,
			querybuilding.AccountsUserMembershipTableUserOwnershipColumn: userID,
		}),
	)
}

// BuildMarkAccountAsUserDefaultQuery does .
func (q *BaseQueryBuilder) BuildMarkAccountAsUserDefaultQuery(userID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.AccountsUserMembershipTableName).
		Set(querybuilding.AccountsUserMembershipTableDefaultUserAccountColumn, squirrel.And{
			squirrel.Eq{querybuilding.AccountsUserMembershipTableUserOwnershipColumn: userID},
			squirrel.Eq{querybuilding.AccountsUserMembershipTableAccountOwnershipColumn: accountID},
		}).
		Where(squirrel.Eq{
			querybuilding.AccountsUserMembershipTableUserOwnershipColumn: userID,
			querybuilding.ArchivedOnColumn:                               nil,
		}),
	)
}

// BuildTransferAccountOwnershipQuery does .
func (q *BaseQueryBuilder) BuildTransferAccountOwnershipQuery(oldOwnerID, newOwnerID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.AccountsUserMembershipTableName).
		Set(querybuilding.AccountsUserMembershipTableUserOwnershipColumn, newOwnerID).
		Where(squirrel.Eq{
			querybuilding.ArchivedOnColumn:                                  nil,
			querybuilding.AccountsUserMembershipTableUserOwnershipColumn:    oldOwnerID,
			querybuilding.AccountsUserMembershipTableAccountOwnershipColumn: accountID,
		}),
	)
}

// BuildModifyUserPermissionsQuery builds.
func (q *BaseQueryBuilder) BuildModifyUserPermissionsQuery(userID, accountID uint64, perms permissions.ServiceUserPermissions) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.AccountsUserMembershipTableName).
		Set(querybuilding.AccountsUserMembershipTableUserPermissionsColumn, perms).
		Where(squirrel.Eq{
			querybuilding.AccountsUserMembershipTableUserOwnershipColumn:    userID,
			querybuilding.AccountsUserMembershipTableAccountOwnershipColumn: accountID,
		}),
	)
}

// BuildCreateMembershipForNewUserQuery builds a query that .
func (q *BaseQueryBuilder) BuildCreateMembershipForNewUserQuery(userID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Insert(querybuilding.AccountsUserMembershipTableName).
		Columns(
			querybuilding.AccountsUserMembershipTableUserOwnershipColumn,
			querybuilding.AccountsUserMembershipTableAccountOwnershipColumn,
			querybuilding.AccountsUserMembershipTableDefaultUserAccountColumn,
			querybuilding.AccountsUserMembershipTableUserPermissionsColumn,
		).
		Values(
			userID,
			accountID,
			true,
			math.MaxUint32,
		),
	)
}

// BuildMarkAccountAsUserPrimaryQuery builds a query that marks a user's account as their primary.
func (q *BaseQueryBuilder) BuildMarkAccountAsUserPrimaryQuery(userID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.AccountsUserMembershipTableName).
		Set(querybuilding.AccountsUserMembershipTableDefaultUserAccountColumn, squirrel.And{
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
func (q *BaseQueryBuilder) BuildUserIsMemberOfAccountQuery(userID, accountID uint64) (query string, args []interface{}) {
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
func (q *BaseQueryBuilder) BuildAddUserToAccountQuery(accountID uint64, input *types.AddUserToAccountInput) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Insert(querybuilding.AccountsUserMembershipTableName).
		Columns(
			querybuilding.AccountsUserMembershipTableUserOwnershipColumn,
			querybuilding.AccountsUserMembershipTableAccountOwnershipColumn,
			querybuilding.AccountsUserMembershipTableUserPermissionsColumn,
		).
		Values(
			input.UserID,
			accountID,
			input.UserAccountPermissions,
		),
	)
}

// BuildRemoveUserFromAccountQuery builds a query that removes a user from an account.
func (q *BaseQueryBuilder) BuildRemoveUserFromAccountQuery(userID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Delete(querybuilding.AccountsUserMembershipTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.AccountsUserMembershipTableAccountOwnershipColumn): accountID,
			fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.AccountsUserMembershipTableUserOwnershipColumn):    userID,
			fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.ArchivedOnColumn):                                  nil,
		}),
	)
}
