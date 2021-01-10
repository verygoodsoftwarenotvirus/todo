package postgres

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
)

/*
func (q *Postgres) scanAccountMembership(scan database.Scanner) (*types.AccountMembership, error) {
	var (
		x = &types.AccountMembership{}
	)

	targetVars := []interface{}{
		&x.ID,
		&x.BelongsToUser,
		&x.BelongsToAccount,
		&x.CreatedOn,
		&x.ArchivedOn,
	}

	if err := scan.Scan(targetVars...); err != nil {
		return nil, err
	}

	return x, nil
}

// scanAccountMemberships takes some database rows and turns them into a slice of accounts.
func (q *Postgres) scanAccountMemberships(rows database.ResultIterator, includeCount bool) ([]types.Account, uint64, error) {
	var (
		list  []types.Account
		count uint64
	)

	for rows.Next() {
		x, c, err := q.scanAccount(rows, includeCount)
		if err != nil {
			return nil, 0, err
		}

		if count == 0 && includeCount {
			count = c
		}

		list = append(list, *x)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if closeErr := rows.Close(); closeErr != nil {
		q.logger.Error(closeErr, "closing database rows")
	}

	return list, count, nil
}
*/

func (q *Postgres) buildMarkAccountAsUserPrimaryQuery(userID, accountID uint64) (query string, args []interface{}) {
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
	query, args := q.buildMarkAccountAsUserPrimaryQuery(userID, accountID)

	// create the user/account association.
	if _, err := q.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("error executing account association creation query: %w", err)
	}

	return nil
}

func (q *Postgres) buildUserIsMemberOfAccountQuery(userID, accountID uint64) (query string, args []interface{}) {
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

func (q *Postgres) buildAddUserToAccountQuery(userID, accountID uint64) (query string, args []interface{}) {
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
	query, args := q.buildAddUserToAccountQuery(userID, accountID)

	// create the user/account association.
	if _, err := q.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("error executing account association creation query: %w", err)
	}

	return nil
}

func (q *Postgres) buildRemoveUserFromAccountQuery(userID, accountID uint64) (query string, args []interface{}) {
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
	query, args := q.buildRemoveUserFromAccountQuery(userID, accountID)

	// remove the user/account association.
	if _, err := q.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("error executing account association creation query: %w", err)
	}

	return nil
}
