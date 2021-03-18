package querier

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.UserDataManager = (*Client)(nil)
)

// scanUser provides a consistent way to scan something like a *sql.Row into a User struct.
func (c *Client) scanUser(scan database.Scanner, includeCounts bool) (user *types.User, filteredCount, totalCount uint64, err error) {
	user = &types.User{}

	var perms uint32

	targetVars := []interface{}{
		&user.ID,
		&user.ExternalID,
		&user.Username,
		&user.AvatarSrc,
		&user.HashedPassword,
		&user.Salt,
		&user.RequiresPasswordChange,
		&user.PasswordLastChangedOn,
		&user.TwoFactorSecret,
		&user.TwoFactorSecretVerifiedOn,
		&perms,
		&user.Reputation,
		&user.ReputationExplanation,
		&user.CreatedOn,
		&user.LastUpdatedOn,
		&user.ArchivedOn,
	}

	if includeCounts {
		targetVars = append(targetVars, &filteredCount, &totalCount)
	}

	if scanErr := scan.Scan(targetVars...); scanErr != nil {
		return nil, 0, 0, scanErr
	}

	user.ServiceAdminPermissions = permissions.NewServiceAdminPermissions(perms)

	return user, filteredCount, totalCount, nil
}

// scanUsers takes database rows and loads them into a slice of User structs.
func (c *Client) scanUsers(rows database.ResultIterator, includeCounts bool) (users []*types.User, filteredCount, totalCount uint64, err error) {
	for rows.Next() {
		user, fc, tc, scanErr := c.scanUser(rows, includeCounts)
		if scanErr != nil {
			return nil, 0, 0, fmt.Errorf("scanning user result: %w", scanErr)
		}

		if includeCounts {
			if filteredCount == 0 {
				filteredCount = fc
			}

			if totalCount == 0 {
				totalCount = tc
			}
		}

		users = append(users, user)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, 0, 0, rowsErr
	}

	if closeErr := rows.Close(); closeErr != nil {
		c.logger.Error(closeErr, "closing rows")
		return nil, 0, 0, closeErr
	}

	return users, filteredCount, totalCount, nil
}

// getUser fetches a user.
func (c *Client) getUser(ctx context.Context, userID uint64, withVerifiedTOTPSecret bool) (*types.User, error) {
	logger := c.logger.WithValue(keys.UserIDKey, userID)

	logger.Debug("GetUser called")

	var (
		query string
		args  []interface{}
	)

	if withVerifiedTOTPSecret {
		query, args = c.sqlQueryBuilder.BuildGetUserQuery(userID)
	} else {
		query, args = c.sqlQueryBuilder.BuildGetUserWithUnverifiedTwoFactorSecretQuery(userID)
	}

	row := c.db.QueryRowContext(ctx, query, args...)

	u, _, _, err := c.scanUser(row, false)
	if err != nil {
		return nil, fmt.Errorf("scanning user: %w", err)
	}

	return u, nil
}

// UserHasStatus fetches whether or not an item exists from the database.
func (c *Client) UserHasStatus(ctx context.Context, userID uint64, statuses ...string) (banned bool, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	c.logger.WithValues(map[string]interface{}{
		keys.UserIDKey: userID,
	}).Debug("UserIsBanned called")

	query, args := c.sqlQueryBuilder.BuildUserHasStatusQuery(userID, statuses...)

	return c.performBooleanQuery(ctx, c.db, query, args)
}

// GetUser fetches a user.
func (c *Client) GetUser(ctx context.Context, userID uint64) (*types.User, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	return c.getUser(ctx, userID, true)
}

// GetUserWithUnverifiedTwoFactorSecret fetches a user with an unverified 2FA secret.
func (c *Client) GetUserWithUnverifiedTwoFactorSecret(ctx context.Context, userID uint64) (*types.User, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)

	return c.getUser(ctx, userID, false)
}

// GetUserByUsername fetches a user by their username.
func (c *Client) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUsernameToSpan(span, username)
	logger := c.logger.WithValue(keys.UsernameKey, username)

	logger.Debug("GetUserByUsername called")

	query, args := c.sqlQueryBuilder.BuildGetUserByUsernameQuery(username)
	row := c.db.QueryRowContext(ctx, query, args...)

	u, _, _, err := c.scanUser(row, false)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, fmt.Errorf("scanning user: %w", err)
	}

	return u, nil
}

