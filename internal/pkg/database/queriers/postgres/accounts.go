package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

var _ types.AccountDataManager = (*Postgres)(nil)

// scanAccount takes a database Scanner (i.e. *sql.Row) and scans the result into an Account struct.
func (q *Postgres) scanAccount(scan database.Scanner, includeCount bool) (*types.Account, uint64, error) {
	var (
		x     = &types.Account{}
		count uint64
	)

	targetVars := []interface{}{
		&x.ID,
		&x.Name,
		&x.PlanID,
		&x.PersonalAccount,
		&x.CreatedOn,
		&x.LastUpdatedOn,
		&x.ArchivedOn,
		&x.BelongsToUser,
	}

	if includeCount {
		targetVars = append(targetVars, &count)
	}

	if err := scan.Scan(targetVars...); err != nil {
		return nil, 0, err
	}

	return x, count, nil
}

// scanAccounts takes some database rows and turns them into a slice of accounts.
func (q *Postgres) scanAccounts(rows database.ResultIterator, includeCount bool) ([]types.Account, uint64, error) {
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

// buildAccountExistsQuery constructs a SQL query for checking if an account with a given ID belong to a user with a given ID exists.
func (q *Postgres) buildAccountExistsQuery(accountID, userID uint64) (query string, args []interface{}) {
	query, args, err := q.sqlBuilder.
		Select(fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.IDColumn)).
		Prefix(queriers.ExistencePrefix).
		From(queriers.AccountsTableName).
		Suffix(queriers.ExistenceSuffix).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.IDColumn):                         accountID,
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.AccountsTableUserOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.ArchivedOnColumn):                 nil,
		}).ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// AccountExists queries the database to see if a given account belonging to a given user exists.
func (q *Postgres) AccountExists(ctx context.Context, accountID, userID uint64) (exists bool, err error) {
	query, args := q.buildAccountExistsQuery(accountID, userID)

	err = q.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	return exists, err
}

// buildGetAccountQuery constructs a SQL query for fetching an account with a given ID belong to a user with a given ID.
func (q *Postgres) buildGetAccountQuery(accountID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Select(queriers.AccountsTableColumns...).
		From(queriers.AccountsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.IDColumn):                         accountID,
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.AccountsTableUserOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.ArchivedOnColumn):                 nil,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// GetAccount fetches an account from the database.
func (q *Postgres) GetAccount(ctx context.Context, accountID, userID uint64) (*types.Account, error) {
	query, args := q.buildGetAccountQuery(accountID, userID)
	row := q.db.QueryRowContext(ctx, query, args...)

	account, _, err := q.scanAccount(row, false)

	return account, err
}

// buildGetAllAccountsCountQuery returns a query that fetches the total number of accounts in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (q *Postgres) buildGetAllAccountsCountQuery() string {
	allAccountsCountQuery, _, err := q.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, queriers.AccountsTableName)).
		From(queriers.AccountsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.ArchivedOnColumn): nil,
		}).
		ToSql()
	q.logQueryBuildingError(err)

	return allAccountsCountQuery
}

// GetAllAccountsCount will fetch the count of accounts from the database.
func (q *Postgres) GetAllAccountsCount(ctx context.Context) (count uint64, err error) {
	err = q.db.QueryRowContext(ctx, q.buildGetAllAccountsCountQuery()).Scan(&count)
	return count, err
}

