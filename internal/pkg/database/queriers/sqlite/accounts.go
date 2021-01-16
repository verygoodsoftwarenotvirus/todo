package sqlite

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

var _ types.AccountDataManager = (*Sqlite)(nil)

// scanAccount takes a database Scanner (i.e. *sql.Row) and scans the result into an Account struct.
func (q *Sqlite) scanAccount(scan database.Scanner, includeCounts bool) (account *types.Account, filteredCount, totalCount uint64, err error) {
	account = &types.Account{}

	targetVars := []interface{}{
		&account.ID,
		&account.Name,
		&account.PlanID,
		&account.PersonalAccount,
		&account.CreatedOn,
		&account.LastUpdatedOn,
		&account.ArchivedOn,
		&account.BelongsToUser,
	}

	if includeCounts {
		targetVars = append(targetVars, &filteredCount, &totalCount)
	}

	if scanErr := scan.Scan(targetVars...); scanErr != nil {
		return nil, 0, 0, scanErr
	}

	return account, filteredCount, totalCount, nil
}

// scanAccounts takes some database rows and turns them into a slice of accounts.
func (q *Sqlite) scanAccounts(rows database.ResultIterator, includeCounts bool) (accounts []types.Account, filteredCount, totalCount uint64, err error) {
	for rows.Next() {
		x, fc, tc, scanErr := q.scanAccount(rows, includeCounts)
		if scanErr != nil {
			return nil, 0, 0, scanErr
		}

		if includeCounts {
			if filteredCount == 0 {
				filteredCount = fc
			}

			if totalCount == 0 {
				totalCount = tc
			}
		}

		accounts = append(accounts, *x)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, 0, 0, rowsErr
	}

	if closeErr := rows.Close(); closeErr != nil {
		q.logger.Error(closeErr, "closing database rows")
		return nil, 0, 0, closeErr
	}

	return accounts, filteredCount, totalCount, nil
}

// buildAccountExistsQuery constructs a SQL query for checking if an account with a given ID belong to a user with a given ID exists.
func (q *Sqlite) buildAccountExistsQuery(accountID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Select(fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.IDColumn)).
		Prefix(queriers.ExistencePrefix).
		From(queriers.AccountsTableName).
		Suffix(queriers.ExistenceSuffix).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.IDColumn):                         accountID,
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.AccountsTableUserOwnershipColumn): userID,
		}).ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// AccountExists queries the database to see if a given account belonging to a given user exists.
func (q *Sqlite) AccountExists(ctx context.Context, accountID, userID uint64) (exists bool, err error) {
	query, args := q.buildAccountExistsQuery(accountID, userID)

	err = q.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	return exists, err
}

// buildGetAccountQuery constructs a SQL query for fetching an account with a given ID belong to a user with a given ID.
func (q *Sqlite) buildGetAccountQuery(accountID, userID uint64) (query string, args []interface{}) {
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
func (q *Sqlite) GetAccount(ctx context.Context, accountID, userID uint64) (*types.Account, error) {
	query, args := q.buildGetAccountQuery(accountID, userID)
	row := q.db.QueryRowContext(ctx, query, args...)

	account, _, _, err := q.scanAccount(row, false)

	return account, err
}

// buildGetAllAccountsCountQuery returns a query that fetches the total number of accounts in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (q *Sqlite) buildGetAllAccountsCountQuery() string {
	var err error

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
func (q *Sqlite) GetAllAccountsCount(ctx context.Context) (count uint64, err error) {
	err = q.db.QueryRowContext(ctx, q.buildGetAllAccountsCountQuery()).Scan(&count)
	return count, err
}

// buildGetBatchOfAccountsQuery returns a query that fetches every account in the database within a bucketed range.
func (q *Sqlite) buildGetBatchOfAccountsQuery(beginID, endID uint64) (query string, args []interface{}) {
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
func (q *Sqlite) GetAllAccounts(ctx context.Context, resultChannel chan []*types.Account) error {
	count, countErr := q.GetAllAccountsCount(ctx)
	if countErr != nil {
		return fmt.Errorf("error fetching count of accounts: %w", countErr)
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

			rows, queryErr := q.db.Query(query, args...)
			if errors.Is(queryErr, sql.ErrNoRows) {
				return
			} else if queryErr != nil {
				logger.Error(queryErr, "querying for database rows")
				return
			}

			accounts, _, _, scanErr := q.scanAccounts(rows, false)
			if scanErr != nil {
				logger.Error(scanErr, "scanning database rows")
				return
			}

			resultChannel <- accounts
		}(beginID, endID)
	}

	return nil
}

// buildGetAccountsQuery builds a SQL query selecting accounts that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (q *Sqlite) buildGetAccountsQuery(userID uint64, forAdmin bool, filter *types.QueryFilter) (query string, args []interface{}) {
	return q.buildListQuery(
		queriers.AccountsTableName,
		queriers.AccountsTableUserOwnershipColumn,
		queriers.AccountsTableColumns,
		userID,
		forAdmin,
		filter,
	)
}