// SearchForUsersByUsername fetches a list of users whose usernames begin with a given query.
func (c *Client) SearchForUsersByUsername(ctx context.Context, usernameQuery string) ([]*types.User, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("SearchForUsersByUsername called")

	query, args := c.sqlQueryBuilder.BuildSearchForUserByUsernameQuery(usernameQuery)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, fmt.Errorf("querying database for users: %w", err)
	}

	u, _, _, err := c.scanUsers(rows, false)
	if err != nil {
		return nil, fmt.Errorf("scanning user: %w", err)
	}

	return u, nil
}

// GetAllUsersCount fetches a count of users from the database that meet a particular filter.
func (c *Client) GetAllUsersCount(ctx context.Context) (count uint64, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllUsersCount called")

	return c.performCountQuery(ctx, c.db, c.sqlQueryBuilder.BuildGetAllUsersCountQuery())
}

// GetUsers fetches a list of users from the database that meet a particular filter.
func (c *Client) GetUsers(ctx context.Context, filter *types.QueryFilter) (x *types.UserList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	x = &types.UserList{}

	c.logger.WithValue(keys.FilterIsNilKey, filter == nil).Debug("GetUsers called")

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit)
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := c.sqlQueryBuilder.BuildGetUsersQuery(filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("scanning user: %w", err)
	}

	if x.Users, x.FilteredCount, x.TotalCount, err = c.scanUsers(rows, true); err != nil {
		return nil, fmt.Errorf("loading response from database: %w", err)
	}

	return x, nil
}

// createUser creates a user. The `user` and `account` parameters are meant to be filled out.
func (c *Client) createUser(ctx context.Context, user *types.User, account *types.Account, userCreationQuery string, userCreationArgs []interface{}) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue("username", user.Username)

	tx, transactionStartErr := c.db.BeginTx(ctx, nil)
	if transactionStartErr != nil {
		return fmt.Errorf("beginning transaction: %w", transactionStartErr)
	}

	userID, userCreateErr := c.performWriteQuery(ctx, tx, false, "user creation", userCreationQuery, userCreationArgs)
	if userCreateErr != nil {
		c.rollbackTransaction(ctx, tx)
		return userCreateErr
	}

	user.ID = userID
	account.BelongsToUser = user.ID

	if auditLogEntryWriteErr := c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserCreationEventEntry(user.ID)); auditLogEntryWriteErr != nil {
		logger.Error(auditLogEntryWriteErr, "writing <> audit log entry")
		c.rollbackTransaction(ctx, tx)

		return fmt.Errorf("writing <> audit log entry: %w", auditLogEntryWriteErr)
	}

	// create the account.
	accountCreationInput := types.NewAccountCreationInputForUser(user)
	accountCreationInput.DefaultUserPermissions = account.DefaultUserPermissions
	accountCreationQuery, accountCreationArgs := c.sqlQueryBuilder.BuildAccountCreationQuery(accountCreationInput)

	accountID, accountCreateErr := c.performWriteQuery(ctx, tx, false, "account creation", accountCreationQuery, accountCreationArgs)
	if accountCreateErr != nil {
		c.rollbackTransaction(ctx, tx)
		return accountCreateErr
	}

	account.ID = accountID

	if auditLogEntryWriteErr := c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountCreationEventEntry(account, user.ID)); auditLogEntryWriteErr != nil {
		logger.Error(auditLogEntryWriteErr, "writing <> audit log entry")
		c.rollbackTransaction(ctx, tx)

		return fmt.Errorf("writing <> audit log entry: %w", auditLogEntryWriteErr)
	}

	addUserToAccountQuery, addUserToAccountArgs := c.sqlQueryBuilder.BuildCreateMembershipForNewUserQuery(userID, accountID)
	if accountMembershipErr := c.performWriteQueryIgnoringReturn(ctx, tx, "account user membership creation", addUserToAccountQuery, addUserToAccountArgs); accountMembershipErr != nil {
		c.rollbackTransaction(ctx, tx)
		return accountMembershipErr
	}

	addToAccountInput := &types.AddUserToAccountInput{
		UserID:                 user.ID,
		UserAccountPermissions: account.DefaultUserPermissions,
		Reason:                 "account creation",
	}

	if auditLogEntryWriteErr := c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserAddedToAccountEventEntry(userID, account.ID, addToAccountInput)); auditLogEntryWriteErr != nil {
		logger.Error(auditLogEntryWriteErr, "writing <> audit log entry")
		c.rollbackTransaction(ctx, tx)

		return fmt.Errorf("writing <> audit log entry: %w", auditLogEntryWriteErr)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("committing transaction: %w", commitErr)
	}

	return nil
}

