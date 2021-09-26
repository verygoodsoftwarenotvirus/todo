package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

const (
	// accountsUserMembershipTableName is what the accounts membership table calls itself.
	accountsUserMembershipTableName = "account_user_memberships"
)

var (
	_ types.AccountUserMembershipDataManager = (*SQLQuerier)(nil)

	// accountsUserMembershipTableColumns are the columns for the account user memberships table.
	accountsUserMembershipTableColumns = []string{
		"account_user_memberships.id",
		"account_user_memberships.belongs_to_user",
		"account_user_memberships.belongs_to_account",
		"account_user_memberships.account_roles",
		"account_user_memberships.default_account",
		"account_user_memberships.created_on",
		"account_user_memberships.last_updated_on",
		"account_user_memberships.archived_on",
	}
)

const (
	accountMemberRolesSeparator = commaSeparator
)

// scanAccountUserMembership takes a database Scanner (i.e. *sql.Row) and scans the result into an AccountUserMembership struct.
func (q *SQLQuerier) scanAccountUserMembership(ctx context.Context, scan database.Scanner) (x *types.AccountUserMembership, err error) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	x = &types.AccountUserMembership{
		AccountRoles: []string{},
	}

	var (
		rawAccountRoles string
	)

	targetVars := []interface{}{
		&x.ID,
		&x.BelongsToUser,
		&x.BelongsToAccount,
		&rawAccountRoles,
		&x.DefaultAccount,
		&x.CreatedOn,
		&x.LastUpdatedOn,
		&x.ArchivedOn,
	}

	if err = scan.Scan(targetVars...); err != nil {
		return nil, observability.PrepareError(err, q.logger, span, "scanning account user memberships")
	}

	if roles := strings.Split(rawAccountRoles, accountMemberRolesSeparator); len(roles) > 0 {
		x.AccountRoles = roles
	}

	return x, nil
}

// scanAccountUserMemberships takes some database rows and turns them into a slice of memberships.
func (q *SQLQuerier) scanAccountUserMemberships(ctx context.Context, rows database.ResultIterator) (defaultAccount string, accountRolesMap map[string][]string, err error) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	accountRolesMap = map[string][]string{}
	logger := q.logger

	for rows.Next() {
		x, scanErr := q.scanAccountUserMembership(ctx, rows)
		if scanErr != nil {
			return "", nil, scanErr
		}

		if x.DefaultAccount && defaultAccount == "" {
			defaultAccount = x.BelongsToAccount
		}

		accountRolesMap[x.BelongsToAccount] = x.AccountRoles
	}

	if err = q.checkRowsForErrorAndClose(ctx, rows); err != nil {
		return "", nil, observability.PrepareError(err, logger, span, "handling rows")
	}

	return defaultAccount, accountRolesMap, nil
}

const getAccountMembershipsForUserQuery = `
	SELECT
		account_user_memberships.id,
		account_user_memberships.belongs_to_user,
		account_user_memberships.belongs_to_account,
		account_user_memberships.account_roles,
		account_user_memberships.default_account,
		account_user_memberships.created_on,
		account_user_memberships.last_updated_on,
		account_user_memberships.archived_on
	FROM account_user_memberships
	JOIN accounts ON accounts.id = account_user_memberships.belongs_to_account
	WHERE account_user_memberships.archived_on IS NULL
	AND account_user_memberships.belongs_to_user = $1
`

// BuildSessionContextDataForUser queries the database for the memberships of a user and constructs a SessionContextData struct from the results.
func (q *SQLQuerier) BuildSessionContextDataForUser(ctx context.Context, userID string) (*types.SessionContextData, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, ErrInvalidIDProvided
	}

	logger := q.logger.WithValue(keys.UserIDKey, userID)
	tracing.AttachUserIDToSpan(span, userID)

	user, err := q.GetUser(ctx, userID)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "fetching user from database")
	}

	getAccountMembershipsArgs := []interface{}{userID}

	membershipRows, err := q.performReadQuery(ctx, q.db, "account memberships for user", getAccountMembershipsForUserQuery, getAccountMembershipsArgs)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "fetching user's memberships from database")
	}

	defaultAccountID, accountRolesMap, err := q.scanAccountUserMemberships(ctx, membershipRows)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning user's memberships from database")
	}

	actualAccountRolesMap := map[string]authorization.AccountRolePermissionsChecker{}
	for accountID, roles := range accountRolesMap {
		actualAccountRolesMap[accountID] = authorization.NewAccountRolePermissionChecker(roles...)
	}

	sessionCtxData := &types.SessionContextData{
		Requester: types.RequesterInfo{
			UserID:                user.ID,
			Reputation:            user.ServiceAccountStatus,
			ReputationExplanation: user.ReputationExplanation,
			ServicePermissions:    authorization.NewServiceRolePermissionChecker(user.ServiceRoles...),
		},
		AccountPermissions: actualAccountRolesMap,
		ActiveAccountID:    defaultAccountID,
	}

	return sessionCtxData, nil
}

