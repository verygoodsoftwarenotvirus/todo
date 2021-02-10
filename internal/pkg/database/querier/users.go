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
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions/bitmask"
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
		&user.IsSiteAdmin,
		&perms,
		&user.AccountStatus,
		&user.AccountStatusExplanation,
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

	user.SiteAdminPermissions = bitmask.NewSiteAdminPermissions(perms)

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

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	if user.ID, err = c.performWriteQuery(ctx, tx, false, "user creation", userCreationQuery, userCreationArgs); err != nil {
		c.rollbackTransaction(tx)
		return err
	}

	c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserCreationEventEntry(user.ID))

	account.BelongsToUser = user.ID

	// create the account.
	accountCreationInput := types.NewAccountCreationInputForUser(user)
	accountCreationQuery, accountCreationArgs := c.sqlQueryBuilder.BuildCreateAccountQuery(accountCreationInput)

	if account.ID, err = c.performWriteQuery(ctx, tx, false, "account creation", accountCreationQuery, accountCreationArgs); err != nil {
		c.rollbackTransaction(tx)
		return err
	}

	c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildAccountCreationEventEntry(account))

	// LATER: create account user membership in transaction here

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
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
		Name:            input.Username,
		PlanID:          nil,
		PersonalAccount: true,
		CreatedOn:       c.currentTime(),
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
	c.logger.WithValue(keys.UsernameKey, updated.Username).Debug("UpdateUser called")

	query, args := c.sqlQueryBuilder.BuildUpdateUserQuery(updated)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	if execErr := c.performWriteQueryIgnoringReturn(ctx, tx, "user update", query, args); execErr != nil {
		c.rollbackTransaction(tx)
		return fmt.Errorf("error doing something: %w", err)
	}

	c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserUpdateEventEntry(updated.ID, nil))

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// UpdateUserPassword updates a user's authentication hash in the database.
func (c *Client) UpdateUserPassword(ctx context.Context, userID uint64, newHash string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue(keys.UserIDKey, userID).Debug("UpdateUserPassword called")

	query, args := c.sqlQueryBuilder.BuildUpdateUserPasswordQuery(userID, newHash)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	if execErr := c.performWriteQueryIgnoringReturn(ctx, tx, "user authentication update", query, args); execErr != nil {
		c.rollbackTransaction(tx)
		return fmt.Errorf("error doing something: %w", err)
	}

	c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserUpdatePasswordEventEntry(userID))

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// UpdateUserTwoFactorSecret marks a user's two factor secret as validated.
func (c *Client) UpdateUserTwoFactorSecret(ctx context.Context, userID uint64, newSecret string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue(keys.UserIDKey, userID).Debug("UpdateUserTwoFactorSecret called")

	query, args := c.sqlQueryBuilder.BuildUpdateUserTwoFactorSecretQuery(userID, newSecret)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	if execErr := c.performWriteQueryIgnoringReturn(ctx, tx, "user two factor secret update", query, args); execErr != nil {
		c.rollbackTransaction(tx)
		return fmt.Errorf("error doing something: %w", err)
	}

	c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserUpdateTwoFactorSecretEventEntry(userID))

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// VerifyUserTwoFactorSecret marks a user's two factor secret as validated.
func (c *Client) VerifyUserTwoFactorSecret(ctx context.Context, userID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue(keys.UserIDKey, userID).Debug("VerifyUserTwoFactorSecret called")

	query, args := c.sqlQueryBuilder.BuildVerifyUserTwoFactorSecretQuery(userID)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	if execErr := c.performWriteQueryIgnoringReturn(ctx, tx, "user two factor secret verification", query, args); execErr != nil {
		c.rollbackTransaction(tx)
		return fmt.Errorf("error doing something: %w", err)
	}

	c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserVerifyTwoFactorSecretEventEntry(userID))

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// ArchiveUser archives a user.
func (c *Client) ArchiveUser(ctx context.Context, userID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue(keys.UserIDKey, userID).Debug("ArchiveUser called")

	query, args := c.sqlQueryBuilder.BuildArchiveUserQuery(userID)

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	if execErr := c.performWriteQueryIgnoringReturn(ctx, tx, "user archive", query, args); execErr != nil {
		c.rollbackTransaction(tx)
		return fmt.Errorf("error doing something: %w", err)
	}

	c.createAuditLogEntryInTransaction(ctx, tx, audit.BuildUserArchiveEventEntry(userID))

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("error committing transaction: %w", err)
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
