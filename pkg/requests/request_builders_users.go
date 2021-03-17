package requests

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
func (c *Builder) BuildGetUserRequest(ctx context.Context, userID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := c.BuildURL(ctx, nil, usersBasePath, strconv.FormatUint(userID, 10))

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildGetUsersRequest builds an HTTP request for fetching a user.
func (c *Builder) BuildGetUsersRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(ctx, filter.ToValues(), usersBasePath)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildSearchForUsersByUsernameRequest builds an HTTP request that searches for a user.
func (c *Builder) BuildSearchForUsersByUsernameRequest(ctx context.Context, username string) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if username == "" {
		return nil, ErrEmptyUsernameProvided
	}

	u := c.buildRawURL(ctx, nil, usersBasePath, "search")
	q := u.Query()
	q.Set(types.SearchQueryKey, username)
	u.RawQuery = q.Encode()
	uri := u.String()

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildCreateUserRequest builds an HTTP request for creating a user.
func (c *Builder) BuildCreateUserRequest(ctx context.Context, input *types.NewUserCreationInput) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	// deliberately not validating here

	uri := c.buildVersionlessURL(ctx, nil, usersBasePath)

	return c.buildDataRequest(ctx, http.MethodPost, uri, input)
}

// BuildArchiveUserRequest builds an HTTP request for updating a user.
func (c *Builder) BuildArchiveUserRequest(ctx context.Context, userID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return nil, ErrInvalidIDProvided
	}

	// deliberately not validating here
	// maybe I should make a client-side validate method vs a server-side?

	uri := c.buildRawURL(ctx, nil, usersBasePath, strconv.FormatUint(userID, 10)).String()

	return http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
}

// BuildGetAuditLogForUserRequest builds an HTTP request for fetching a list of audit log entries for a user.
func (c *Builder) BuildGetAuditLogForUserRequest(ctx context.Context, userID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if userID == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := c.BuildURL(ctx, nil, usersBasePath, strconv.FormatUint(userID, 10), "audit")
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildAvatarUploadRequest builds a new avatar upload request.
func (c *Builder) BuildAvatarUploadRequest(ctx context.Context, avatar []byte, extension string) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if len(avatar) == 0 {
		return nil, fmt.Errorf("invalid length avatar passed: %d", len(avatar))
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

	uri := c.BuildURL(ctx, nil, usersBasePath, "avatar", "upload")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, body)
	if err != nil {
		return nil, fmt.Errorf("building HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Upload-Content-Type", ct)

	return req, nil
}
