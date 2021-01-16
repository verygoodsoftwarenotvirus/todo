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
func (q *Postgres) scanAccount(scan database.Scanner, includeCounts bool) (account *types.Account, filteredCount, totalCount uint64, err error) {
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
func (q *Postgres) scanAccounts(rows database.ResultIterator, includeCounts bool) (accounts []types.Account, filteredCount, totalCount uint64, err error) {
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

// BuildGetAccountQuery constructs a SQL query for fetching an account with a given ID belong to a user with a given ID.
func (q *Postgres) BuildGetAccountQuery(accountID, userID uint64) (query string, args []interface{}) {
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
	query, args := q.BuildGetAccountQuery(accountID, userID)
	row := q.db.QueryRowContext(ctx, query, args...)

	account, _, _, err := q.scanAccount(row, false)

	return account, err
}

// BuildGetAllAccountsCountQuery returns a query that fetches the total number of accounts in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (q *Postgres) BuildGetAllAccountsCountQuery() string {
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
	err = q.db.QueryRowContext(ctx, q.BuildGetAllAccountsCountQuery()).Scan(&count)
	return count, err
}

// BuildGetBatchOfAccountsQuery returns a query that fetches every account in the database within a bucketed range.
func (q *Postgres) BuildGetBatchOfAccountsQuery(beginID, endID uint64) (query string, args []interface{}) {
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
func (q *Postgres) GetAllAccounts(ctx context.Context, resultChannel chan []*types.Account, bucketSize uint16) error {
	count, countErr := q.GetAllAccountsCount(ctx)
	if countErr != nil {
		return fmt.Errorf("error fetching count of webhooks: %w", countErr)
	}

	for beginID := uint64(1); beginID <= count; beginID += uint64(bucketSize) {
		endID := beginID + uint64(bucketSize)
		go func(begin, end uint64) {
			query, args := q.BuildGetBatchOfAccountsQuery(begin, end)
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

// BuildGetAccountsQuery builds a SQL query selecting accounts that adhere to a given QueryFilter and belong to a given user,
// and returns both the query and the relevant args to pass to the query executor.
func (q *Postgres) BuildGetAccountsQuery(userID uint64, forAdmin bool, filter *types.QueryFilter) (query string, args []interface{}) {
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
func (q *Postgres) GetAccounts(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.AccountList, error) {
	query, args := q.BuildGetAccountsQuery(userID, false, filter)

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
func (q *Postgres) GetAccountsForAdmin(ctx context.Context, filter *types.QueryFilter) (*types.AccountList, error) {
	query, args := q.BuildGetAccountsQuery(0, true, filter)

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

// BuildCreateAccountQuery takes an account and returns a creation query for that account and the relevant arguments.
func (q *Postgres) BuildCreateAccountQuery(input *types.Account) (query string, args []interface{}) {
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

	query, args := q.BuildCreateAccountQuery(x)

	// create the account.
	err := q.db.QueryRowContext(ctx, query, args...).Scan(&x.ID, &x.CreatedOn)
	if err != nil {
		return nil, fmt.Errorf("error executing account creation query: %w", err)
	}

	return x, nil
}

// BuildUpdateAccountQuery takes an account and returns an update SQL query, with the relevant query parameters.
func (q *Postgres) BuildUpdateAccountQuery(input *types.Account) (query string, args []interface{}) {
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
	query, args := q.BuildUpdateAccountQuery(input)
	return q.db.QueryRowContext(ctx, query, args...).Scan(&input.LastUpdatedOn)
}

// BuildArchiveAccountQuery returns a SQL query which marks a given account belonging to a given user as archived.
func (q *Postgres) BuildArchiveAccountQuery(accountID, userID uint64) (query string, args []interface{}) {
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
	query, args := q.BuildArchiveAccountQuery(accountID, userID)

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
	q.CreateAuditLogEntry(ctx, audit.BuildAccountCreationEventEntry(account))
}

// LogAccountUpdateEvent saves a AccountUpdateEvent in the audit log table.
func (q *Postgres) LogAccountUpdateEvent(ctx context.Context, userID, accountID uint64, changes []types.FieldChangeSummary) {
	q.CreateAuditLogEntry(ctx, audit.BuildAccountUpdateEventEntry(userID, accountID, changes))
}

// LogAccountArchiveEvent saves a AccountArchiveEvent in the audit log table.
func (q *Postgres) LogAccountArchiveEvent(ctx context.Context, userID, accountID uint64) {
	q.CreateAuditLogEntry(ctx, audit.BuildAccountArchiveEventEntry(userID, accountID))
}

// BuildGetAuditLogEntriesForAccountQuery constructs a SQL query for fetching audit log entries
// associated with a given account.
func (q *Postgres) BuildGetAuditLogEntriesForAccountQuery(accountID uint64) (query string, args []interface{}) {
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
	query, args := q.BuildGetAuditLogEntriesForAccountQuery(accountID)

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