const getDefaultAccountIDForUserQuery = `
	SELECT accounts.id
	FROM accounts 
	JOIN account_user_memberships ON account_user_memberships.belongs_to_account = accounts.id 
	WHERE account_user_memberships.belongs_to_user = $1
	AND account_user_memberships.default_account = $2
`

// GetDefaultAccountIDForUser retrieves the default account ID for a user.
func (q *SQLQuerier) GetDefaultAccountIDForUser(ctx context.Context, userID string) (string, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return "", ErrInvalidIDProvided
	}

	logger := q.logger.WithValue(keys.UserIDKey, userID)
	args := []interface{}{userID, true}

	var id string
	if err := q.getOneRow(ctx, q.db, "default account ID query", getDefaultAccountIDForUserQuery, args).Scan(&id); err != nil {
		return "", observability.PrepareError(err, logger, span, "executing default account ID query")
	}

	return id, nil
}

const markAccountAsUserDefaultQuery = `
	UPDATE account_user_memberships
	SET default_account = (belongs_to_user = $1 AND belongs_to_account = $2)
	WHERE archived_on IS NULL
	AND belongs_to_user = $3
`

// MarkAccountAsUserDefault does a thing.
func (q *SQLQuerier) MarkAccountAsUserDefault(ctx context.Context, userID, accountID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" || accountID == "" {
		return ErrInvalidIDProvided
	}

	logger := q.logger.WithValues(map[string]interface{}{
		keys.UserIDKey:    userID,
		keys.AccountIDKey: accountID,
	})

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)

	args := []interface{}{
		userID,
		accountID,
		userID,
	}

	// create the account.
	if err := q.performWriteQuery(ctx, q.db, "user default account assignment", markAccountAsUserDefaultQuery, args); err != nil {
		return observability.PrepareError(err, logger, span, "assigning user default account")
	}

	logger.Info("account marked as default")

	return nil
}

const userIsMemberOfAccountQuery = `
	SELECT EXISTS ( 
		SELECT account_user_memberships.id 
		FROM account_user_memberships 
		WHERE account_user_memberships.archived_on IS NULL 
		AND account_user_memberships.belongs_to_account = $1 
		AND account_user_memberships.belongs_to_user = $2 
	)
`

// UserIsMemberOfAccount does a thing.
func (q *SQLQuerier) UserIsMemberOfAccount(ctx context.Context, userID, accountID string) (bool, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" || accountID == "" {
		return false, ErrInvalidIDProvided
	}

	logger := q.logger.WithValues(map[string]interface{}{
		keys.UserIDKey:    userID,
		keys.AccountIDKey: accountID,
	})

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)

	args := []interface{}{
		accountID,
		userID,
	}

	result, err := q.performBooleanQuery(ctx, q.db, userIsMemberOfAccountQuery, args)
	if err != nil {
		return false, observability.PrepareError(err, logger, span, "performing user membership check query")
	}

	return result, nil
}

const modifyUserPermissionsQuery = `
	UPDATE account_user_memberships SET account_roles = $1 WHERE belongs_to_account = $2 AND belongs_to_user = $3
`

// ModifyUserPermissions does a thing.
func (q *SQLQuerier) ModifyUserPermissions(ctx context.Context, accountID, userID string, input *types.ModifyUserPermissionsInput) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" || userID == "" {
		return ErrInvalidIDProvided
	}

	if input == nil {
		return ErrNilInputProvided
	}

	logger := q.logger.WithValues(map[string]interface{}{
		keys.AccountIDKey: accountID,
		keys.UserIDKey:    userID,
		"new_roles":       input.NewRoles,
	})

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)

	args := []interface{}{
		strings.Join(input.NewRoles, accountMemberRolesSeparator),
		accountID,
		userID,
	}

	// modify the membership.
	if err := q.performWriteQuery(ctx, q.db, "user account permissions modification", modifyUserPermissionsQuery, args); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return observability.PrepareError(err, logger, span, "modifying user account permissions")
	}

	logger.Info("user permissions modified")

	return nil
}

