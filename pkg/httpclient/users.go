package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	usersBasePath = "users"
)

// BuildGetUserRequest builds an HTTP request for fetching a user.
func (c *Client) BuildGetUserRequest(ctx context.Context, userID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(nil, usersBasePath, strconv.FormatUint(userID, 10))

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetUser retrieves a user.
func (c *Client) GetUser(ctx context.Context, userID uint64) (user *types.User, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildGetUserRequest(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.retrieve(ctx, req, &user)

	return user, err
}

// BuildGetUsersRequest builds an HTTP request for fetching a user.
func (c *Client) BuildGetUsersRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(filter.ToValues(), usersBasePath)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetUsers retrieves a list of users.
func (c *Client) GetUsers(ctx context.Context, filter *types.QueryFilter) (*types.UserList, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	users := &types.UserList{}

	req, err := c.BuildGetUsersRequest(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.retrieve(ctx, req, &users)

	return users, err
}

// BuildSearchForUsersByUsernameRequest builds an HTTP request that searches for a user.
func (c *Client) BuildSearchForUsersByUsernameRequest(ctx context.Context, username string) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	u := c.buildRawURL(nil, usersBasePath, "search")
	q := u.Query()
	q.Set(types.SearchQueryKey, username)
	u.RawQuery = q.Encode()
	uri := u.String()

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// SearchForUsersByUsername retrieves a list of users.
func (c *Client) SearchForUsersByUsername(ctx context.Context, username string) (users []*types.User, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildSearchForUsersByUsernameRequest(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.retrieve(ctx, req, &users)

	return users, err
}

// BuildCreateUserRequest builds an HTTP request for creating a user.
func (c *Client) BuildCreateUserRequest(ctx context.Context, body *types.NewUserCreationInput) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.buildVersionlessURL(nil, usersBasePath)

	return c.buildDataRequest(ctx, http.MethodPost, uri, body)
}

// CreateUser creates a new user.
func (c *Client) CreateUser(ctx context.Context, input *types.NewUserCreationInput) (*types.UserCreationResponse, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	user := &types.UserCreationResponse{}

	req, err := c.BuildCreateUserRequest(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.executeUnauthenticatedDataRequest(ctx, req, &user)

	return user, err
}

// BuildArchiveUserRequest builds an HTTP request for updating a user.
func (c *Client) BuildArchiveUserRequest(ctx context.Context, userID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.buildRawURL(nil, usersBasePath, strconv.FormatUint(userID, 10)).String()

	return http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
}

// ArchiveUser archives a user.
func (c *Client) ArchiveUser(ctx context.Context, userID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildArchiveUserRequest(ctx, userID)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// BuildGetAuditLogForUserRequest builds an HTTP request for fetching a list of audit log entries for a user.
func (c *Client) BuildGetAuditLogForUserRequest(ctx context.Context, userID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(
		nil,
		usersBasePath,
		strconv.FormatUint(userID, 10),
		"audit",
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetAuditLogForUser retrieves a list of audit log entries pertaining to a user.
func (c *Client) GetAuditLogForUser(ctx context.Context, userID uint64) (entries []*types.AuditLogEntry, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildGetAuditLogForUserRequest(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &entries); retrieveErr != nil {
		return nil, retrieveErr
	}

	return entries, nil
}

// BuildAvatarUploadRequest builds a new avatar upload request.
func (c *Client) BuildAvatarUploadRequest(ctx context.Context, avatar []byte, extension string) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("avatar", fmt.Sprintf("avatar.%s", extension))
	if err != nil {
		return nil, fmt.Errorf("writing to form file: %w", err)
	}

	if _, copyErr := io.Copy(part, bytes.NewReader(avatar)); copyErr != nil {
		return nil, fmt.Errorf("copying file contents to request: %w", copyErr)
	}

	if closeErr := writer.Close(); closeErr != nil {
		return nil, fmt.Errorf("closing avatar file: %w", closeErr)
	}

	uri := c.BuildURL(
		nil,
		usersBasePath,
		"avatar",
		"upload",
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, body)
	if err != nil {
		return nil, fmt.Errorf("building HTTP request: %w", err)
	}

	var ct string

	switch strings.ToLower(strings.TrimSpace(extension)) {
	case "jpeg":
		ct = "image/jpeg"
	case "png":
		ct = "image/png"
	case "gif":
		ct = "image/gif"
	default:
		return nil, fmt.Errorf("invalid extension: %q", extension)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Upload-Content-Type", ct)

	return req, nil
}

// UploadAvatar uploads a new avatar.
func (c *Client) UploadAvatar(ctx context.Context, avatar []byte, extension string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildAvatarUploadRequest(ctx, avatar, extension)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}
