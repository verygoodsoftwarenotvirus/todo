package requests

import (
	"context"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	apiClientsBasePath = "api_clients"
)

// BuildGetAPIClientRequest builds an HTTP request for fetching an OAuth2 client.
func (c *Builder) BuildGetAPIClientRequest(ctx context.Context, id uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if id == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := c.BuildURL(
		ctx,
		nil,
		apiClientsBasePath,
		strconv.FormatUint(id, 10),
	)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildGetAPIClientsRequest builds an HTTP request for fetching a list of OAuth2 clients.
func (c *Builder) BuildGetAPIClientsRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(ctx, filter.ToValues(), apiClientsBasePath)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildCreateAPIClientRequest builds an HTTP request for creating OAuth2 clients.
func (c *Builder) BuildCreateAPIClientRequest(ctx context.Context, cookie *http.Cookie, input *types.APICientCreationInput) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if cookie == nil {
		return nil, ErrCookieRequired
	}

	if input == nil {
		return nil, ErrNilInputProvided
	}

	// deliberately not validating here because it requires settings awareness

	uri := c.BuildURL(ctx, nil, apiClientsBasePath)

	req, err := c.buildDataRequest(ctx, http.MethodPost, uri, input)
	if err != nil {
		return nil, err
	}

	req.AddCookie(cookie)

	return req, nil
}

// BuildArchiveAPIClientRequest builds an HTTP request for archiving an oauth2 client.
func (c *Builder) BuildArchiveAPIClientRequest(ctx context.Context, id uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if id == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := c.BuildURL(
		ctx,
		nil,
		apiClientsBasePath,
		strconv.FormatUint(id, 10),
	)

	return http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
}

// BuildGetAuditLogForAPIClientRequest builds an HTTP request for fetching a list of audit log entries for an oauth2 client.
func (c *Builder) BuildGetAuditLogForAPIClientRequest(ctx context.Context, clientID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if clientID == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := c.BuildURL(ctx, nil, apiClientsBasePath, strconv.FormatUint(clientID, 10), "audit")
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}
