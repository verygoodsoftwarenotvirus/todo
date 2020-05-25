package dbclient

import (
	"context"
	"errors"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

var (
	_ models.UserDataManager = (*Client)(nil)

	// ErrUserExists is a sentinel error for returning when a username is taken.
	ErrUserExists = errors.New("error: username already exists")
)

// GetUser fetches a user.
func (c *Client) GetUser(ctx context.Context, userID uint64) (*models.User, error) {
	ctx, span := tracing.StartSpan(ctx, "GetUser")
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue("user_id", userID).Debug("GetUser called")

	return c.querier.GetUser(ctx, userID)
}

// GetUserWithUnverifiedTwoFactorSecret fetches a user.
func (c *Client) GetUserWithUnverifiedTwoFactorSecret(ctx context.Context, userID uint64) (*models.User, error) {
	ctx, span := tracing.StartSpan(ctx, "GetUserWithUnverifiedTwoFactorSecret")
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue("user_id", userID).Debug("GetUserWithUnverifiedTwoFactorSecret called")

	return c.querier.GetUserWithUnverifiedTwoFactorSecret(ctx, userID)
}

// VerifyUserTwoFactorSecret marks a user's two factor secret as validated.
func (c *Client) VerifyUserTwoFactorSecret(ctx context.Context, userID uint64) error {
	ctx, span := tracing.StartSpan(ctx, "VerifyUserTwoFactorSecret")
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue("user_id", userID).Debug("VerifyUserTwoFactorSecret called")

	return c.querier.VerifyUserTwoFactorSecret(ctx, userID)
}

// GetUserByUsername fetches a user by their username.
func (c *Client) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	ctx, span := tracing.StartSpan(ctx, "GetUserByUsername")
	defer span.End()

	tracing.AttachUsernameToSpan(span, username)
	c.logger.WithValue("username", username).Debug("GetUserByUsername called")

	return c.querier.GetUserByUsername(ctx, username)
}

// GetAllUserCount fetches a count of users from the database that meet a particular filter.
func (c *Client) GetAllUserCount(ctx context.Context) (count uint64, err error) {
	ctx, span := tracing.StartSpan(ctx, "GetAllUserCount")
	defer span.End()

	c.logger.Debug("GetAllUserCount called")

	return c.querier.GetAllUserCount(ctx)
}

// GetUsers fetches a list of users from the database that meet a particular filter.
func (c *Client) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	ctx, span := tracing.StartSpan(ctx, "GetUsers")
	defer span.End()

	tracing.AttachFilterToSpan(span, filter)
	c.logger.WithValue("filter", filter).Debug("GetUsers called")

	return c.querier.GetUsers(ctx, filter)
}

// CreateUser creates a user.
func (c *Client) CreateUser(ctx context.Context, input models.UserDatabaseCreationInput) (*models.User, error) {
	ctx, span := tracing.StartSpan(ctx, "CreateUser")
	defer span.End()

	tracing.AttachUsernameToSpan(span, input.Username)
	c.logger.WithValue("username", input.Username).Debug("CreateUser called")

	return c.querier.CreateUser(ctx, input)
}

// UpdateUser receives a complete User struct and updates its record in the database.
// NOTE: this function uses the ID provided in the input to make its query.
func (c *Client) UpdateUser(ctx context.Context, updated *models.User) error {
	ctx, span := tracing.StartSpan(ctx, "UpdateUser")
	defer span.End()

	tracing.AttachUsernameToSpan(span, updated.Username)
	c.logger.WithValue("username", updated.Username).Debug("UpdateUser called")

	return c.querier.UpdateUser(ctx, updated)
}

// ArchiveUser archives a user.
func (c *Client) ArchiveUser(ctx context.Context, userID uint64) error {
	ctx, span := tracing.StartSpan(ctx, "ArchiveUser")
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue("user_id", userID).Debug("ArchiveUser called")

	return c.querier.ArchiveUser(ctx, userID)
}
