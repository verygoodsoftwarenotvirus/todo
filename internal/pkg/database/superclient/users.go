package superclient

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
	_ types.UserDataManager  = (*Client)(nil)
	_ types.UserAuditManager = (*Client)(nil)

	// ErrUserExists is a sentinel error for returning when a username is taken.
	ErrUserExists = errors.New("error: username already exists")
)

// scanUser provides a consistent way to scan something like a *sql.Row into a User struct.
func (c *Client) scanUser(scan database.Scanner, includeCounts bool) (user *types.User, filteredCount, totalCount uint64, err error) {
	user = &types.User{}

	var perms uint32

	targetVars := []interface{}{
		&user.ID,
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

	user.AdminPermissions = bitmask.NewPermissionBitmask(perms)

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

// GetUser fetches a user.
func (c *Client) GetUser(ctx context.Context, userID uint64) (*types.User, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	logger := c.logger.WithValue(keys.UserIDKey, userID)

	logger.Debug("GetUser called")

	query, args := c.sqlQueryBuilder.BuildGetUserQuery(userID)
	row := c.db.QueryRowContext(ctx, query, args...)

	u, _, _, err := c.scanUser(row, false)
	if err != nil {
		return nil, fmt.Errorf("fetching user from database: %w", err)
	}

	return u, err
}

// GetUserWithUnverifiedTwoFactorSecret fetches a user with an unverified 2FA secret.
func (c *Client) GetUserWithUnverifiedTwoFactorSecret(ctx context.Context, userID uint64) (*types.User, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	logger := c.logger.WithValue(keys.UserIDKey, userID)

	logger.Debug("GetUserWithUnverifiedTwoFactorSecret called")

	query, args := c.sqlQueryBuilder.BuildGetUserWithUnverifiedTwoFactorSecretQuery(userID)
	row := c.db.QueryRowContext(ctx, query, args...)

	u, _, _, err := c.scanUser(row, false)
	if err != nil {
		return nil, fmt.Errorf("fetching user from database: %w", err)
	}

	return u, err
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

		return nil, fmt.Errorf("fetching user from database: %w", err)
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
		return nil, fmt.Errorf("error querying database for users: %w", err)
	}

	u, _, _, err := c.scanUsers(rows, false)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, fmt.Errorf("fetching user from database: %w", err)
	}

	return u, nil
}

// GetAllUsersCount fetches a count of users from the database that meet a particular filter.
func (c *Client) GetAllUsersCount(ctx context.Context) (count uint64, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllUsersCount called")

	query := c.sqlQueryBuilder.BuildGetAllUsersCountQuery()
	err = c.db.QueryRowContext(ctx, query).Scan(&count)

	return count, err
}

// GetUsers fetches a list of users from the database that meet a particular filter.
func (c *Client) GetUsers(ctx context.Context, filter *types.QueryFilter) (x *types.UserList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	x = &types.UserList{}
	logger := c.logger.WithValue(keys.FilterIsNilKey, filter == nil)

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit)
		x.Page, x.Limit = filter.Page, filter.Limit
	}

	logger.Debug("GetUsers called")

	query, args := c.sqlQueryBuilder.BuildGetUsersQuery(filter)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fetching user from database: %w", err)
	}

	x.Users, x.FilteredCount, x.TotalCount, err = c.scanUsers(rows, true)
	if err != nil {
		return nil, fmt.Errorf("loading response from database: %w", err)
	}

	return x, nil
}

// CreateUser creates a user.
func (c *Client) CreateUser(ctx context.Context, input types.UserDataStoreCreationInput) (*types.User, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUsernameToSpan(span, input.Username)
	logger := c.logger.WithValue(keys.UsernameKey, input.Username)

	logger.Debug("CreateUser called")

	x := &types.User{
		Username:        input.Username,
		HashedPassword:  input.HashedPassword,
		TwoFactorSecret: input.TwoFactorSecret,
	}
	query, args := c.sqlQueryBuilder.BuildCreateUserQuery(input)

	// create the user.
	res, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing user creation query: %w", err)
	}

	x.CreatedOn = c.timeTeller.Now()
	x.ID = c.getIDFromResult(res)

	return x, nil
}

// UpdateUser receives a complete User struct and updates its record in the database.
// NOTE: this function uses the ID provided in the input to make its query.
func (c *Client) UpdateUser(ctx context.Context, updated *types.User) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUsernameToSpan(span, updated.Username)
	c.logger.WithValue(keys.UsernameKey, updated.Username).Debug("UpdateUser called")

	query, args := c.sqlQueryBuilder.BuildUpdateUserQuery(updated)
	_, err := c.db.ExecContext(ctx, query, args...)

	return err
}

// UpdateUserPassword updates a user's password hash in the database.
func (c *Client) UpdateUserPassword(ctx context.Context, userID uint64, newHash string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue(keys.UserIDKey, userID).Debug("UpdateUserPassword called")

	query, args := c.sqlQueryBuilder.BuildUpdateUserPasswordQuery(userID, newHash)
	_, err := c.db.ExecContext(ctx, query, args...)

	return err
}

// VerifyUserTwoFactorSecret marks a user's two factor secret as validated.
func (c *Client) VerifyUserTwoFactorSecret(ctx context.Context, userID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue(keys.UserIDKey, userID).Debug("VerifyUserTwoFactorSecret called")

	query, args := c.sqlQueryBuilder.BuildVerifyUserTwoFactorSecretQuery(userID)
	_, err := c.db.ExecContext(ctx, query, args...)

	return err
}

// ArchiveUser archives a user.
func (c *Client) ArchiveUser(ctx context.Context, userID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue(keys.UserIDKey, userID).Debug("ArchiveUser called")

	query, args := c.sqlQueryBuilder.BuildArchiveUserQuery(userID)
	_, err := c.db.ExecContext(ctx, query, args...)

	return err
}

// LogUserCreationEvent saves a UserCreationEvent in the audit log table.
func (c *Client) LogUserCreationEvent(ctx context.Context, user *types.User) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, user.ID)
	c.logger.WithValue(keys.UserIDKey, user.ID).Debug("LogUserCreationEvent called")

	c.createAuditLogEntry(ctx, audit.BuildUserCreationEventEntry(user))
}

// LogUserVerifyTwoFactorSecretEvent saves a UserVerifyTwoFactorSecretEvent in the audit log table.
func (c *Client) LogUserVerifyTwoFactorSecretEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogUserVerifyTwoFactorSecretEvent called")

	c.createAuditLogEntry(ctx, audit.BuildUserVerifyTwoFactorSecretEventEntry(userID))
}

// LogUserUpdateTwoFactorSecretEvent saves a UserUpdateTwoFactorSecretEvent in the audit log table.
func (c *Client) LogUserUpdateTwoFactorSecretEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogUserUpdateTwoFactorSecretEvent called")

	c.createAuditLogEntry(ctx, audit.BuildUserUpdateTwoFactorSecretEventEntry(userID))
}

// LogUserUpdatePasswordEvent saves a UserUpdatePasswordEvent in the audit log table.
func (c *Client) LogUserUpdatePasswordEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogUserUpdatePasswordEvent called")

	c.createAuditLogEntry(ctx, audit.BuildUserUpdatePasswordEventEntry(userID))
}

// LogUserArchiveEvent saves a UserArchiveEvent in the audit log table.
func (c *Client) LogUserArchiveEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogUserArchiveEvent called")

	c.createAuditLogEntry(ctx, audit.BuildUserArchiveEventEntry(userID))
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
