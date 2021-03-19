package http

import (
	"context"
	"fmt"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// GetUser retrieves a user.
func (c *Client) GetUser(ctx context.Context, userID uint64) (*types.User, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return nil, ErrInvalidIDProvided
	}

	logger := c.logger.WithValue(keys.UserIDKey, userID)

	req, err := c.requestBuilder.BuildGetUserRequest(ctx, userID)
	if err != nil {
		return nil, prepareError(err, logger, span, "building get user request")
	}

	var user *types.User
	if err = c.fetchAndUnmarshal(ctx, req, &user); err != nil {
		return nil, prepareError(err, logger, span, "fetching user")
	}

	return user, nil
}

// GetUsers retrieves a list of users.
func (c *Client) GetUsers(ctx context.Context, filter *types.QueryFilter) (*types.UserList, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.loggerWithFilter(filter)

	tracing.AttachQueryFilterToSpan(span, filter)

	req, err := c.requestBuilder.BuildGetUsersRequest(ctx, filter)
	if err != nil {
		return nil, prepareError(err, logger, span, "building users list request")
	}

	var users *types.UserList
	if err = c.fetchAndUnmarshal(ctx, req, &users); err != nil {
		return nil, prepareError(err, logger, span, "retrieving users")
	}

	return users, nil
}

// SearchForUsersByUsername searches for a user from a list of users by their username.
func (c *Client) SearchForUsersByUsername(ctx context.Context, username string) ([]*types.User, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if username == "" {
		return nil, ErrEmptyUsernameProvided
	}

	logger := c.logger.WithValue(keys.UsernameKey, username)

	req, err := c.requestBuilder.BuildSearchForUsersByUsernameRequest(ctx, username)
	if err != nil {
		return nil, prepareError(err, logger, span, "building username search request")
	}

	var users []*types.User
	if err = c.fetchAndUnmarshal(ctx, req, &users); err != nil {
		return nil, prepareError(err, logger, span, "searching for users")
	}

	return users, nil
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

	logger := c.logger.WithValue(keys.UsernameKey, input.Username)

	req, err := c.requestBuilder.BuildCreateUserRequest(ctx, input)
	if err != nil {
		return nil, prepareError(err, logger, span, "building create user request")
	}

	var user *types.UserCreationResponse
	if err = c.fetchAndUnmarshalWithoutAuthentication(ctx, req, &user); err != nil {
		return nil, prepareError(err, logger, span, "creating user")
	}

	return user, nil
}

// ArchiveUser archives a user.
func (c *Client) ArchiveUser(ctx context.Context, userID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return ErrInvalidIDProvided
	}

	logger := c.logger.WithValue(keys.UserIDKey, userID)

	req, err := c.requestBuilder.BuildArchiveUserRequest(ctx, userID)
	if err != nil {
		return prepareError(err, logger, span, "building archive user request")
	}

	if err = c.fetchAndUnmarshal(ctx, req, nil); err != nil {
		return prepareError(err, logger, span, "archiving user")
	}

	return nil
}

// GetAuditLogForUser retrieves a list of audit log entries pertaining to a user.
func (c *Client) GetAuditLogForUser(ctx context.Context, userID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return nil, ErrInvalidIDProvided
	}

	logger := c.logger.WithValue(keys.UserIDKey, userID)

	req, err := c.requestBuilder.BuildGetAuditLogForUserRequest(ctx, userID)
	if err != nil {
		return nil, prepareError(err, logger, span, "building get audit log entries for user request")
	}

	var entries []*types.AuditLogEntry
	if err = c.fetchAndUnmarshal(ctx, req, &entries); err != nil {
		return nil, prepareError(err, logger, span, "retrieving audit log entries for user")
	}

	return entries, nil
}

// UploadNewAvatar uploads a new avatar.
func (c *Client) UploadNewAvatar(ctx context.Context, avatar []byte, extension string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if len(avatar) == 0 {
		return fmt.Errorf("invalid length avatar passed: %d", len(avatar))
	}

	logger := c.logger

	switch strings.ToLower(strings.TrimSpace(extension)) {
	case "jpeg", "png", "gif":
		//
	default:
		err := fmt.Errorf("invalid extension: %q", extension)
		return prepareError(err, logger, span, "uploading avatar")
	}

	req, err := c.requestBuilder.BuildAvatarUploadRequest(ctx, avatar, extension)
	if err != nil {
		return prepareError(err, logger, span, "building avatar upload request")
	}

	if err = c.fetchAndUnmarshal(ctx, req, nil); err != nil {
		return prepareError(err, logger, span, "uploading avatar")
	}

	return nil
}
