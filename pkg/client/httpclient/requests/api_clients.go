package requests

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

const (
	apiClientsBasePath = "api_clients"
)

// BuildGetAPIClientRequest builds an HTTP request for fetching an API client.
func (b *Builder) BuildGetAPIClientRequest(ctx context.Context, clientID string) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if clientID == "" {
		return nil, ErrInvalidIDProvided
	}

	logger := b.logger.WithValue(keys.APIClientDatabaseIDKey, clientID)

	uri := b.BuildURL(
		ctx,
		nil,
		apiClientsBasePath,
		clientID,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building user status request")
	}

	return req, nil
}

// BuildGetAPIClientsRequest builds an HTTP request for fetching a list of API clients.
func (b *Builder) BuildGetAPIClientsRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	logger := filter.AttachToLogger(b.logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	uri := b.BuildURL(ctx, filter.ToValues(), apiClientsBasePath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building user status request")
	}

	return req, nil
}

// BuildCreateAPIClientRequest builds an HTTP request for creating an API client.
func (b *Builder) BuildCreateAPIClientRequest(ctx context.Context, cookie *http.Cookie, input *types.APIClientCreationInput) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if cookie == nil {
		return nil, ErrCookieRequired
	}

	if input == nil {
		return nil, ErrNilInputProvided
	}

	uri := b.BuildURL(ctx, nil, apiClientsBasePath)

	req, err := b.buildDataRequest(ctx, http.MethodPost, uri, input)
	if err != nil {
		return nil, err
	}

	req.AddCookie(cookie)

	return req, nil
}

// BuildArchiveAPIClientRequest builds an HTTP request for archiving an API client.
func (b *Builder) BuildArchiveAPIClientRequest(ctx context.Context, clientID string) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if clientID == "" {
		return nil, ErrInvalidIDProvided
	}

	logger := b.logger.WithValue(keys.APIClientDatabaseIDKey, clientID)

	uri := b.BuildURL(
		ctx,
		nil,
		apiClientsBasePath,
		clientID,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building user status request")
	}

	return req, nil
}
