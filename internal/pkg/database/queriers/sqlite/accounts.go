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

var (
	_ types.AccountDataManager  = (*Sqlite)(nil)
	_ types.AccountAuditManager = (*Sqlite)(nil)
)

// scanAccount takes a database Scanner (i.e. *sql.Row) and scans the result into an Account struct.
func (c *Sqlite) scanAccount(scan database.Scanner, includeCounts bool) (account *types.Account, filteredCount, totalCount uint64, err error) {
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
func (c *Sqlite) scanAccounts(rows database.ResultIterator, includeCounts bool) (accounts []*types.Account, filteredCount, totalCount uint64, err error) {
	for rows.Next() {
		x, fc, tc, scanErr := c.scanAccount(rows, includeCounts)
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

		accounts = append(accounts, x)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, 0, 0, rowsErr
	}

	if closeErr := rows.Close(); closeErr != nil {
		c.logger.Error(closeErr, "closing database rows")
		return nil, 0, 0, closeErr
	}

	return accounts, filteredCount, totalCount, nil
}

// buildGetAccountQuery constructs a SQL query for fetching an account with a given ID belong to a user with a given ID.
func (c *Sqlite) buildGetAccountQuery(accountID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = c.sqlBuilder.
		Select(queriers.AccountsTableColumns...).
		From(queriers.AccountsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.IDColumn):                         accountID,
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.AccountsTableUserOwnershipColumn): userID,
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.ArchivedOnColumn):                 nil,
		}).
		ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// GetAccount fetches an account from the database.
func (c *Sqlite) GetAccount(ctx context.Context, accountID, userID uint64) (*types.Account, error) {
	query, args := c.buildGetAccountQuery(accountID, userID)
	row := c.db.QueryRowContext(ctx, query, args...)

	account, _, _, err := c.scanAccount(row, false)

	return account, err
}

// buildGetAllAccountsCountQuery returns a query that fetches the total number of accounts in the database.
// This query only gets generated once, and is otherwise returned from cache.
func (c *Sqlite) buildGetAllAccountsCountQuery() string {
	var err error

	allAccountsCountQuery, _, err := c.sqlBuilder.
		Select(fmt.Sprintf(columnCountQueryTemplate, queriers.AccountsTableName)).
		From(queriers.AccountsTableName).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.ArchivedOnColumn): nil,
		}).
		ToSql()
	c.logQueryBuildingError(err)

	return allAccountsCountQuery
}

// GetAllAccountsCount will fetch the count of accounts from the database.
func (c *Sqlite) GetAllAccountsCount(ctx context.Context) (count uint64, err error) {
	err = c.db.QueryRowContext(ctx, c.buildGetAllAccountsCountQuery()).Scan(&count)
	return count, err
}

// buildGetBatchOfAccountsQuery returns a query that fetches every account in the database within a bucketed range.
func (c *Sqlite) buildGetBatchOfAccountsQuery(beginID, endID uint64) (query string, args []interface{}) {
	query, args, err := c.sqlBuilder.
		Select(queriers.AccountsTableColumns...).
		From(queriers.AccountsTableName).
		Where(squirrel.Gt{
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.IDColumn): beginID,
		}).
		Where(squirrel.Lt{
			fmt.Sprintf("%s.%s", queriers.AccountsTableName, queriers.IDColumn): endID,
		}).
		ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// GetAllAccounts fetches every account from the database and writes them to a channel. This method primarily exists