// CreateUser creates a user.
func (c *Client) CreateUser(ctx context.Context, input types.UserDataStoreCreationInput) (*types.User, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUsernameToSpan(span, input.Username)
	c.logger.WithValue(keys.UsernameKey, input.Username).Debug("CreateUser called")

	// create the user.
	userCreationQuery, userCreationArgs := c.sqlQueryBuilder.BuildCreateUserQuery(input)

	user := &types.User{
		Username:        input.Username,
		HashedPassword:  input.HashedPassword,
		TwoFactorSecret: input.TwoFactorSecret,
		CreatedOn:       c.currentTime(),
	}

	account := &types.Account{
		Name:                   input.Username,
		PlanID:                 nil,
		CreatedOn:              c.currentTime(),
		DefaultUserPermissions: permissions.ServiceUserPermissions(math.MaxUint32),
	}

	if err := c.createUser(ctx, user, account, userCreationQuery, userCreationArgs); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser receives a complete User struct and updates its record in the database.
// NOTE: this function uses the ID provided in the input to make its query.
func (c *Client) UpdateUser(ctx context.Context, updated *types.User, changes []types.FieldChangeSummary) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUsernameToSpan(span, updated.Username)
	logger := c.logger.WithValue(keys.UsernameKey, updated.Username)

	logger.Debug("UpdateUser called")

	query, args := c.sqlQueryBuilder.BuildUpdateUserQuery(updated)

	tx, transactionStartErr := c.db.BeginTx(ctx, nil)
	if transactionStartErr != nil {
		return fmt.Errorf("beginning transaction: %w", transactionStartErr)
	}

	if execErr := c.performWriteQueryIgnoringReturn(ctx, tx, "user update", query, args); execErr != nil {
		c.rollbackTransaction(ctx, tx)
		return fmt.Errorf("updating user: %w", execErr)
	}

	if auditLogEntryWriteErr := c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserUpdateEventEntry(updated.ID, nil)); auditLogEntryWriteErr != nil {
		logger.Error(auditLogEntryWriteErr, "writing <> audit log entry")
		c.rollbackTransaction(ctx, tx)

		return fmt.Errorf("writing <> audit log entry: %w", auditLogEntryWriteErr)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("committing transaction: %w", commitErr)
	}

	return nil
}

// UpdateUserPassword updates a user's authentication hash in the database.
func (c *Client) UpdateUserPassword(ctx context.Context, userID uint64, newHash string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	logger := c.logger.WithValue(keys.UserIDKey, userID)

	logger.Debug("UpdateUserPassword called")

	query, args := c.sqlQueryBuilder.BuildUpdateUserPasswordQuery(userID, newHash)

	tx, transactionStartErr := c.db.BeginTx(ctx, nil)
	if transactionStartErr != nil {
		return fmt.Errorf("beginning transaction: %w", transactionStartErr)
	}

	if execErr := c.performWriteQueryIgnoringReturn(ctx, tx, "user authentication update", query, args); execErr != nil {
		c.rollbackTransaction(ctx, tx)
		return fmt.Errorf("updating user's password: %w", execErr)
	}

	if auditLogEntryWriteErr := c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserUpdatePasswordEventEntry(userID)); auditLogEntryWriteErr != nil {
		logger.Error(auditLogEntryWriteErr, "writing <> audit log entry")
		c.rollbackTransaction(ctx, tx)

		return fmt.Errorf("writing <> audit log entry: %w", auditLogEntryWriteErr)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("committing transaction: %w", commitErr)
	}

	return nil
}

