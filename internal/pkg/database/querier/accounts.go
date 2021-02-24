package querier

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.AccountDataManager = (*Client)(nil)
)

// scanAccount takes a database Scanner (i.e. *sql.Row) and scans the result into an Account struct.
func (c *Client) scanAccount(scan database.Scanner, includeCounts bool) (account *types.Account, filteredCount, totalCount uint64, err error) {
	account = &types.Account{}

	targetVars := []interface{}{
		&account.ID,
		&account.ExternalID,
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
func (c *Client) scanAccounts(rows database.ResultIterator, includeCounts bool) (accounts []*types.Account, filteredCount, totalCount uint64, err error) {
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

	if handleErr := c.handleRows(rows); handleErr != nil {
		return nil, 0, 0, handleErr
	}

	return accounts, filteredCount, totalCount, nil
}

// GetAccount fetches an account from the database.
func (c *Client) GetAccount(ctx context.Context, accountID, userID uint64) (*types.Account, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAccountIDToSpan(span, accountID)
	tracing.AttachUserIDToSpan(span, userID)

	c.logger.WithValues(map[string]interface{}{
		keys.AccountIDKey: accountID,
		keys.UserIDKey:    userID,
	}).Debug("GetAccount called")

	query, args := c.sqlQueryBuilder.BuildGetAccountQuery(accountID, userID)
	row := c.db.QueryRowContext(ctx, query, args...)

	account, _, _, err := c.scanAccount(row, false)
	if err != nil {
		return nil, fmt.Errorf("scanning account: %w", err)
	}

	return account, nil
}

// GetAllAccountsCount fetches the count of accounts from the database that meet a particular filter.
func (c *Client) GetAllAccountsCount(ctx context.Context) (count uint64, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllAccountsCount called")

	return c.performCountQuery(ctx, c.db, c.sqlQueryBuilder.BuildGetAllAccountsCountQuery())
}

// GetAllAccounts fetches a list of all accounts in the database.
func (c *Client) GetAllAccounts(ctx context.Context, results chan []*types.Account, batchSize uint16) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllAccounts called")

	count, countErr := c.GetAllAccountsCount(ctx)
	if countErr != nil {
		return fmt.Errorf("fetching count of accounts: %w", countErr)
	}

	for beginID := uint64(1); beginID <= count; beginID += uint64(batchSize) {
		endID := beginID + uint64(batchSize)
		go func(begin, end uint64) {
			query, args := c.sqlQueryBuilder.BuildGetBatchOfAccountsQuery(begin, end)
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

			results <- accounts
		}(beginID, endID)
	}

	return nil
}

// GetAccounts fetches a list of accounts from the database that meet a particular filter.
func (c *Client) GetAccounts(ctx context.Context, userID uint64, filter *types.QueryFilter) (x *types.AccountList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	x = &types.AccountList{}

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit)
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	tracing.AttachUserIDToSpan(span, userID)

	c.logger.WithValues(map[string]interface{}{
		keys.UserIDKey: userID,
	}).Debug("GetAccounts called")

	query, args := c.sqlQueryBuilder.BuildGetAccountsQuery(userID, false, filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing accounts list retrieval query: %w", err)
	}

	if x.Accounts, x.FilteredCount, x.TotalCount, err = c.scanAccounts(rows, true); err != nil {
		return nil, fmt.Errorf("scanning response from database: %w", err)
	}

	return x, nil
}

// GetAccountsForAdmin fetches a list of accounts from the database that meet a particular filter for all users.
func (c *Client) GetAccountsForAdmin(ctx context.Context, filter *types.QueryFilter) (x *types.AccountList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	x = &types.AccountList{}

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit)
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	c.logger.Debug("GetAccounts called")

	query, args := c.sqlQueryBuilder.BuildGetAccountsQuery(0, true, filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing accounts list retrieval query for admin: %w", err)
	}

	if x.Accounts, x.FilteredCount, x.TotalCount, err = c.scanAccounts(rows, true); err != nil {
		return nil, fmt.Errorf("scanning accounts: %w", err)
	}

	return x, nil
}

// CreateAccount creates an account in the database.
func (c *Client) CreateAccount(ctx context.Context, input *types.AccountCreationInput, createdByUser uint64) (*types.Account, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("CreateAccount called")

	query, args := c.sqlQueryBuilder.BuildCreateAccountQuery(input)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error beginning transaction: %w", err)
	}

	// create the account.
	id, err := c.performWriteQuery(ctx, tx, false, "account creation", query, args)
	if err != nil {
		c.rollbackTransaction(tx)
		return nil, err
	}

	x := &types.Account{
		ID:            id,
		Name:          input.Name,
		BelongsToUser: input.BelongsToUser,
		CreatedOn:     c.currentTime(),
	}

	c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountCreationEventEntry(x))

	if commitErr := tx.Commit(); commitErr != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return x, nil
}

// UpdateAccount updates a particular account. Note that UpdateAccount expects the
// provided input to have a valid ID.
func (c *Client) UpdateAccount(ctx context.Context, updated *types.Account, changedByUser uint64, changes []types.FieldChangeSummary) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAccountIDToSpan(span, updated.ID)
	c.logger.WithValue(keys.AccountIDKey, updated.ID).Debug("UpdateAccount called")

	query, args := c.sqlQueryBuilder.BuildUpdateAccountQuery(updated)

	tx, transactionStartErr := c.db.BeginTx(ctx, nil)
	if transactionStartErr != nil {
		return fmt.Errorf("error beginning transaction: %w", transactionStartErr)
	}

	if execErr := c.performWriteQueryIgnoringReturn(ctx, tx, "account update", query, args); execErr != nil {
		c.rollbackTransaction(tx)
		return fmt.Errorf("error updating account: %w", execErr)
	}

	c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountUpdateEventEntry(updated.BelongsToUser, updated.ID, changes))

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("error committing transaction: %w", commitErr)
	}

	return nil
}

// ArchiveAccount archives an account from the database by its ID.
func (c *Client) ArchiveAccount(ctx context.Context, accountID, userID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)

	c.logger.WithValues(map[string]interface{}{
		keys.AccountIDKey: accountID,
		keys.UserIDKey:    userID,
	}).Debug("ArchiveAccount called")

	query, args := c.sqlQueryBuilder.BuildArchiveAccountQuery(accountID, userID)

	tx, transactionStartErr := c.db.BeginTx(ctx, nil)
	if transactionStartErr != nil {
		return fmt.Errorf("error beginning transaction: %w", transactionStartErr)
	}

	if execErr := c.performWriteQueryIgnoringReturn(ctx, tx, "account archive", query, args); execErr != nil {
		c.rollbackTransaction(tx)
		return fmt.Errorf("error updating account: %w", execErr)
	}

	c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountArchiveEventEntry(userID, accountID))

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("error committing transaction: %w", commitErr)
	}

	return nil
}

// GetAuditLogEntriesForAccount fetches a list of audit log entries from the database that relate to a given account.
func (c *Client) GetAuditLogEntriesForAccount(ctx context.Context, accountID uint64) ([]*types.AuditLogEntry, error) {
	query, args := c.sqlQueryBuilder.BuildGetAuditLogEntriesForAccountQuery(accountID)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying database for audit log entries: %w", err)
	}

	auditLogEntries, _, err := c.scanAuditLogEntries(rows, false)
	if err != nil {
		return nil, fmt.Errorf("scanning audit log entries: %w", err)
	}

	return auditLogEntries, nil
}
