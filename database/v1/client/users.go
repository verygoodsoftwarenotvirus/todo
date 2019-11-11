package dbclient

import (
	"context"
	"errors"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"go.opencensus.io/trace"
)

var (
	_ models.UserDataManager = (*Client)(nil)

	// ErrUserExists is a sentinel error for returning when a username is taken
	ErrUserExists = errors.New("error: username already exists")
)

func attachUsernameToSpan(span *trace.Span, username string) {
	if span != nil {
		span.AddAttributes(trace.StringAttribute("username", username))
	}
}

// GetUser fetches a user
func (c *Client) GetUser(ctx context.Context, userID uint64) (*models.User, error) {
	ctx, span := trace.StartSpan(ctx, "GetUser")
	defer span.End()

	attachUserIDToSpan(span, userID)
	c.logger.WithValue("user_id", userID).Debug("GetUser called")

	return c.querier.GetUser(ctx, userID)
}

// GetUserByUsername fetches a user by their username
func (c *Client) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	ctx, span := trace.StartSpan(ctx, "GetUserByUsername")
	defer span.End()

	attachUsernameToSpan(span, username)
	c.logger.WithValue("username", username).Debug("GetUserByUsername called")

	return c.querier.GetUserByUsername(ctx, username)
}

// GetUserCount fetches a count of users from the database that meet a particular filter
func (c *Client) GetUserCount(ctx context.Context, filter *models.QueryFilter) (count uint64, err error) {
	ctx, span := trace.StartSpan(ctx, "GetUserCount")
	defer span.End()

	attachFilterToSpan(span, filter)
	c.logger.Debug("GetUserCount called")

	return c.querier.GetUserCount(ctx, filter)
}

// GetUsers fetches a list of users from the database that meet a particular filter
func (c *Client) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	ctx, span := trace.StartSpan(ctx, "GetUsers")
	defer span.End()

	attachFilterToSpan(span, filter)
	c.logger.WithValue("filter", filter).Debug("GetUsers called")

	return c.querier.GetUsers(ctx, filter)
}

// CreateUser creates a user
func (c *Client) CreateUser(ctx context.Context, input *models.UserInput) (*models.User, error) {
	ctx, span := trace.StartSpan(ctx, "CreateUser")
	defer span.End()

	attachUsernameToSpan(span, input.Username)
	c.logger.WithValue("username", input.Username).Debug("CreateUser called")

	return c.querier.CreateUser(ctx, input)
}

// UpdateUser receives a complete User struct and updates its record in the database.
// NOTE: this function uses the ID provided in the input to make its query.
func (c *Client) UpdateUser(ctx context.Context, updated *models.User) error {
	ctx, span := trace.StartSpan(ctx, "UpdateUser")
	defer span.End()

	attachUsernameToSpan(span, updated.Username)
	c.logger.WithValue("username", updated.Username).Debug("UpdateUser called")

	return c.querier.UpdateUser(ctx, updated)
}

// ArchiveUser archives a user
func (c *Client) ArchiveUser(ctx context.Context, userID uint64) error {
	ctx, span := trace.StartSpan(ctx, "ArchiveUser")
	defer span.End()

	attachUserIDToSpan(span, userID)
	c.logger.WithValue("user_id", userID).Debug("ArchiveUser called")

	return c.querier.ArchiveUser(ctx, userID)
}
