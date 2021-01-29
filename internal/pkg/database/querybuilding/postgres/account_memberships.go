package postgres

import (
	"fmt"

	"github.com/Masterminds/squirrel"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
)

// BuildMarkAccountAsUserPrimaryQuery builds a query that marks a user's account as their primary.
func (q *Postgres) BuildMarkAccountAsUserPrimaryQuery(userID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(querybuilding.AccountsMembershipTableName).
		Set(querybuilding.AccountsMembershipTablePrimaryUserAccountColumn, squirrel.And{
			squirrel.Eq{querybuilding.AccountsMembershipTableUserOwnershipColumn: userID},
			squirrel.Eq{querybuilding.AccountsMembershipTableAccountOwnershipColumn: accountID},
		}).
		Where(squirrel.Eq{
			querybuilding.AccountsMembershipTableUserOwnershipColumn: userID,
		}).
		Where(squirrel.NotEq{
			querybuilding.ArchivedOnColumn: nil,
		}),
	)
}

// BuildUserIsMemberOfAccountQuery builds a query that checks to see if the user is the member of a given account.
func (q *Postgres) BuildUserIsMemberOfAccountQuery(userID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(fmt.Sprintf("%s.%s", querybuilding.AccountsMembershipTableName, querybuilding.IDColumn)).
		Prefix(querybuilding.ExistencePrefix).
		From(querybuilding.AccountsMembershipTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.AccountsMembershipTableName, querybuilding.AccountsMembershipTableUserOwnershipColumn): accountID,
			fmt.Sprintf("%s.%s", querybuilding.AccountsMembershipTableName, querybuilding.AccountsMembershipTableUserOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", querybuilding.AccountsMembershipTableName, querybuilding.ArchivedOnColumn):                           nil,
		}).
		Suffix(querybuilding.ExistenceSuffix),
	)
}

// BuildAddUserToAccountQuery builds a query that adds a user to an account.
func (q *Postgres) BuildAddUserToAccountQuery(userID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Insert(querybuilding.AccountsMembershipTableName).
		Columns(
			querybuilding.AccountsMembershipTableUserOwnershipColumn,
			querybuilding.AccountsMembershipTableAccountOwnershipColumn,
		).
		Values(
			userID,
			accountID,
		),
	)
}

// BuildRemoveUserFromAccountQuery builds a query that removes a user from an account.
func (q *Postgres) BuildRemoveUserFromAccountQuery(userID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Delete(querybuilding.AccountsMembershipTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", querybuilding.AccountsMembershipTableName, querybuilding.AccountsMembershipTableAccountOwnershipColumn): accountID,
			fmt.Sprintf("%s.%s", querybuilding.AccountsMembershipTableName, querybuilding.AccountsMembershipTableUserOwnershipColumn):    userID,
			fmt.Sprintf("%s.%s", querybuilding.AccountsMembershipTableName, querybuilding.ArchivedOnColumn):                              nil,
		}),
	)
}