// to aid in administrative data tasks.
func (c *Sqlite) GetAllAccounts(ctx context.Context, resultChannel chan []*types.Account, batchSize uint16) error {
	count, countErr := c.GetAllAccountsCount(ctx)
	if countErr != nil {
		return fmt.Errorf("error fetching count of accounts: %w", countErr)
	}

	for beginID := uint64(1); beginID <= count; beginID += uint64(batchSize) {
		endID := beginID + uint64(batchSize)
		go func(begin, end uint64) {
			query, args := c.buildGetBatchOfAccountsQuery(begin, end)
			logger := c.logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, queryErr := c.db.Query(query, args...)
			if errors.Is(queryErr, sql.ErrNoRows) {
				return
			} else if queryErr != nil {
				logger.Error(queryErr, "querying for database rows")
				return
			}

			accounts, _, _, scanErr := c.scanAccounts(rows, false)
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
func (c *Sqlite) buildGetAccountsQuery(userID uint64, forAdmin bool, filter *types.QueryFilter) (query string, args []interface{}) {
	return c.buildListQuery(
		queriers.AccountsTableName,
		queriers.AccountsTableUserOwnershipColumn,
		queriers.AccountsTableColumns,
		userID,
		forAdmin,
		filter,
	)
}

// GetAccounts fetches a list of accounts from the database that meet a particular filter.
func (c *Sqlite) GetAccounts(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.AccountList, error) {
	query, args := c.buildGetAccountsQuery(userID, false, filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for accounts: %w", err)
	}

	accounts, filteredCount, totalCount, err := c.scanAccounts(rows, true)
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
func (c *Sqlite) GetAccountsForAdmin(ctx context.Context, filter *types.QueryFilter) (*types.AccountList, error) {
	query, args := c.buildGetAccountsQuery(0, true, filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetching accounts from database: %w", err)
	}

	accounts, filteredCount, totalCount, err := c.scanAccounts(rows, true)
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
func (c *Sqlite) buildCreateAccountQuery(input *types.Account) (query string, args []interface{}) {
	var err error

	query, args, err = c.sqlBuilder.
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

	c.logQueryBuildingError(err)

	return query, args
}

// CreateAccount creates an account in the database.
func (c *Sqlite) CreateAccount(ctx context.Context, input *types.AccountCreationInput) (*types.Account, error) {
	x := &types.Account{
		Name:          input.Name,
		BelongsToUser: input.BelongsToUser,
	}

	query, args := c.buildCreateAccountQuery(x)

	// create the account.
	res, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing account creation query: %w", err)
	}

	x.CreatedOn = c.timeTeller.Now()
	x.ID = c.getIDFromResult(res)

	return x, nil
}

// buildUpdateAccountQuery takes an account and returns an update SQL query, with the relevant query parameters.
func (c *Sqlite) buildUpdateAccountQuery(input *types.Account) (query string, args []interface{}) {
	var err error

	query, args, err = c.sqlBuilder.
		Update(queriers.AccountsTableName).
		Set(queriers.AccountsTableNameColumn, input.Name).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                         input.ID,
			queriers.AccountsTableUserOwnershipColumn: input.BelongsToUser,
		}).
		ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// UpdateAccount updates a particular account. Note that UpdateAccount expects the provided input to have a valid ID.
func (c *Sqlite) UpdateAccount(ctx context.Context, input *types.Account) error {
	query, args := c.buildUpdateAccountQuery(input)
	_, err := c.db.ExecContext(ctx, query, args...)

	return err
}

// buildArchiveAccountQuery returns a SQL query which marks a given account belonging to a given user as archived.
func (c *Sqlite) buildArchiveAccountQuery(accountID, userID uint64) (query string, args []interface{}) {
	var err error

	query, args, err = c.sqlBuilder.
		Update(queriers.AccountsTableName).
		Set(queriers.LastUpdatedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Set(queriers.ArchivedOnColumn, squirrel.Expr(currentUnixTimeQuery)).
		Where(squirrel.Eq{
			queriers.IDColumn:                         accountID,
			queriers.ArchivedOnColumn:                 nil,
			queriers.AccountsTableUserOwnershipColumn: userID,
		}).
		ToSql()

	c.logQueryBuildingError(err)

	return query, args
}

// ArchiveAccount marks an account as archived in the database.
func (c *Sqlite) ArchiveAccount(ctx context.Context, accountID, userID uint64) error {
	query, args := c.buildArchiveAccountQuery(accountID, userID)

	res, err := c.db.ExecContext(ctx, query, args...)
	if res != nil {
		if rowCount, rowCountErr := res.RowsAffected(); rowCountErr == nil && rowCount == 0 {
			return sql.ErrNoRows
		}
	}

	return err
}

// LogAccountCreationEvent saves a AccountCreationEvent in the audit log table.
func (c *Sqlite) LogAccountCreationEvent(ctx context.Context, account *types.Account) {
	c.createAuditLogEntry(ctx, audit.BuildAccountCreationEventEntry(account))
}

// LogAccountUpdateEvent saves a AccountUpdateEvent in the audit log table.
func (c *Sqlite) LogAccountUpdateEvent(ctx context.Context, userID, accountID uint64, changes []types.FieldChangeSummary) {
	c.createAuditLogEntry(ctx, audit.BuildAccountUpdateEventEntry(userID, accountID, changes))
}

// LogAccountArchiveEvent saves a AccountArchiveEvent in the audit log table.
func (c *Sqlite) LogAccountArchiveEvent(ctx context.Context, userID, accountID uint64) {
	c.createAuditLogEntry(ctx, audit.BuildAccountArchiveEventEntry(userID, accountID))
}

// buildGetAuditLogEntriesForAccountQuery constructs a SQL query for fetching an audit log entry with a given ID belong to a user with a given ID.
func (c *Sqlite) buildGetAuditLogEntriesForAccountQuery(accountID uint64) (query string, args []interface{}) {
	var err error

	accountIDKey := fmt.Sprintf(jsonPluckQuery, queriers.AuditLogEntriesTableName, queriers.AuditLogEntriesTableContextColumn, audit.AccountAssignmentKey)
	builder := c.sqlBuilder.
		Select(queriers.AuditLogEntriesTableColumns...).
		From(queriers.AuditLogEntriesTableName).
		Where(squirrel.Eq{accountIDKey: accountID}).
		OrderBy(fmt.Sprintf("%s.%s", queriers.AuditLogEntriesTableName, queriers.CreatedOnColumn))

	query, args, err = builder.ToSql()
	c.logQueryBuildingError(err)

	return query, args
}

// GetAuditLogEntriesForAccount fetches an audit log entry from the database.
func (c *Sqlite) GetAuditLogEntriesForAccount(ctx context.Context, accountID uint64) ([]*types.AuditLogEntry, error) {
	query, args := c.buildGetAuditLogEntriesForAccountQuery(accountID)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, _, err := c.scanAuditLogEntries(rows, false)
	if err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return auditLogEntries, nil
}
