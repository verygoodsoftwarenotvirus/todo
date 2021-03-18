package httpclient

import (
	"context"
	"fmt"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
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
		tracing.AttachErrorToSpan(span, err)
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &user); retrieveErr != nil {
		tracing.AttachErrorToSpan(span, retrieveErr)
		return nil, fmt.Errorf("fetching user: %w", retrieveErr)
	}

	return user, nil
}

// GetUsers retrieves a list of users.
func (c *Client) GetUsers(ctx context.Context, filter *types.QueryFilter) (*types.UserList, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	users := &types.UserList{}

	req, err := c.requestBuilder.BuildGetUsersRequest(ctx, filter)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &users); retrieveErr != nil {
		tracing.AttachErrorToSpan(span, retrieveErr)
		return nil, fmt.Errorf("fetching users: %w", retrieveErr)
	}

	return users, nil
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
		tracing.AttachErrorToSpan(span, err)
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &users); retrieveErr != nil {
		tracing.AttachErrorToSpan(span, retrieveErr)
		return nil, fmt.Errorf("searching for users: %w", retrieveErr)
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

	user := &types.UserCreationResponse{}

	req, err := c.requestBuilder.BuildCreateUserRequest(ctx, input)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return nil, fmt.Errorf("building request: %w", err)
	}

	if createErr := c.executeUnauthenticatedDataRequest(ctx, req, &user); createErr != nil {
		tracing.AttachErrorToSpan(span, createErr)
		return nil, fmt.Errorf("creating user: %w", createErr)
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

	req, err := c.requestBuilder.BuildArchiveUserRequest(ctx, userID)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return fmt.Errorf("building request: %w", err)
	}

	if archiveErr := c.executeRequest(ctx, req, nil); archiveErr != nil {
		tracing.AttachErrorToSpan(span, archiveErr)
		return fmt.Errorf("archiving user: %w", archiveErr)
	}

	return nil
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
		tracing.AttachErrorToSpan(span, err)
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &entries); retrieveErr != nil {
		tracing.AttachErrorToSpan(span, retrieveErr)
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
		err := fmt.Errorf("invalid extension: %q", extension)
		tracing.AttachErrorToSpan(span, err)
		return err
	}

	req, err := c.requestBuilder.BuildAvatarUploadRequest(ctx, avatar, extension)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return fmt.Errorf("building request: %w", err)
	}

	if uploadErr := c.executeRequest(ctx, req, nil); uploadErr != nil {
		tracing.AttachErrorToSpan(span, uploadErr)
		return fmt.Errorf("uploading avatar: %w", uploadErr)
	}

	return nil
}