// UpdateUserTwoFactorSecret marks a user's two factor secret as validated.
func (c *Client) UpdateUserTwoFactorSecret(ctx context.Context, userID uint64, newSecret string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	logger := c.logger.WithValue(keys.UserIDKey, userID)
	logger.Debug("UpdateUserTwoFactorSecret called")

	query, args := c.sqlQueryBuilder.BuildUpdateUserTwoFactorSecretQuery(userID, newSecret)

	tx, transactionStartErr := c.db.BeginTx(ctx, nil)
	if transactionStartErr != nil {
		return fmt.Errorf("beginning transaction: %w", transactionStartErr)
	}

	if execErr := c.performWriteQueryIgnoringReturn(ctx, tx, "user two factor secret update", query, args); execErr != nil {
		c.rollbackTransaction(ctx, tx)
		return fmt.Errorf("updating user's two factor secret: %w", execErr)
	}

	if auditLogEntryWriteErr := c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserUpdateTwoFactorSecretEventEntry(userID)); auditLogEntryWriteErr != nil {
		logger.Error(auditLogEntryWriteErr, "writing <> audit log entry")
		c.rollbackTransaction(ctx, tx)

		return fmt.Errorf("writing <> audit log entry: %w", auditLogEntryWriteErr)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("committing transaction: %w", commitErr)
	}

	return nil
}

// VerifyUserTwoFactorSecret marks a user's two factor secret as validated.
func (c *Client) VerifyUserTwoFactorSecret(ctx context.Context, userID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	logger := c.logger.WithValue(keys.UserIDKey, userID)

	logger.Debug("VerifyUserTwoFactorSecret called")

	query, args := c.sqlQueryBuilder.BuildVerifyUserTwoFactorSecretQuery(userID)

	tx, transactionStartErr := c.db.BeginTx(ctx, nil)
	if transactionStartErr != nil {
		return fmt.Errorf("beginning transaction: %w", transactionStartErr)
	}

	if execErr := c.performWriteQueryIgnoringReturn(ctx, tx, "user two factor secret verification", query, args); execErr != nil {
		c.rollbackTransaction(ctx, tx)
		return fmt.Errorf("writing verified two factor status to database: %w", execErr)
	}

	if auditLogEntryWriteErr := c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserVerifyTwoFactorSecretEventEntry(userID)); auditLogEntryWriteErr != nil {
		logger.Error(auditLogEntryWriteErr, "writing <> audit log entry")
		c.rollbackTransaction(ctx, tx)

		return fmt.Errorf("writing <> audit log entry: %w", auditLogEntryWriteErr)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("committing transaction: %w", commitErr)
	}

	return nil
}

// ArchiveUser archives a user.
func (c *Client) ArchiveUser(ctx context.Context, userID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	logger := c.logger.WithValue(keys.UserIDKey, userID)

	logger.Debug("ArchiveUser called")

	archiveUserQuery, archiveUserArgs := c.sqlQueryBuilder.BuildArchiveUserQuery(userID)

	tx, transactionStartErr := c.db.BeginTx(ctx, nil)
	if transactionStartErr != nil {
		return fmt.Errorf("beginning transaction: %w", transactionStartErr)
	}

	if execErr := c.performWriteQueryIgnoringReturn(ctx, tx, "user archive", archiveUserQuery, archiveUserArgs); execErr != nil {
		c.rollbackTransaction(ctx, tx)
		return fmt.Errorf("archiving user: %w", execErr)
	}

	if auditLogEntryWriteErr := c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserArchiveEventEntry(userID)); auditLogEntryWriteErr != nil {
		logger.Error(auditLogEntryWriteErr, "writing <> audit log entry")
		c.rollbackTransaction(ctx, tx)

		return fmt.Errorf("writing <> audit log entry: %w", auditLogEntryWriteErr)
	}

	archiveMembershipsQuery, archiveMembershipsArgs := c.sqlQueryBuilder.BuildArchiveAccountMembershipsForUserQuery(userID)

	if execErr := c.performWriteQueryIgnoringReturn(ctx, tx, "user memberships archive", archiveMembershipsQuery, archiveMembershipsArgs); execErr != nil {
		c.rollbackTransaction(ctx, tx)
		return fmt.Errorf("archiving user account memberships: %w", execErr)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("committing transaction: %w", commitErr)
	}

	return nil
}

// GetAuditLogEntriesForUser fetches a list of audit log entries from the database that relate to a given user.
func (c *Client) GetAuditLogEntriesForUser(ctx context.Context, userID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("GetAuditLogEntriesForUser called")

	query, args := c.sqlQueryBuilder.BuildGetAuditLogEntriesForUserQuery(userID)

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