// buildGetBatchOfAccountsQuery returns a query that fetches every account in the database within a bucketed range.
func (q *Postgres) buildGetBatchOfAccountsQuery(beginID, endID uint64) (query string, args []interface{}) {
	query, args, err := q.sqlBuilder.
		Select(queriers.AccountsTableColumns...).
		From(queriers.AccountsTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.IDColumn): endID,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// GetAllAccounts fetches every account from the database and writes them to a channel. This method primarily exists
// to aid in administrative data tasks.
func (q *Postgres) GetAllAccounts(ctx context.Context, resultChannel chan []types.Account) error {
	count, err := q.GetAllAccountsCount(ctx)
	if err != nil {
		return fmt.Errorf("error fetching count of accounts: %w", err)
	}

	for beginID := uint64(1); beginID <= count; beginID += defaultBucketSize {
		endID := beginID + defaultBucketSize
		go func(begin, end uint64) {
			query, args := q.buildGetBatchOfAccountsQuery(begin, end)
			logger := q.logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, err := q.db.Query(query, args...)
			if errors.Is(err, sql.ErrNoRows) {
				return
			} else if err != nil {
				logger.Error(err, "querying for database rows")
				return
			}

			accounts, _, err := q.scanAccounts(rows, false)
			if err != nil {
				logger.Error(err, "scanning database rows")
				return
			}

			resultChannel <- accounts
		}(beginID, endID)
	}

	return nil
}

// buildGetAccountsQuery builds a SQL query selecting accounts that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (q *Postgres) buildGetAccountsQuery(userID uint64, forAdmin bool, filter *types.QueryFilter) (query string, args []interface{}) {
	where := squirrel.Eq{
		fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.ArchivedOnColumn):                 nil,
		fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.AccountsTableUserOwnershipColumn): userID,
	}

	countQueryBuilder := q.sqlBuilder.
		PlaceholderFormat(squirrel.Question).
		Select(allCountQuery).
		From(queriers.AccountsTableName)

	if !forAdmin {
		countQueryBuilder = countQueryBuilder.Where(where)
	}

	countQuery, countQueryArgs := q.buildQuery(countQueryBuilder)
	builder := q.sqlBuilder.
		Select(append(queriers.AccountsTableColumns, fmt.Sprintf("(%s)", countQuery))...).
		From(queriers.AccountsTableName)

	if !forAdmin {
		builder = builder.Where(where)
	}

	builder = builder.OrderBy(fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.CreatedOnColumn))

	if filter != nil {
		builder = queriers.ApplyFilterToQueryBuilder(filter, builder, queriers.AccountsTableName)
	}

	query, selectArgs := q.buildQuery(builder)

	return query, append(countQueryArgs, selectArgs...)
}

// GetAccounts fetches a list of accounts from the database that meet a particular filter.
func (q *Postgres) GetAccounts(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.AccountList, error) {
	query, args := q.buildGetAccountsQuery(userID, false, filter)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for accounts: %w", err)
	}

	accounts, count, err := q.scanAccounts(rows, true)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	list := &types.AccountList{
		Pagination: types.Pagination{
			Page:          filter.Page,
			Limit:         filter.Limit,
			FilteredCount: count,
			TotalCount:    count,
		},
		Accounts: accounts,
	}

	return list, nil
}

// GetAccountsForAdmin fetches a list of accounts from the database that meet a particular filter for all users.
func (q *Postgres) GetAccountsForAdmin(ctx context.Context, filter *types.QueryFilter) (*types.AccountList, error) {
	query, args := q.buildGetAccountsQuery(0, true, filter)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for accounts: %w", err)
	}

	accounts, count, err := q.scanAccounts(rows, true)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	list := &types.AccountList{
		Pagination: types.Pagination{
			Page:          filter.Page,
			Limit:         filter.Limit,
			FilteredCount: count,
			TotalCount:    count,
		},
		Accounts: accounts,
	}

	return list, nil
}

// buildCreateAccountQuery takes an account and returns a creation query for that account and the relevant arguments.
func (q *Postgres) buildCreateAccountQuery(input *types.Account) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Insert(queriers.AccountsTableName).
		Columns(
			queriers.AccountsTableNameColumn,
			queriers.AccountsTableUserOwnershipColumn,
		).
		Values(
			input.Name,
			input.BelongsToUser,
		).
		Suffix(fmt.Sprintf("RETURNING %s, %s", queriers.IDColumn, queriers.CreatedOnColumn)).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// CreateAccount creates an account in the database.