// GetAccounts fetches a list of accounts from the database that meet a particular filter.
func (q *Sqlite) GetAccounts(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.AccountList, error) {
	query, args := q.buildGetAccountsQuery(userID, false, filter)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for accounts: %w", err)
	}

	accounts, filteredCount, totalCount, err := q.scanAccounts(rows, true)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	list := &types.AccountList{
		Pagination: types.Pagination{
			Page:          filter.Page,
			Limit:         filter.Limit,
			FilteredCount: filteredCount,
			TotalCount:    totalCount,
		},
		Accounts: accounts,
	}

	return list, nil
}

// GetAccountsForAdmin fetches a list of accounts from the database that meet a particular filter for all users.
func (q *Sqlite) GetAccountsForAdmin(ctx context.Context, filter *types.QueryFilter) (*types.AccountList, error) {
	query, args := q.buildGetAccountsQuery(0, true, filter)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetching accounts from database: %w", err)
	}

	accounts, filteredCount, totalCount, err := q.scanAccounts(rows, true)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	list := &types.AccountList{
		Pagination: types.Pagination{
			Page:          filter.Page,
			Limit:         filter.Limit,
			FilteredCount: filteredCount,
			TotalCount:    totalCount,
		},
		Accounts: accounts,
	}

	return list, nil
}

// buildCreateAccountQuery takes an account and returns a creation query for that account and the relevant arguments.
func (q *Sqlite) buildCreateAccountQuery(input *types.Account) (query string, args []interface{}) {
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
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// CreateAccount creates an account in the database.
func (q *Sqlite) CreateAccount(ctx context.Context, input *types.AccountCreationInput) (*types.Account, error) {
	x := &types.Account{
		Name:          input.Name,
		BelongsToUser: input.BelongsToUser,
	}

	query, args := q.buildCreateAccountQuery(x)

	// create the account.
	res, err := q.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing account creation query: %w", err)
	}

	x.CreatedOn = q.timeTeller.Now()
	x.ID = q.getIDFromResult(res)

	return x, nil
}

// buildUpdateAccountQuery takes an account and returns an update SQL query, with the relevant query parameters.
func (q *Sqlite) buildUpdateAccountQuery(input *types.Account) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(queriers.AccountsTableName).
		Set(queriers.AccountsTableNameColumn, input.Name).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                         input.ID,
			queriers.AccountsTableUserOwnershipColumn: input.BelongsToUser,
		}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// UpdateAccount updates a particular account. Note that UpdateAccount expects the provided input to have a valid ID.
func (q *Sqlite) UpdateAccount(ctx context.Context, input *types.Account) error {
	query, args := q.buildUpdateAccountQuery(input)
	_, err := q.db.ExecContext(ctx, query, args...)

	return err
}

// buildArchiveAccountQuery returns a SQL query which marks a given account belonging to a given user as archived.
func (q *Sqlite) buildArchiveAccountQuery(accountID, userID uint64) (query string, args []interface{}) {
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
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// ArchiveAccount marks an account as archived in the database.
func (q *Sqlite) ArchiveAccount(ctx context.Context, accountID, userID uint64) error {
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
func (q *Sqlite) LogAccountCreationEvent(ctx context.Context, account *types.Account) {
	q.createAuditLogEntry(ctx, audit.BuildAccountCreationEventEntry(account))
}

// LogAccountUpdateEvent saves a AccountUpdateEvent in the audit log table.
func (q *Sqlite) LogAccountUpdateEvent(ctx context.Context, userID, accountID uint64, changes []types.FieldChangeSummary) {
	q.createAuditLogEntry(ctx, audit.BuildAccountUpdateEventEntry(userID, accountID, changes))
}

// LogAccountArchiveEvent saves a AccountArchiveEvent in the audit log table.
func (q *Sqlite) LogAccountArchiveEvent(ctx context.Context, userID, accountID uint64) {
	q.createAuditLogEntry(ctx, audit.BuildAccountArchiveEventEntry(userID, accountID))
}

// buildGetAuditLogEntriesForAccountQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (q *Sqlite) buildGetAuditLogEntriesForAccountQuery(accountID uint64) (query string, args []interface{}) {
	var err error

	accountIDKey := fmt.Sprintf(jsonPluckQuery, queriers.AuditLogEntriesTableName, queriers.AuditLogEntriesTableContextColumn, audit.AccountAssignmentKey)
	builder := q.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Eq{accountIDKey: accountID}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.CreatedOnColumn))

	query, args, err = builder.ToSql()
	q.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntriesForAccount fetches an audit log entry from the database.
func (q *Sqlite) GetAuditLogEntriesForAccount(ctx context.Context, accountID uint64) ([]types.AuditLogEntry, error) {
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
