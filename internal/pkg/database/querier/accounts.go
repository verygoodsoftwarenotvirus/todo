package querier

import (
	"context"
	"database/sql"
	"errors"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.AccountDataManager = (*Client)(nil)
)

// scanAccount takes a database Scanner (i.e. *sql.Row) and scans the result into an Account struct.
func (c *Client) scanAccount(ctx context.Context, scan database.Scanner, includeCounts bool) (account *types.Account, filteredCount, totalCount uint64, err error) {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue("include_counts", includeCounts)

	account = &types.Account{}

	targetVars := []interface{}{
		&account.ID,
		&account.ExternalID,
		&account.Name,
		&account.PlanID,
		&account.DefaultUserPermissions,
		&account.CreatedOn,
		&account.LastUpdatedOn,
		&account.ArchivedOn,
		&account.BelongsToUser,
	}

	if includeCounts {
		targetVars = append(targetVars, &filteredCount, &totalCount)
	}

	if err = scan.Scan(targetVars...); err != nil {
		return nil, 0, 0, observability.PrepareError(err, logger, span, "fetching memberships from database")
	}

	return account, filteredCount, totalCount, nil
}

// scanAccounts takes some database rows and turns them into a slice of accounts.
func (c *Client) scanAccounts(ctx context.Context, rows database.ResultIterator, includeCounts bool) (accounts []*types.Account, filteredCount, totalCount uint64, err error) {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue("include_counts", includeCounts)

	for rows.Next() {
		x, fc, tc, scanErr := c.scanAccount(ctx, rows, includeCounts)
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

	if err = c.checkRowsForErrorAndClose(ctx, rows); err != nil {
		return nil, 0, 0, observability.PrepareError(err, logger, span, "handling rows")
	}

	return accounts, filteredCount, totalCount, nil
}

// GetAccount fetches an account from the database.
func (c *Client) GetAccount(ctx context.Context, accountID, userID uint64) (*types.Account, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAccountIDToSpan(span, accountID)
	tracing.AttachUserIDToSpan(span, userID)

	logger := c.logger.WithValues(map[string]interface{}{
		keys.AccountIDKey: accountID,
		keys.UserIDKey:    userID,
	})

	query, args := c.sqlQueryBuilder.BuildGetAccountQuery(accountID, userID)
	row := c.db.QueryRowContext(ctx, query, args...)

	account, _, _, err := c.scanAccount(ctx, row, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "beginning transaction")
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

	count, err := c.GetAllAccountsCount(ctx)
	if err != nil {
		return observability.PrepareError(err, c.logger.WithValue("batch_size", batchSize), span, "fetching count of accounts")
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

			rows, err := c.db.Query(query, args...)
			if errors.Is(err, sql.ErrNoRows) {
				return
			} else if err != nil {
				observability.AcknowledgeError(err, logger, span, "querying for database rows")
				return
			}

			accounts, _, _, err := c.scanAccounts(ctx, rows, false)
			if err != nil {
				observability.AcknowledgeError(err, logger, span, "scanning database rows")
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

	logger := filter.AttachToLogger(c.logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	x = &types.AccountList{}
	if filter != nil {
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	tracing.AttachUserIDToSpan(span, userID)

	c.logger.WithValues(map[string]interface{}{
		keys.UserIDKey: userID,
	}).Debug("GetAccounts called")

	query, args := c.sqlQueryBuilder.BuildGetAccountsQuery(userID, false, filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "executing accounts list retrieval query")
	}

	if x.Accounts, x.FilteredCount, x.TotalCount, err = c.scanAccounts(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning accounts from database")
	}

	return x, nil
}

// GetAccountsForAdmin fetches a list of accounts from the database that meet a particular filter for all users.
func (c *Client) GetAccountsForAdmin(ctx context.Context, filter *types.QueryFilter) (x *types.AccountList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := filter.AttachToLogger(c.logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	x = &types.AccountList{}
	if filter != nil {
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := c.sqlQueryBuilder.BuildGetAccountsQuery(0, true, filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "querying database for accounts")
	}

	if x.Accounts, x.FilteredCount, x.TotalCount, err = c.scanAccounts(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning accounts")
	}

	return x, nil
}

// CreateAccount creates an account in the database.
func (c *Client) CreateAccount(ctx context.Context, input *types.AccountCreationInput, createdByUser uint64) (*types.Account, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue(keys.RequesterKey, createdByUser).WithValue(keys.UserIDKey, input.BelongsToUser)
	accountCreationQuery, accountCreationArgs := c.sqlQueryBuilder.BuildAccountCreationQuery(input)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "beginning transaction")
	}

	// create the account.
	id, err := c.performWriteQuery(ctx, tx, false, "account creation", accountCreationQuery, accountCreationArgs)
	if err != nil {
		c.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "creating account")
	}

	logger = logger.WithValue(keys.AccountIDKey, id)

	x := &types.Account{
		ID:                     id,
		Name:                   input.Name,
		BelongsToUser:          input.BelongsToUser,
		DefaultUserPermissions: input.DefaultUserPermissions,
		CreatedOn:              c.currentTime(),
	}

	logger.Debug("account created")

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountCreationEventEntry(x, createdByUser)); err != nil {
		c.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "writing account creation audit log event entry")
	}

	addInput := &types.AddUserToAccountInput{
		UserID:                 input.BelongsToUser,
		UserAccountPermissions: x.DefaultUserPermissions,
		Reason:                 "account creation",
	}

	addUserToAccountQuery, addUserToAccountArgs := c.sqlQueryBuilder.BuildAddUserToAccountQuery(x.ID, addInput)
	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "account user membership creation", addUserToAccountQuery, addUserToAccountArgs); err != nil {
		c.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "creating account membership")
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserAddedToAccountEventEntry(createdByUser, x.ID, addInput)); err != nil {
		c.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "writing account membership creation audit log event entry")
	}

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareError(err, logger, span, "committing transaction")
	}

	logger.Debug("account created")

	return x, nil
}

