package dbclient

import (
	"context"
	"errors"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"go.opencensus.io/trace"
)

var (
	_ models.UserDataManager = (*Client)(nil)

	// ErrUserExists is a sentinel error for returning when a username is taken
	ErrUserExists = errors.New("error: username already exists")
)

// GetUser fetches a user
func (c *Client) GetUser(ctx context.Context, userID uint64) (*models.User, error) {
	ctx, span := trace.StartSpan(ctx, "GetUser")
	defer span.End()

	span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))
	c.logger.WithValue("user_id", userID).Debug("GetUser called")

	return c.querier.GetUser(ctx, userID)
}

// GetUserByUsername fetches a user by their username
func (c *Client) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	ctx, span := trace.StartSpan(ctx, "GetUserByUsername")
	defer span.End()

	span.AddAttributes(trace.StringAttribute("username", username))
	c.logger.WithValue("username", username).Debug("GetUserByUsername called")

	return c.querier.GetUserByUsername(ctx, username)
}

// GetUserCount fetches a count of users from the postgres querier that meet a particular filter
func (c *Client) GetUserCount(ctx context.Context, filter *models.QueryFilter) (count uint64, err error) {
	ctx, span := trace.StartSpan(ctx, "GetUserCount")
	defer span.End()

	logger := c.logger.WithValue("filter", filter)
	logger.Debug("GetUserCount called")

	if filter == nil {
		c.logger.Debug("using default query filter")
		filter = models.DefaultQueryFilter()
	}

	c.logger.WithValue("filter", filter).Debug("GetUserCount called")

	return c.querier.GetUserCount(ctx, filter)
}

// GetUsers fetches a list of users from the postgres querier that meet a particular filter
func (c *Client) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	ctx, span := trace.StartSpan(ctx, "GetUsers")
	defer span.End()

	logger := c.logger.WithValue("filter", filter)
	logger.Debug("GetUsers called")

	if filter == nil {
		logger.Debug("using default query filter")
		filter = models.DefaultQueryFilter()
	}

	c.logger.WithValue("filter", filter).Debug("GetUsers called")

	return c.querier.GetUsers(ctx, filter)
}

// CreateUser creates a user
func (c *Client) CreateUser(ctx context.Context, input *models.UserInput) (*models.User, error) {
	ctx, span := trace.StartSpan(ctx, "CreateUser")
	defer span.End()
	span.AddAttributes(
		trace.StringAttribute("username", input.Username),
	)

	logger := c.logger.WithValues(map[string]interface{}{
		"username": input.Username,
	})
	logger.Debug("CreateUser called")

	return c.querier.CreateUser(ctx, input)
}

// UpdateUser receives a complete User struct and updates its place in the querier.
// NOTE this function uses the ID provided in the input to make its query.
func (c *Client) UpdateUser(ctx context.Context, updated *models.User) error {
	ctx, span := trace.StartSpan(ctx, "UpdateUser")
	defer span.End()
	span.AddAttributes(
		trace.StringAttribute("username", updated.Username),
	)

	c.logger.WithValues(map[string]interface{}{
		"username": updated.Username,
	}).Debug("UpdateUser called")

	return c.querier.UpdateUser(ctx, updated)
}

// DeleteUser deletes a user by their username
func (c *Client) DeleteUser(ctx context.Context, userID uint64) error {
	ctx, span := trace.StartSpan(ctx, "DeleteUser")
	defer span.End()
	span.AddAttributes(trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)))

	c.logger.WithValue("user_id", userID).Debug("DeleteUser called")

	return c.querier.DeleteUser(ctx, userID)
}