func (q *Postgres) CreateAccount(ctx context.Context, input *types.AccountCreationInput) (*types.Account, error) {
	x := &types.Account{
		Name:          input.Name,
		BelongsToUser: input.BelongsToUser,
	}

	query, args := q.buildCreateAccountQuery(x)

	// create the account.
	err := q.db.QueryRowContext(ctx, query, args...).Scan(&x.ID, &x.CreatedOn)
	if err != nil {
		return nil, fmt.Errorf("error executing account creation query: %w", err)
	}

	return x, nil
}

// buildUpdateAccountQuery takes an account and returns an update SQL query, with the relevant query parameters.
func (q *Postgres) buildUpdateAccountQuery(input *types.Account) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(queriers.AccountsTableName).
		Set(queriers.AccountsTableNameColumn, input.Name).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                         input.ID,
			queriers.AccountsTableUserOwnershipColumn: input.BelongsToUser,
		}).
		Suffix(fmt.Sprintf("RETURNING %s", queriers.LastUpdatedOnColumn)).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// UpdateAccount updates a particular account. Note that UpdateAccount expects the provided input to have a valid ID.
func (q *Postgres) UpdateAccount(ctx context.Context, input *types.Account) error {
	query, args := q.buildUpdateAccountQuery(input)
	return q.db.QueryRowContext(ctx, query, args...).Scan(&input.LastUpdatedOn)
}

// buildArchiveAccountQuery returns a SQL query which marks a given account belonging to a given user as archived.
func (q *Postgres) buildArchiveAccountQuery(accountID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(queriers.AccountsTableName).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                         accountID,
			queriers.ArchivedOnColumn:                 nil,
			queriers.AccountsTableUserOwnershipColumn: userID,
		}).
		Suffix(fmt.Sprintf("RETURNING %s", queriers.ArchivedOnColumn)).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// ArchiveAccount marks an account as archived in the database.
func (q *Postgres) ArchiveAccount(ctx context.Context, accountID, userID uint64) error {
	query, args := q.buildArchiveAccountQuery(accountID, userID)

	res, err := q.db.ExecContext(ctx, query, args...)
	if res != nil {
		if rowCount, rowCountErr := res.RowsAffected(); rowCountErr == nil && rowCount == 0 {
			return sql.ErrNoRows
		}
	}

	return err
}

// LogAccountCreationEvent saves a AccountCreationEvent in the audit log table.
func (q *Postgres) LogAccountCreationEvent(ctx context.Context, account *types.Account) {
	q.createAuditLogEntry(ctx, audit.BuildAccountCreationEventEntry(account))
}

// LogAccountUpdateEvent saves a AccountUpdateEvent in the audit log table.
func (q *Postgres) LogAccountUpdateEvent(ctx context.Context, userID, accountID uint64, changes []types.FieldChangeSummary) {
	q.createAuditLogEntry(ctx, audit.BuildAccountUpdateEventEntry(userID, accountID, changes))
}

// LogAccountArchiveEvent saves a AccountArchiveEvent in the audit log table.
func (q *Postgres) LogAccountArchiveEvent(ctx context.Context, userID, accountID uint64) {
	q.createAuditLogEntry(ctx, audit.BuildAccountArchiveEventEntry(userID, accountID))
}

// buildGetAuditLogEntriesForAccountQuery constructs a SQL query for fetching audit log entries
// associated with a given account.
func (q *Postgres) buildGetAuditLogEntriesForAccountQuery(accountID uint64) (query string, args []interface{}) {
	accountIDKey := fmt.Sprintf(jsonPluckQuery, queriers.AuditLogEntriesTableName, queriers.AuditLogEntriesTableContextColumn, audit.AccountAssignmentKey)
	builder := q.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Eq{accountIDKey: accountID}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.CreatedOnColumn))

	return q.buildQuery(builder)
}

// GetAuditLogEntriesForAccount fetches a audit log entries for a given account from the database.
func (q *Postgres) GetAuditLogEntriesForAccount(ctx context.Context, accountID uint64) ([]types.AuditLogEntry, error) {
	query, args := q.buildGetAuditLogEntriesForAccountQuery(accountID)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, _, err := q.scanAuditLogEntries(rows, false)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return auditLogEntries, nil
}
