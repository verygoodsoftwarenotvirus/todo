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

// BuildGetAPIClientRequest builds an HTTP request for fetching an API client.
func (b *Builder) BuildGetAPIClientRequest(ctx context.Context, id uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if id == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := b.BuildURL(
		ctx,
		nil,
		apiClientsBasePath,
		strconv.FormatUint(id, 10),
	)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildGetAPIClientsRequest builds an HTTP request for fetching a list of API clients.
func (b *Builder) BuildGetAPIClientsRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	uri := b.BuildURL(ctx, filter.ToValues(), apiClientsBasePath)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildCreateAPIClientRequest builds an HTTP request for creating API clients.
func (b *Builder) BuildCreateAPIClientRequest(ctx context.Context, cookie *http.Cookie, input *types.APICientCreationInput) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if cookie == nil {
		return nil, ErrCookieRequired
	}

	if input == nil {
		return nil, ErrNilInputProvided
	}

	// deliberately not validating here because it requires settings awareness

	uri := b.BuildURL(ctx, nil, apiClientsBasePath)

	req, err := b.buildDataRequest(ctx, http.MethodPost, uri, input)
	if err != nil {
		return nil, err
	}

	req.AddCookie(cookie)

	return req, nil
}

// BuildArchiveAPIClientRequest builds an HTTP request for archiving an API client.
func (b *Builder) BuildArchiveAPIClientRequest(ctx context.Context, id uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if id == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := b.BuildURL(
		ctx,
		nil,
		apiClientsBasePath,
		strconv.FormatUint(id, 10),
	)

	return http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
}

// BuildGetAuditLogForAPIClientRequest builds an HTTP request for fetching a list of audit log entries for an API client.
func (b *Builder) BuildGetAuditLogForAPIClientRequest(ctx context.Context, clientID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if clientID == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := b.BuildURL(ctx, nil, apiClientsBasePath, strconv.FormatUint(clientID, 10), "audit")
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}
