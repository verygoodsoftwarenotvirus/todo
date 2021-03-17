package httpclient

import (
	"context"
	"fmt"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// GetUser retrieves a user.
func (c *Client) GetUser(ctx context.Context, userID uint64) (user *types.User, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return nil, ErrInvalidIDProvided
	}

	req, err := c.requestBuilder.BuildGetUserRequest(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.retrieve(ctx, req, &user)

	return user, err
}

// GetUsers retrieves a list of users.
func (c *Client) GetUsers(ctx context.Context, filter *types.QueryFilter) (*types.UserList, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	users := &types.UserList{}

	req, err := c.requestBuilder.BuildGetUsersRequest(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.retrieve(ctx, req, &users)

	return users, err
}

// SearchForUsersByUsername retrieves a list of users.
func (c *Client) SearchForUsersByUsername(ctx context.Context, username string) (users []*types.User, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if username == "" {
		return nil, ErrEmptyUsernameProvided
	}

	req, err := c.requestBuilder.BuildSearchForUsersByUsernameRequest(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.retrieve(ctx, req, &users)

	return users, err
}

// CreateUser creates a new user.
func (c *Client) CreateUser(ctx context.Context, input *types.NewUserCreationInput) (*types.UserCreationResponse, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	// deliberately not validating here
	// maybe I should make a client-side validate method vs a server-side?

	user := &types.UserCreationResponse{}

	req, err := c.requestBuilder.BuildCreateUserRequest(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.executeUnauthenticatedDataRequest(ctx, req, &user)

	return user, err
}

// ArchiveUser archives a user.
func (c *Client) ArchiveUser(ctx context.Context, userID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return ErrInvalidIDProvided
	}

	req, err := c.requestBuilder.BuildArchiveUserRequest(ctx, userID)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// GetAuditLogForUser retrieves a list of audit log entries pertaining to a user.
func (c *Client) GetAuditLogForUser(ctx context.Context, userID uint64) (entries []*types.AuditLogEntry, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return nil, ErrInvalidIDProvided
	}

	req, err := c.requestBuilder.BuildGetAuditLogForUserRequest(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &entries); retrieveErr != nil {
		return nil, retrieveErr
	}

	return entries, nil
}

// UploadAvatar uploads a new avatar.
func (c *Client) UploadAvatar(ctx context.Context, avatar []byte, extension string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if len(avatar) == 0 {
		return fmt.Errorf("invalid length avatar passed: %d", len(avatar))
	}

	switch strings.ToLower(strings.TrimSpace(extension)) {
	case "jpeg", "png", "gif":
		//
	default:
		return fmt.Errorf("invalid extension: %q", extension)
	}

	req, err := c.requestBuilder.BuildAvatarUploadRequest(ctx, avatar, extension)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}
