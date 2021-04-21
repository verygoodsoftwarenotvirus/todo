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
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.AccountDataManager = (*SQLQuerier)(nil)
)

// scanAccount takes a database Scanner (i.e. *sql.Row) and scans the result into an Account struct.
func (q *SQLQuerier) scanAccount(ctx context.Context, scan database.Scanner, includeCounts bool) (account *types.Account, membership *types.AccountUserMembership, filteredCount, totalCount uint64, err error) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("include_counts", includeCounts)

	account = &types.Account{Members: []*types.AccountUserMembership{}}
	membership = &types.AccountUserMembership{}

	var perms int64

	targetVars := []interface{}{
		&account.ID,
		&account.ExternalID,
		&account.Name,
		&account.AccountSubscriptionPlanID,
		&account.DefaultNewMemberPermissions,
		&account.CreatedOn,
		&account.LastUpdatedOn,
		&account.ArchivedOn,
		&account.BelongsToUser,
		&membership.ID,
		&membership.BelongsToUser,
		&membership.BelongsToAccount,
		&perms,
		&membership.DefaultAccount,
		&membership.CreatedOn,
		&membership.LastUpdatedOn,
		&membership.ArchivedOn,
	}

	if includeCounts {
		targetVars = append(targetVars, &filteredCount, &totalCount)
	}

	if err = scan.Scan(targetVars...); err != nil {
		return nil, nil, 0, 0, observability.PrepareError(err, logger, span, "fetching memberships from database")
	}

	membership.UserAccountPermissions = permissions.ServiceUserPermission(perms)

	return account, membership, filteredCount, totalCount, nil
}

// scanAccounts takes some database rows and turns them into a slice of accounts.
func (q *SQLQuerier) scanAccounts(ctx context.Context, rows database.ResultIterator, includeCounts bool) (accounts []*types.Account, filteredCount, totalCount uint64, err error) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("include_counts", includeCounts)

	accounts = []*types.Account{}

	var currentAccount *types.Account
	for rows.Next() {
		account, membership, fc, tc, scanErr := q.scanAccount(ctx, rows, includeCounts)
		if scanErr != nil {
			return nil, 0, 0, scanErr
		}

		if currentAccount == nil {
			currentAccount = account
		}

		if currentAccount.ID != account.ID {
			accounts = append(accounts, currentAccount)
			currentAccount = account
		}

		currentAccount.Members = append(currentAccount.Members, membership)

		if includeCounts {
			if filteredCount == 0 {
				filteredCount = fc
			}

			if totalCount == 0 {
				totalCount = tc
			}
		}
	}

	if currentAccount != nil {
		accounts = append(accounts, currentAccount)
	}

	if err = q.checkRowsForErrorAndClose(ctx, rows); err != nil {
		return nil, 0, 0, observability.PrepareError(err, logger, span, "handling rows")
	}

	return accounts, filteredCount, totalCount, nil
}

// GetAccount fetches an account from the database.
func (q *SQLQuerier) GetAccount(ctx context.Context, accountID, userID uint64) (*types.Account, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 || userID == 0 {
		return nil, ErrInvalidIDProvided
	}

	tracing.AttachAccountIDToSpan(span, accountID)
	tracing.AttachUserIDToSpan(span, userID)

	logger := q.logger.WithValues(map[string]interface{}{
		keys.AccountIDKey: accountID,
		keys.UserIDKey:    userID,
	})

	query, args := q.sqlQueryBuilder.BuildGetAccountQuery(ctx, accountID, userID)
	rows, err := q.performReadQuery(ctx, q.db, "account", query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "executing accounts list retrieval query")
	}

	accounts, _, _, err := q.scanAccounts(ctx, rows, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "beginning transaction")
	}

	var account *types.Account
	if len(accounts) > 0 {
		account = accounts[0]
	}

	if account == nil {
		return nil, sql.ErrNoRows
	}

	return account, nil
}

