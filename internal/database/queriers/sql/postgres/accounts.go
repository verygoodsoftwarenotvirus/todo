package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/Masterminds/squirrel"
	"github.com/segmentio/ksuid"
)

const (
	// accountsTableName is what the accounts table calls itself.
	accountsTableName = "accounts"

	accountUserMembershipsOnAccountsJoinClause = "account_user_memberships ON account_user_memberships.belongs_to_account = accounts.id"
)

var (
	_ types.AccountDataManager = (*SQLQuerier)(nil)

	accountsTableColumns = []string{
		"accounts.id",
		"accounts.name",
		"accounts.billing_status",
		"accounts.contact_email",
		"accounts.contact_phone",
		"accounts.payment_processor_customer_id",
		"accounts.subscription_plan_id",
		"accounts.created_on",
		"accounts.last_updated_on",
		"accounts.archived_on",
		"accounts.belongs_to_user",
	}
)

// scanAccount takes a database Scanner (i.e. *sql.Row) and scans the result into an Account struct.
func (q *SQLQuerier) scanAccount(ctx context.Context, scan database.Scanner, includeCounts bool) (account *types.Account, membership *types.AccountUserMembership, filteredCount, totalCount uint64, err error) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("include_counts", includeCounts)

	account = &types.Account{Members: []*types.AccountUserMembership{}}
	membership = &types.AccountUserMembership{}

	var (
		rawRoles string
	)

	targetVars := []interface{}{
		&account.ID,
		&account.Name,
		&account.BillingStatus,
		&account.ContactEmail,
		&account.ContactPhone,
		&account.PaymentProcessorCustomerID,
		&account.SubscriptionPlanID,
		&account.CreatedOn,
		&account.LastUpdatedOn,
		&account.ArchivedOn,
		&account.BelongsToUser,
		&membership.ID,
		&membership.BelongsToUser,
		&membership.BelongsToAccount,
		&rawRoles,
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

	membership.AccountRoles = strings.Split(rawRoles, accountMemberRolesSeparator)

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

const getAccountQuery = `
	SELECT 
		accounts.id, 
		accounts.name, 
		accounts.billing_status, 
		accounts.contact_email, 
		accounts.contact_phone, 
		accounts.payment_processor_customer_id, 
		accounts.subscription_plan_id, 
		accounts.created_on, 
		accounts.last_updated_on, 
		accounts.archived_on, 
		accounts.belongs_to_user, 
		account_user_memberships.id, 
		account_user_memberships.belongs_to_user, 
		account_user_memberships.belongs_to_account, 
		account_user_memberships.account_roles, 
		account_user_memberships.default_account, 
		account_user_memberships.created_on, 
		account_user_memberships.last_updated_on,
		account_user_memberships.archived_on 
	FROM accounts
	JOIN account_user_memberships ON account_user_memberships.belongs_to_account = accounts.id 
	WHERE accounts.archived_on IS NULL 
	AND accounts.belongs_to_user = $1
	AND accounts.id = $2
`

// GetAccount fetches an account from the database.
func (q *SQLQuerier) GetAccount(ctx context.Context, accountID, userID string) (*types.Account, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" || userID == "" {
		return nil, ErrInvalidIDProvided
	}

	tracing.AttachAccountIDToSpan(span, accountID)
	tracing.AttachUserIDToSpan(span, userID)

	logger := q.logger.WithValues(map[string]interface{}{
		keys.AccountIDKey: accountID,
		keys.UserIDKey:    userID,
	})

	args := []interface{}{
		userID,
		accountID,
	}

	rows, err := q.performReadQuery(ctx, q.db, "account", getAccountQuery, args)
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

const getAllAccountsCountQuery = `
	SELECT COUNT(accounts.id) FROM accounts WHERE accounts.archived_on IS NULL
`

// GetAllAccountsCount fetches the count of accounts from the database that meet a particular filter.
func (q *SQLQuerier) GetAllAccountsCount(ctx context.Context) (uint64, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger

	count, err := q.performCountQuery(ctx, q.db, getAllAccountsCountQuery, "fetching count of all accounts")
	if err != nil {
		return 0, observability.PrepareError(err, logger, span, "querying for count of accounts")
	}

	return count, nil
}

var (
	accountAndMembershipColumns = append(accountsTableColumns, accountsUserMembershipTableColumns...)
)

// buildGetAccountsQuery builds a SQL query selecting accounts that adhere to a given QueryFilter and belong to a given account,
// and returns both the query and the relevant args to pass to the query executor.
func (q *SQLQuerier) buildGetAccountsQuery(ctx context.Context, userID string, forAdmin bool, filter *types.QueryFilter) (query string, args []interface{}) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))
	}

	var includeArchived bool
	if filter != nil {
		includeArchived = filter.IncludeArchived
	}

	filteredCountQuery, filteredCountQueryArgs := q.buildFilteredCountQuery(ctx, accountsTableName, nil, nil, userOwnershipColumn, userID, forAdmin, includeArchived, filter)
	totalCountQuery, totalCountQueryArgs := q.buildTotalCountQuery(ctx, accountsTableName, nil, nil, userOwnershipColumn, userID, forAdmin, includeArchived)

	builder := q.sqlBuilder.Select(append(
		accountAndMembershipColumns,
		fmt.Sprintf("(%s) as total_count", totalCountQuery),
		fmt.Sprintf("(%s) as filtered_count", filteredCountQuery),
	)...).
		From(accountsTableName).
		Join(accountUserMembershipsOnAccountsJoinClause)

	if !forAdmin {
		where := squirrel.Eq{
			"accounts.archived_on": nil,
		}

		if userID != "" {
			where["accounts.belongs_to_user"] = userID
		}

		builder = builder.Where(where)
	}

	builder = builder.GroupBy(fmt.Sprintf(
		"%s.%s, %s.%s",
		accountsTableName,
		"id",
		accountsUserMembershipTableName,
		"id",
	))

	if filter != nil {
		builder = applyFilterToQueryBuilder(filter, accountsTableName, builder)
	}

	query, selectArgs := q.buildQuery(span, builder)

	return query, append(append(filteredCountQueryArgs, totalCountQueryArgs...), selectArgs...)
}

