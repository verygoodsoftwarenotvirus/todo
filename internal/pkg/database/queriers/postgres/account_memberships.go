package postgres

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
)

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

// MarkAccountAsUserPrimary marks an account as the primary account.
func (q *Postgres) MarkAccountAsUserPrimary(ctx context.Context, userID, accountID uint64) error {
	query, args := q.BuildMarkAccountAsUserPrimaryQuery(userID, accountID)

	// create the user/account association.
	if _, err := q.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("executing account association creation query: %w", err)
	}

	return nil
}

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

// AddUserToAccount adds a user to a given account.
func (q *Postgres) AddUserToAccount(ctx context.Context, userID, accountID uint64) error {
	query, args := q.BuildAddUserToAccountQuery(userID, accountID)

	// create the user/account association.
	if _, err := q.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("executing account association creation query: %w", err)
	}

	return nil
}

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

// RemoveUserFromAccount removes a user from a given account.
func (q *Postgres) RemoveUserFromAccount(ctx context.Context, userID, accountID uint64) error {
	query, args := q.BuildRemoveUserFromAccountQuery(userID, accountID)

	// remove the user/account association.
	if _, err := q.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("executing account association creation query: %w", err)
	}

	return nil
}