// GetAllAccountsCount fetches the count of accounts from the database that meet a particular filter.
func (q *SQLQuerier) GetAllAccountsCount(ctx context.Context) (uint64, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger

	count, err := q.performCountQuery(ctx, q.db, q.sqlQueryBuilder.BuildGetAllAccountsCountQuery(ctx), "fetching count of all accounts")
	if err != nil {
		return 0, observability.PrepareError(err, logger, span, "querying for count of accounts")
	}

	return count, nil
}

// GetAllAccounts fetches a list of all accounts in the database.
func (q *SQLQuerier) GetAllAccounts(ctx context.Context, results chan []*types.Account, batchSize uint16) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if results == nil {
		return ErrNilInputProvided
	}

	if batchSize == 0 {
		batchSize = defaultBatchSize
	}

	logger := q.logger.WithValue("batch_size", batchSize)

	count, err := q.GetAllAccountsCount(ctx)
	if err != nil {
		return observability.PrepareError(err, logger, span, "fetching count of accounts")
	}

	for beginID := uint64(1); beginID <= count; beginID += uint64(batchSize) {
		endID := beginID + uint64(batchSize)
		go func(begin, end uint64) {
			query, args := q.sqlQueryBuilder.BuildGetBatchOfAccountsQuery(ctx, begin, end)
			logger = logger.WithValues(map[string]interface{}{
				"query": query,
				"begin": begin,
				"end":   end,
			})

			rows, queryErr := q.db.Query(query, args...)
			if errors.Is(queryErr, sql.ErrNoRows) {
				return
			} else if queryErr != nil {
				observability.AcknowledgeError(queryErr, logger, span, "querying for database rows")
				return
			}

			accounts, _, _, scanErr := q.scanAccounts(ctx, rows, false)
			if scanErr != nil {
				observability.AcknowledgeError(scanErr, logger, span, "scanning database rows")
				return
			}

			results <- accounts
		}(beginID, endID)
	}

	return nil
}

// GetAccounts fetches a list of accounts from the database that meet a particular filter.
func (q *SQLQuerier) GetAccounts(ctx context.Context, userID uint64, filter *types.QueryFilter) (x *types.AccountList, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return nil, ErrInvalidIDProvided
	}

	logger := filter.AttachToLogger(q.logger).WithValue(keys.UserIDKey, userID)
	tracing.AttachQueryFilterToSpan(span, filter)
	tracing.AttachUserIDToSpan(span, userID)

	x = &types.AccountList{}
	if filter != nil {
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := q.sqlQueryBuilder.BuildGetAccountsQuery(ctx, userID, false, filter)

	rows, err := q.performReadQuery(ctx, q.db, "accounts", query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "executing accounts list retrieval query")
	}

	if x.Accounts, x.FilteredCount, x.TotalCount, err = q.scanAccounts(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning accounts from database")
	}

	return x, nil
}

// GetAccountsForAdmin fetches a list of accounts from the database that meet a particular filter for all users.
func (q *SQLQuerier) GetAccountsForAdmin(ctx context.Context, filter *types.QueryFilter) (x *types.AccountList, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := filter.AttachToLogger(q.logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	x = &types.AccountList{}
	if filter != nil {
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := q.sqlQueryBuilder.BuildGetAccountsQuery(ctx, 0, true, filter)

	rows, err := q.performReadQuery(ctx, q.db, "accounts for admin", query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "querying database for accounts")
	}

	if x.Accounts, x.FilteredCount, x.TotalCount, err = q.scanAccounts(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning accounts")
	}

	return x, nil
}

// CreateAccount creates an account in the database.
func (q *SQLQuerier) CreateAccount(ctx context.Context, input *types.AccountCreationInput, createdByUser uint64) (*types.Account, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if createdByUser == 0 {
		return nil, ErrInvalidIDProvided
	}

	if input == nil {
		return nil, ErrNilInputProvided
	}

	logger := q.logger.WithValue(keys.RequesterIDKey, createdByUser).WithValue(keys.UserIDKey, input.BelongsToUser)
	tracing.AttachRequestingUserIDToSpan(span, createdByUser)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "beginning transaction")
	}

	accountCreationQuery, accountCreationArgs := q.sqlQueryBuilder.BuildAccountCreationQuery(ctx, input)

	// create the account.
	id, err := q.performWriteQuery(ctx, tx, false, "account creation", accountCreationQuery, accountCreationArgs)
	if err != nil {
		q.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "creating account")
	}

	logger = logger.WithValue(keys.AccountIDKey, id)

	account := &types.Account{
		ID:                          id,
		Name:                        input.Name,
		BelongsToUser:               input.BelongsToUser,
		DefaultNewMemberPermissions: input.DefaultUserPermissions,
		CreatedOn:                   q.currentTime(),
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountCreationEventEntry(account, createdByUser)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "writing account creation audit log event entry")
	}

	addInput := &types.AddUserToAccountInput{
		UserID:                 input.BelongsToUser,
		AccountID:              account.ID,
		UserAccountPermissions: account.DefaultNewMemberPermissions,
		Reason:                 "account creation",
	}

	addUserToAccountQuery, addUserToAccountArgs := q.sqlQueryBuilder.BuildAddUserToAccountQuery(ctx, addInput)
	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "account user membership creation", addUserToAccountQuery, addUserToAccountArgs); err != nil {
		q.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "creating account membership")
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserAddedToAccountEventEntry(createdByUser, addInput)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "writing account membership creation audit log event entry")
	}

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareError(err, logger, span, "committing transaction")
	}

	tracing.AttachAccountIDToSpan(span, account.ID)
	logger.Info("account created")

	return account, nil
}