// UpdateAccount updates a particular account. Note that UpdateAccount expects the
// provided input to have a valid ID.
func (c *Client) UpdateAccount(ctx context.Context, updated *types.Account, changedByUser uint64, changes []types.FieldChangeSummary) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachAccountIDToSpan(span, updated.ID)
	logger := c.logger.WithValue(keys.AccountIDKey, updated.ID)

	query, args := c.sqlQueryBuilder.BuildUpdateAccountQuery(updated)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "account update", query, args); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "updating account")
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountUpdateEventEntry(updated.BelongsToUser, updated.ID, changedByUser, changes)); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing account update audit log event entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	return nil
}

// ArchiveAccount archives an account from the database by its ID.
func (c *Client) ArchiveAccount(ctx context.Context, accountID, userID, archivedByUser uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)

	logger := c.logger.WithValues(map[string]interface{}{
		keys.RequesterKey: archivedByUser,
		keys.AccountIDKey: accountID,
		keys.UserIDKey:    userID,
	})

	logger.Debug("ArchiveAccount called")

	query, args := c.sqlQueryBuilder.BuildArchiveAccountQuery(accountID, userID)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "account archive", query, args); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "archiving account")
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountArchiveEventEntry(userID, accountID, archivedByUser)); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing account archive audit log event entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	return nil
}

// GetAuditLogEntriesForAccount fetches a list of audit log entries from the database that relate to a given account.
func (c *Client) GetAuditLogEntriesForAccount(ctx context.Context, accountID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	query, args := c.sqlQueryBuilder.BuildGetAuditLogEntriesForAccountQuery(accountID)

	logger := c.logger.WithValue(keys.AccountIDKey, accountID)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "querying database for audit log entries")
	}

	auditLogEntries, _, err := c.scanAuditLogEntries(ctx, rows, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning audit log entries")
	}

	return auditLogEntries, nil
}
