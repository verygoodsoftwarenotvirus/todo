package sqlite

import (
	"fmt"

	"github.com/Masterminds/squirrel"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.AccountUserMembershipSQLQueryBuilder = (*Sqlite)(nil)

// BuildGetAccountUserMembershipQuery does .
func (q *Sqlite) BuildGetAccountUserMembershipQuery(accountUserMembershipID, userID uint64) (query string, args []interface{}) {
	panic("implement me")
}

// BuildGetAllAccountUserMembershipsCountQuery does .
func (q *Sqlite) BuildGetAllAccountUserMembershipsCountQuery() string {
	panic("implement me")
}

// BuildGetBatchOfAccountUserMembershipsQuery does .
func (q *Sqlite) BuildGetBatchOfAccountUserMembershipsQuery(beginID, endID uint64) (query string, args []interface{}) {
	panic("implement me")
}

// BuildGetAccountUserMembershipsQuery does .
func (q *Sqlite) BuildGetAccountUserMembershipsQuery(userID uint64, forAdmin bool, filter *types.QueryFilter) (query string, args []interface{}) {
	panic("implement me")
}

// BuildCreateAccountUserMembershipQuery does .
func (q *Sqlite) BuildCreateAccountUserMembershipQuery(input *types.AccountUserMembershipCreationInput) (query string, args []interface{}) {
	panic("implement me")
}

// BuildArchiveAccountUserMembershipQuery does .
func (q *Sqlite) BuildArchiveAccountUserMembershipQuery(accountUserMembershipID, userID uint64) (query string, args []interface{}) {
	panic("implement me")
}

// BuildGetAuditLogEntriesForAccountUserMembershipQuery does .
func (q *Sqlite) BuildGetAuditLogEntriesForAccountUserMembershipQuery(accountUserMembershipID uint64) (query string, args []interface{}) {
	panic("implement me")
}

// BuildMarkAccountAsUserPrimaryQuery builds a query that marks a user's account as their primary.
func (q *Sqlite) BuildMarkAccountAsUserPrimaryQuery(userID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.AccountsUserMembershipTableName).
		Set(querybuilding.AccountsUserMembershipTablePrimaryUserAccountColumn, squirrel.And{
			squirrel.Eq{querybuilding.AccountsUserMembershipTableUserOwnershipColumn: userID},
			squirrel.Eq{querybuilding.AccountsUserMembershipTableAccountOwnershipColumn: accountID},
		}).
		Where(squirrel.Eq{
			querybuilding.AccountsUserMembershipTableUserOwnershipColumn: userID,
		}).
		Where(squirrel.NotEq{
			querybuilding.ArchivedOnColumn: nil,
		}),
	)
}

// BuildUserIsMemberOfAccountQuery builds a query that checks to see if the user is the member of a given account.
func (q *Sqlite) BuildUserIsMemberOfAccountQuery(userID, accountID uint64) (query string, args []interface{}) {
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
func (q *Sqlite) BuildAddUserToAccountQuery(userID, accountID uint64) (query string, args []interface{}) {
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
func (q *Sqlite) BuildRemoveUserFromAccountQuery(userID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Delete(querybuilding.AccountsUserMembershipTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.AccountsUserMembershipTableAccountOwnershipColumn): accountID,
			fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.AccountsUserMembershipTableUserOwnershipColumn):    userID,
			fmt.Sprintf("%s.%s", querybuilding.AccountsUserMembershipTableName, querybuilding.ArchivedOnColumn):                                  nil,
		}),
	)
}
