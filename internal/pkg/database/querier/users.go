package querier

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.UserDataManager = (*Client)(nil)
)

// scanUser provides a consistent way to scan something like a *sql.Row into a User struct.
func (c *Client) scanUser(ctx context.Context, scan database.Scanner, includeCounts bool) (user *types.User, filteredCount, totalCount uint64, err error) {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue("include_counts", includeCounts)
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

	if err = scan.Scan(targetVars...); err != nil {
		return nil, 0, 0, observability.PrepareError(err, logger, span, "scanning user")
	}

	user.ServiceAdminPermissions = permissions.NewServiceAdminPermissions(perms)

	return user, filteredCount, totalCount, nil
}

// scanUsers takes database rows and loads them into a slice of User structs.
func (c *Client) scanUsers(ctx context.Context, rows database.ResultIterator, includeCounts bool) (users []*types.User, filteredCount, totalCount uint64, err error) {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue("include_counts", includeCounts)

	for rows.Next() {
		user, fc, tc, scanErr := c.scanUser(ctx, rows, includeCounts)
		if scanErr != nil {
			return nil, 0, 0, observability.PrepareError(scanErr, logger, span, "scanning user result")
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

	if err = c.checkRowsForErrorAndClose(ctx, rows); err != nil {
		return nil, 0, 0, observability.PrepareError(err, logger, span, "handling rows")
	}

	return users, filteredCount, totalCount, nil
}

// getUser fetches a user.
func (c *Client) getUser(ctx context.Context, userID uint64, withVerifiedTOTPSecret bool) (*types.User, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue(keys.UserIDKey, userID)

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

	u, _, _, err := c.scanUser(ctx, row, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning user")
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

	u, _, _, err := c.scanUser(ctx, row, false)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, observability.PrepareError(err, logger, span, "scanning user")
	}

	return u, nil
}

// SearchForUsersByUsername fetches a list of users whose usernames begin with a given query.
func (c *Client) SearchForUsersByUsername(ctx context.Context, usernameQuery string) ([]*types.User, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue(keys.SearchQueryKey, usernameQuery)

	query, args := c.sqlQueryBuilder.BuildSearchForUserByUsernameQuery(usernameQuery)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, observability.PrepareError(err, logger, span, "querying database for users")
	}

	u, _, _, err := c.scanUsers(ctx, rows, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning user")
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

	tracing.AttachQueryFilterToSpan(span, filter)
	logger := filter.AttachToLogger(c.logger)

	if filter != nil {
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	query, args := c.sqlQueryBuilder.BuildGetUsersQuery(filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning user")
	}

	if x.Users, x.FilteredCount, x.TotalCount, err = c.scanUsers(ctx, rows, true); err != nil {
		return nil, observability.PrepareError(err, logger, span, "loading response from database")
	}

	return x, nil
}

// createUser creates a user. The `user` and `account` parameters are meant to be filled out.
func (c *Client) createUser(ctx context.Context, user *types.User, account *types.Account, userCreationQuery string, userCreationArgs []interface{}) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue("username", user.Username)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	userID, err := c.performWriteQuery(ctx, tx, false, "user creation", userCreationQuery, userCreationArgs)
	if err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "creating user")
	}

	user.ID = userID
	account.BelongsToUser = user.ID

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserCreationEventEntry(user.ID)); err != nil {
		logger.Error(err, "writing <> audit log entry")
		c.rollbackTransaction(ctx, tx)

		return fmt.Errorf("writing <> audit log entry: %w", err)
	}

	// create the account.
	accountCreationInput := types.NewAccountCreationInputForUser(user)
	accountCreationInput.DefaultUserPermissions = account.DefaultUserPermissions
	accountCreationQuery, accountCreationArgs := c.sqlQueryBuilder.BuildAccountCreationQuery(accountCreationInput)

	accountID, err := c.performWriteQuery(ctx, tx, false, "account creation", accountCreationQuery, accountCreationArgs)
	if err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "create account")
	}

	account.ID = accountID

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountCreationEventEntry(account, user.ID)); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing account creation audit log entry")
	}

	addUserToAccountQuery, addUserToAccountArgs := c.sqlQueryBuilder.BuildCreateMembershipForNewUserQuery(userID, accountID)
	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "account user membership creation", addUserToAccountQuery, addUserToAccountArgs); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing account user membership creation audit log entry")
	}

	addToAccountInput := &types.AddUserToAccountInput{
		UserID:                 user.ID,
		UserAccountPermissions: account.DefaultUserPermissions,
		Reason:                 "account creation",
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserAddedToAccountEventEntry(userID, account.ID, addToAccountInput)); err != nil {
		logger.Error(err, "writing <> audit log entry")
		c.rollbackTransaction(ctx, tx)

		return fmt.Errorf("writing <> audit log entry: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	return nil
}

// CreateUser creates a user.
func (c *Client) CreateUser(ctx context.Context, input types.UserDataStoreCreationInput) (*types.User, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUsernameToSpan(span, input.Username)
	logger := c.logger.WithValue(keys.UsernameKey, input.Username)

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
		return nil, observability.PrepareError(err, logger, span, "creating user")
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

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "user update", query, args); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "updating user")
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserUpdateEventEntry(updated.ID, nil)); err != nil {
		logger.Error(err, "writing <> audit log entry")
		c.rollbackTransaction(ctx, tx)

		return fmt.Errorf("writing <> audit log entry: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
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

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "user authentication update", query, args); err != nil {
		c.rollbackTransaction(ctx, tx)
		return fmt.Errorf("updating user's password: %w", err)
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserUpdatePasswordEventEntry(userID)); err != nil {
		logger.Error(err, "writing <> audit log entry")
		c.rollbackTransaction(ctx, tx)

		return fmt.Errorf("writing <> audit log entry: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
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

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "user two factor secret update", query, args); err != nil {
		c.rollbackTransaction(ctx, tx)
		return fmt.Errorf("updating user's two factor secret: %w", err)
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserUpdateTwoFactorSecretEventEntry(userID)); err != nil {
		logger.Error(err, "writing <> audit log entry")
		c.rollbackTransaction(ctx, tx)

		return fmt.Errorf("writing <> audit log entry: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
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

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "user two factor secret verification", query, args); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "writing verified two factor status to database")
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserVerifyTwoFactorSecretEventEntry(userID)); err != nil {
		logger.Error(err, "writing <> audit log entry")
		c.rollbackTransaction(ctx, tx)

		return fmt.Errorf("writing <> audit log entry: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
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

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareError(err, logger, span, "beginning transaction")
	}

	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "user archive", archiveUserQuery, archiveUserArgs); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "archiving user")
	}

	if err = c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserArchiveEventEntry(userID)); err != nil {
		logger.Error(err, "writing <> audit log entry")
		c.rollbackTransaction(ctx, tx)

		return fmt.Errorf("writing <> audit log entry: %w", err)
	}

	archiveMembershipsQuery, archiveMembershipsArgs := c.sqlQueryBuilder.BuildArchiveAccountMembershipsForUserQuery(userID)

	if err = c.performWriteQueryIgnoringReturn(ctx, tx, "user memberships archive", archiveMembershipsQuery, archiveMembershipsArgs); err != nil {
		c.rollbackTransaction(ctx, tx)
		return observability.PrepareError(err, logger, span, "archiving user account memberships")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareError(err, logger, span, "committing transaction")
	}

	return nil
}

// GetAuditLogEntriesForUser fetches a list of audit log entries from the database that relate to a given user.
func (c *Client) GetAuditLogEntriesForUser(ctx context.Context, userID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue(keys.UserIDKey, userID)

	query, args := c.sqlQueryBuilder.BuildGetAuditLogEntriesForUserQuery(userID)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "querying database for audit log entries")
	}

	auditLogEntries, _, err := c.scanAuditLogEntries(ctx, rows, false)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "scanning response from database")
	}

	return auditLogEntries, nil
}
