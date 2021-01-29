package postgres

import (
	"fmt"

	"github.com/Masterminds/squirrel"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
)

// BuildMarkAccountAsUserPrimaryQuery builds a query that marks a user's account as their primary.
func (q *Postgres) BuildMarkAccountAsUserPrimaryQuery(userID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Update(queriers.AccountsMembershipTableName).
		Set(queriers.AccountsMembershipTablePrimaryUserAccountColumn, squirrel.And{
			squirrel.Eq{queriers.AccountsMembershipTableUserOwnershipColumn: userID},
			squirrel.Eq{queriers.AccountsMembershipTableAccountOwnershipColumn: accountID},
		}).
		Where(squirrel.Eq{
			queriers.AccountsMembershipTableUserOwnershipColumn: userID,
		}).
		Where(squirrel.NotEq{
			queriers.ArchivedOnColumn: nil,
		}),
	)
}

// BuildUserIsMemberOfAccountQuery builds a query that checks to see if the user is the member of a given account.
func (q *Postgres) BuildUserIsMemberOfAccountQuery(userID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Select(fmt.Sprintf("%s.%s", queriers.AccountsMembershipTableName, queriers.IDColumn)).
		Prefix(queriers.ExistencePrefix).
		From(queriers.AccountsMembershipTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.AccountsMembershipTableName, queriers.AccountsMembershipTableUserOwnershipColumn): accountID,
			fmt.Sprintf("%s.%s", queriers.AccountsMembershipTableName, queriers.AccountsMembershipTableUserOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", queriers.AccountsMembershipTableName, queriers.ArchivedOnColumn):                           nil,
		}).
		Suffix(queriers.ExistenceSuffix),
	)
}

// BuildAddUserToAccountQuery builds a query that adds a user to an account.
func (q *Postgres) BuildAddUserToAccountQuery(userID, accountID uint64) (query string, args []interface{}) {
	return q.buildQuery(q.sqlBuilder.
		Insert(queriers.AccountsMembershipTableName).
		Columns(
			queriers.AccountsMembershipTableUserOwnershipColumn,
			queriers.AccountsMembershipTableAccountOwnershipColumn,
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
		Delete(queriers.AccountsMembershipTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.AccountsMembershipTableName, queriers.AccountsMembershipTableAccountOwnershipColumn): accountID,
			fmt.Sprintf("%s.%s", queriers.AccountsMembershipTableName, queriers.AccountsMembershipTableUserOwnershipColumn):    userID,
			fmt.Sprintf("%s.%s", queriers.AccountsMembershipTableName, queriers.ArchivedOnColumn):                              nil,
		}),
	)
}