// UpdateAccount updates a particular account. Note that UpdateAccount expects the provided input to have a valid ID.
func (q *SQLQuerier) UpdateAccount(ctx context.Context, updated *types.Account, changedByUser uint64, changes []*types.FieldChangeSummary) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if changedByUser == 0 {
		return ErrInvalidIDProvided
	}

	if updated == nil {
		return ErrNilInputProvided
	}

	logger := q.logger.WithValue(keys.AccountIDKey, updated.ID)
	tracing.AttachAccountIDToSpan(span, updated.ID)
	tracing.AttachRequestingUserIDToSpan(span, changedByUser)
	tracing.AttachChangeSummarySpan(span, "account", changes)

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	query, args := q.sqlQueryBuilder.BuildUpdateAccountQuery(ctx, updated)
	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "account update", query, args); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "updating account")
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountUpdateEventEntry(updated.BelongsToUser, updated.ID, changedByUser, changes)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing account update audit log event entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	logger.Info("account updated")

	return nil
}

// ArchiveAccount archives an account from the database by its ID.
func (q *SQLQuerier) ArchiveAccount(ctx context.Context, accountID, userID, archivedByUser uint64) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 || userID == 0 || archivedByUser == 0 {
		return ErrInvalidIDProvided
	}

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)

	logger := q.logger.WithValues(map[string]interface{}{
		keys.RequesterIDKey: archivedByUser,
		keys.AccountIDKey:   accountID,
		keys.UserIDKey:      userID,
	})

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	query, args := q.sqlQueryBuilder.BuildArchiveAccountQuery(ctx, accountID, userID)

	if err = q.performWriteQueryIgnoringReturn(ctx, tx, "account archive", query, args); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "archiving account")
	}

	if err = q.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountArchiveEventEntry(userID, accountID, archivedByUser)); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing account archive audit log event entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	logger.Info("account archived")

	return nil
}

// GetAuditLogEntriesForAccount fetches a list of audit log entries from the database that relate to a given account.
func (q *SQLQuerier) GetAuditLogEntriesForAccount(ctx context.Context, accountID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == 0 {
		return nil, ErrInvalidIDProvided
	}

	tracing.AttachAccountIDToSpan(span, accountID)
	logger := q.logger.WithValue(keys.AccountIDKey, accountID)

	query, args := q.sqlQueryBuilder.BuildGetAuditLogEntriesForAccountQuery(ctx, accountID)

	rows, err := q.performReadQuery(ctx, q.db, "audit log entries for account", query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "querying database for audit log entries")
	}

	auditLogEntries, _, err := q.scanAuditLogEntries(ctx, rows, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning audit log entries")
	}

	return auditLogEntries, nil
}