const transferAccountOwnershipQuery = `
	UPDATE accounts SET belongs_to_user = $1 WHERE archived_on IS NULL AND belongs_to_user = $2 AND id = $3
`

const transferAccountMembershipQuery = `
	UPDATE account_user_memberships SET belongs_to_user = $1 WHERE archived_on IS NULL AND belongs_to_account = $2 AND belongs_to_user = $3
`

// TransferAccountOwnership does a thing.
func (q *SQLQuerier) TransferAccountOwnership(ctx context.Context, accountID string, input *types.AccountOwnershipTransferInput) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" {
		return ErrInvalidIDProvided
	}

	if input == nil {
		return ErrNilInputProvided
	}

	logger := q.logger.WithValues(map[string]interface{}{
		keys.AccountIDKey: accountID,
		"current_owner":   input.CurrentOwner,
		"new_owner":       input.NewOwner,
	})

	tracing.AttachUserIDToSpan(span, input.NewOwner)
	tracing.AttachAccountIDToSpan(span, accountID)

	// begin account transfer transaction
	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	transferAccountOwnershipArgs := []interface{}{
		input.NewOwner,
		input.CurrentOwner,
		accountID,
	}

	// create the membership.
	if err = q.performWriteQuery(ctx, tx, "user ownership transfer", transferAccountOwnershipQuery, transferAccountOwnershipArgs); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "transferring account to new owner")
	}

	transferAccountMembershipArgs := []interface{}{
		input.NewOwner,
		accountID,
		input.CurrentOwner,
	}

	// create the membership.
	if err = q.performWriteQuery(ctx, tx, "user memberships transfer", transferAccountMembershipQuery, transferAccountMembershipArgs); err != nil {
		q.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "transferring account memberships")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	logger.Info("account transferred to new owner")

	return nil
}

const addUserToAccountQuery = `
	INSERT INTO account_user_memberships (id,belongs_to_user,belongs_to_account,account_roles)
	VALUES ($1,$2,$3,$4)
`

// AddUserToAccount does a thing.
func (q *SQLQuerier) AddUserToAccount(ctx context.Context, input *types.AddUserToAccountInput) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return ErrNilInputProvided
	}

	logger := q.logger.WithValues(map[string]interface{}{
		keys.UserIDKey:    input.UserID,
		keys.AccountIDKey: input.AccountID,
	})

	tracing.AttachUserIDToSpan(span, input.UserID)
	tracing.AttachAccountIDToSpan(span, input.AccountID)

	addUserToAccountArgs := []interface{}{
		input.ID,
		input.UserID,
		input.AccountID,
		strings.Join(input.AccountRoles, accountMemberRolesSeparator),
	}

	// create the membership.
	if err := q.performWriteQuery(ctx, q.db, "user account membership creation", addUserToAccountQuery, addUserToAccountArgs); err != nil {
		return observability.PrepareError(err, logger, span, "creating user account membership")
	}

	logger.Info("user added to account")

	return nil
}

const removeUserFromAccountQuery = `
	DELETE FROM account_user_memberships 
	WHERE account_user_memberships.archived_on IS NULL
	AND account_user_memberships.belongs_to_account = $1 
	AND account_user_memberships.belongs_to_user = $2
`

// RemoveUserFromAccount removes a user's membership to an account.
func (q *SQLQuerier) RemoveUserFromAccount(ctx context.Context, userID, accountID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" || accountID == "" {
		return ErrInvalidIDProvided
	}

	logger := q.logger.WithValues(map[string]interface{}{
		keys.UserIDKey:    userID,
		keys.AccountIDKey: accountID,
	})

	tracing.AttachUserIDToSpan(span, userID)
	tracing.AttachAccountIDToSpan(span, accountID)

	args := []interface{}{
		accountID,
		userID,
	}

	// create the membership.
	if err := q.performWriteQuery(ctx, q.db, "user membership removal", removeUserFromAccountQuery, args); err != nil {
		return observability.PrepareError(err, logger, span, "removing user from account")
	}

	logger.Info("user removed from account")

	return nil
}
