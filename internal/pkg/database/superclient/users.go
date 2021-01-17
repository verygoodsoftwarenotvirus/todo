package superclient

import (
	"context"
	"errors"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ types.UserDataManager = (*Client)(nil)

	// ErrUserExists is a sentinel error for returning when a username is taken.
	ErrUserExists = errors.New("error: username already exists")
)

// GetUser fetches a user.
func (c *Client) GetUser(ctx context.Context, userID uint64) (*types.User, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	logger := c.logger.WithValue(keys.UserIDKey, userID)

	user, err := c.querier.GetUser(ctx, userID)
	if err != nil {
		logger.Error(err, "querying database for user")
		return nil, err
	}

	logger.Debug("GetUser called")

	return user, nil
}

// GetUserWithUnverifiedTwoFactorSecret fetches a user with an unverified 2FA secret.
func (c *Client) GetUserWithUnverifiedTwoFactorSecret(ctx context.Context, userID uint64) (*types.User, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	logger := c.logger.WithValue(keys.UserIDKey, userID)

	user, err := c.querier.GetUserWithUnverifiedTwoFactorSecret(ctx, userID)
	if err != nil {
		logger.Error(err, "querying database for user")
		return nil, err
	}

	logger.Debug("GetUserWithUnverifiedTwoFactorSecret called")

	return user, nil
}

// VerifyUserTwoFactorSecret marks a user's two factor secret as validated.
func (c *Client) VerifyUserTwoFactorSecret(ctx context.Context, userID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue(keys.UserIDKey, userID).Debug("VerifyUserTwoFactorSecret called")

	return c.querier.VerifyUserTwoFactorSecret(ctx, userID)
}

// GetUserByUsername fetches a user by their username.
func (c *Client) GetUserByUsername(ctx context.Context, username string) (*types.User, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUsernameToSpan(span, username)
	logger := c.logger.WithValue(keys.UsernameKey, username)

	user, err := c.querier.GetUserByUsername(ctx, username)
	if err != nil {
		logger.Error(err, "querying database for user")
		return nil, err
	}

	logger.Debug("GetUserByUsername called")

	return user, nil
}

// SearchForUsersByUsername fetches a list of users whose usernames begin with a given query.
func (c *Client) SearchForUsersByUsername(ctx context.Context, usernameQuery string) ([]*types.User, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	user, err := c.querier.SearchForUsersByUsername(ctx, usernameQuery)
	if err != nil {
		c.logger.Error(err, "querying database for user")
		return nil, err
	}

	c.logger.Debug("SearchForUsersByUsername called")

	return user, nil
}

// GetAllUsersCount fetches a count of users from the database that meet a particular filter.
func (c *Client) GetAllUsersCount(ctx context.Context) (count uint64, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllUsersCount called")

	return c.querier.GetAllUsersCount(ctx)
}

// GetUsers fetches a list of users from the database that meet a particular filter.
func (c *Client) GetUsers(ctx context.Context, filter *types.QueryFilter) (*types.UserList, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit)
	}

	c.logger.WithValue(keys.FilterIsNilKey, filter == nil).Debug("GetUsers called")

	return c.querier.GetUsers(ctx, filter)
}

// CreateUser creates a user.
func (c *Client) CreateUser(ctx context.Context, input types.UserDataStoreCreationInput) (*types.User, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUsernameToSpan(span, input.Username)
	logger := c.logger.WithValue(keys.UsernameKey, input.Username)

	user, err := c.querier.CreateUser(ctx, input)
	if err != nil {
		logger.Error(err, "querying database for user")
		return nil, err
	}

	logger.Debug("CreateUser called")

	return user, nil
}

// UpdateUser receives a complete User struct and updates its record in the database.
// NOTE: this function uses the ID provided in the input to make its query.
func (c *Client) UpdateUser(ctx context.Context, updated *types.User) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUsernameToSpan(span, updated.Username)
	c.logger.WithValue(keys.UsernameKey, updated.Username).Debug("UpdateUser called")

	return c.querier.UpdateUser(ctx, updated)
}

// UpdateUserPassword updates a user's password hash in the database.
func (c *Client) UpdateUserPassword(ctx context.Context, userID uint64, newHash string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue(keys.UserIDKey, userID).Debug("UpdateUserPassword called")

	return c.querier.UpdateUserPassword(ctx, userID, newHash)
}

// ArchiveUser archives a user.
func (c *Client) ArchiveUser(ctx context.Context, userID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue(keys.UserIDKey, userID).Debug("ArchiveUser called")

	return c.querier.ArchiveUser(ctx, userID)
}

// LogUserCreationEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogUserCreationEvent(ctx context.Context, user *types.User) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, user.ID).Debug("LogUserCreationEvent called")

	c.querier.LogUserCreationEvent(ctx, user)
}

// LogUserVerifyTwoFactorSecretEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogUserVerifyTwoFactorSecretEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogUserVerifyTwoFactorSecretEvent called")

	c.querier.LogUserVerifyTwoFactorSecretEvent(ctx, userID)
}

// LogUserUpdateTwoFactorSecretEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogUserUpdateTwoFactorSecretEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogUserUpdateTwoFactorSecretEvent called")

	c.querier.LogUserUpdateTwoFactorSecretEvent(ctx, userID)
}

// LogUserUpdatePasswordEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogUserUpdatePasswordEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogUserUpdatePasswordEvent called")

	c.querier.LogUserUpdatePasswordEvent(ctx, userID)
}

// LogUserArchiveEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogUserArchiveEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogUserArchiveEvent called")

	c.querier.LogUserArchiveEvent(ctx, userID)
}

// GetAuditLogEntriesForUser fetches a list of audit log entries from the database that relate to a given user.
func (c *Client) GetAuditLogEntriesForUser(ctx context.Context, userID uint64) ([]types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("GetAuditLogEntriesForUser called")

	return c.querier.GetAuditLogEntriesForUser(ctx, userID)
}