// GetAccounts fetches a list of accounts from the database that meet a particular filter.
func (q *SQLQuerier) GetAccounts(ctx context.Context, userID string, filter *types.QueryFilter) (x *types.AccountList, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, ErrInvalidIDProvided
	}

	logger := filter.AttachToLogger(q.logger).WithValue(keys.UserIDKey, userID)
	tracing.AttachQueryFilterToSpan(span, filter)
	tracing.AttachUserIDToSpan(span, userID)

	x = &types.AccountList{}
	if filter != nil {
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := q.buildGetAccountsQuery(ctx, userID, false, filter)

	rows, err := q.performReadQuery(ctx, q.db, "accounts", query, args)
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

	query, args := q.buildGetAccountsQuery(ctx, "", true, filter)

	rows, err := q.performReadQuery(ctx, q.db, "accounts for admin", query, args)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "querying database for accounts")
	}

	if x.Accounts, x.FilteredCount, x.TotalCount, err = q.scanAccounts(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning accounts")
	}

	return x, nil
}

const accountCreationQuery = `
	INSERT INTO accounts (id,name,billing_status,contact_email,contact_phone,belongs_to_user) VALUES ($1,$2,$3,$4,$5,$6)
`

const addUserToAccountDuringCreationQuery = `
	INSERT INTO account_user_memberships (id,belongs_to_user,belongs_to_account,account_roles)
	VALUES ($1,$2,$3,$4)
`

// CreateAccount creates an account in the database.
func (q *SQLQuerier) CreateAccount(ctx context.Context, input *types.AccountCreationInput) (*types.Account, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	logger := q.logger.WithValue(keys.UserIDKey, input.BelongsToUser)

	// begin account creation transaction
	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "beginning transaction")
	}

	accountCreationArgs := []interface{}{
		input.ID,
		input.Name,
		types.UnpaidAccountBillingStatus,
		input.ContactEmail,
		input.ContactPhone,
		input.BelongsToUser,
	}

	// create the account.
	if writeErr := q.performWriteQuery(ctx, tx, "account creation", accountCreationQuery, accountCreationArgs); writeErr != nil {
		q.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(writeErr, logger, span, "creating account")
	}

	account := &types.Account{
		ID:            input.ID,
		Name:          input.Name,
		BelongsToUser: input.BelongsToUser,
		BillingStatus: types.UnpaidAccountBillingStatus,
		ContactEmail:  input.ContactEmail,
		ContactPhone:  input.ContactPhone,
		CreatedOn:     q.currentTime(),
	}

	addInput := &types.AddUserToAccountInput{
		ID:           ksuid.New().String(),
		UserID:       input.BelongsToUser,
		AccountID:    account.ID,
		AccountRoles: []string{authorization.AccountAdminRole.String()},
	}

	addUserToAccountArgs := []interface{}{
		addInput.ID,
		addInput.UserID,
		addInput.AccountID,
		strings.Join(addInput.AccountRoles, accountMemberRolesSeparator),
	}

	if err = q.performWriteQuery(ctx, tx, "account user membership creation", addUserToAccountDuringCreationQuery, addUserToAccountArgs); err != nil {
		q.rollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, logger, span, "creating account membership")
	}

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareError(err, logger, span, "committing transaction")
	}

	tracing.AttachAccountIDToSpan(span, account.ID)
	logger.Info("account created")

	return account, nil
}

const updateAccountQuery = `
	UPDATE accounts SET name = $1, contact_email = $2, contact_phone = $3, last_updated_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND belongs_to_user = $4 AND id = $5
`

// UpdateAccount updates a particular account. Note that UpdateAccount expects the provided input to have a valid ID.
func (q *SQLQuerier) UpdateAccount(ctx context.Context, updated *types.Account) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return ErrNilInputProvided
	}

	logger := q.logger.WithValue(keys.AccountIDKey, updated.ID)
	tracing.AttachAccountIDToSpan(span, updated.ID)

	args := []interface{}{
		updated.Name,
		updated.ContactEmail,
		updated.ContactPhone,
		updated.BelongsToUser,
		updated.ID,
	}

	if err := q.performWriteQuery(ctx, q.db, "account update", updateAccountQuery, args); err != nil {
		return observability.PrepareError(err, logger, span, "updating account")
	}

	logger.Info("account updated")

	return nil
}

const archiveAccountQuery = `
	UPDATE accounts SET last_updated_on = extract(epoch FROM NOW()), archived_on = extract(epoch FROM NOW()) WHERE archived_on IS NULL AND belongs_to_user = $1 AND id = $2
`

// ArchiveAccount archives an account from the database by its ID.
func (q *SQLQuerier) ArchiveAccount(ctx context.Context, accountID, userID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" || userID == "" {
		return ErrInvalidIDProvided
	}

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)

	logger := q.logger.WithValues(map[string]interface{}{
		keys.AccountIDKey: accountID,
		keys.UserIDKey:    userID,
	})

	args := []interface{}{
		userID,
		accountID,
	}

	if err := q.performWriteQuery(ctx, q.db, "account archive", archiveAccountQuery, args); err != nil {
		return observability.PrepareError(err, logger, span, "archiving account")
	}

	logger.Info("account archived")

	return nil
}
