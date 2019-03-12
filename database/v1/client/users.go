package dbclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

var _ models.UserHandler = (*Client)(nil)

// AdminUserExists executes a query to determine if an admin user has been established in the database
func (c *Client) AdminUserExists(ctx context.Context) (bool, error) {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "AdminUserExists")
	defer span.Finish()

	c.logger.Debug("AdminUserExists called")

	return c.database.AdminUserExists(ctx)
}

// GetUser fetches a user
func (c *Client) GetUser(ctx context.Context, userID uint64) (*models.User, error) {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "GetUser")
	defer span.Finish()

	c.logger.WithValue("user_id", userID).Debug("GetUser called")

	return c.database.GetUser(ctx, userID)
}

// GetUserByUsername fetches a user by their username
func (c *Client) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "GetUser")
	defer span.Finish()

	c.logger.WithValue("username", username).Debug("GetUserByUsername called")

	return c.database.GetUserByUsername(ctx, username)
}

// GetUserCount fetches a count of users from the postgres database that meet a particular filter
func (c *Client) GetUserCount(ctx context.Context, filter *models.QueryFilter) (count uint64, err error) {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "GetUserCount")
	defer span.Finish()

	if filter == nil {
		c.logger.Debug("using default query filter")
		filter = models.DefaultQueryFilter
	}

	c.logger.WithValue("filter", filter).Debug("GetUserCount called")

	return c.database.GetUserCount(ctx, filter)
}

// GetUsers fetches a list of users from the postgres database that meet a particular filter
func (c *Client) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "GetUsers")
	defer span.Finish()

	if filter == nil {
		c.logger.Debug("using default query filter")
		filter = models.DefaultQueryFilter
	}

	c.logger.WithValue("filter", filter).Debug("GetUsers called")

	return c.database.GetUsers(ctx, filter)
}

// CreateUser creates a user
func (c *Client) CreateUser(ctx context.Context, input *models.UserInput) (*models.User, error) {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "CreateUser")
	defer span.Finish()

	logger := c.logger.WithValues(map[string]interface{}{
		"username": input.Username,
		"is_admin": input.IsAdmin,
	})
	logger.Debug("CreateUser called")

	return c.database.CreateUser(ctx, input)
}

// UpdateUser receives a complete User struct and updates its place in the database.
// NOTE this function uses the ID provided in the input to make its query.
func (c *Client) UpdateUser(ctx context.Context, updated *models.User) error {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "UpdateUser")
	defer span.Finish()

	c.logger.WithValues(map[string]interface{}{
		"username": updated.Username,
		"is_admin": updated.IsAdmin,
	}).Debug("UpdateUser called")

	return c.database.UpdateUser(ctx, updated)
}

// DeleteUser deletes a user by their username
func (c *Client) DeleteUser(ctx context.Context, userID uint64) error {
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "DeleteUser")
	defer span.Finish()

	c.logger.WithValue("user_id", userID).Debug("DeleteUser called")

	return c.database.DeleteUser(ctx, userID)
}
