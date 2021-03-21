package http

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/errs"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// GetAPIClient gets an API client.
func (c *Client) GetAPIClient(ctx context.Context, apiClientDatabaseID uint64) (*types.APIClient, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if apiClientDatabaseID == 0 {
		return nil, ErrInvalidIDProvided
	}

	logger := c.logger.WithValue(keys.APIClientDatabaseIDKey, apiClientDatabaseID)

	req, err := c.requestBuilder.BuildGetAPIClientRequest(ctx, apiClientDatabaseID)
	if err != nil {
		return nil, errs.PrepareError(err, logger, span, "building retrieve API client request")
	}

	var apiClient *types.APIClient
	if err = c.fetchAndUnmarshal(ctx, req, &apiClient); err != nil {
		return nil, errs.PrepareError(err, logger, span, "fetching api client")
	}

	return apiClient, nil
}

// GetAPIClients gets a list of API clients.
func (c *Client) GetAPIClients(ctx context.Context, filter *types.QueryFilter) (*types.APIClientList, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.loggerWithFilter(filter)

	tracing.AttachQueryFilterToSpan(span, filter)

	req, err := c.requestBuilder.BuildGetAPIClientsRequest(ctx, filter)
	if err != nil {
		return nil, errs.PrepareError(err, logger, span, "building API clients list request")
	}

	var apiClients *types.APIClientList
	if err = c.fetchAndUnmarshal(ctx, req, &apiClients); err != nil {
		return nil, errs.PrepareError(err, logger, span, "fetching api clients")
	}

	return apiClients, nil
}

// CreateAPIClient creates an API client.
func (c *Client) CreateAPIClient(ctx context.Context, cookie *http.Cookie, input *types.APICientCreationInput) (*types.APIClientCreationResponse, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if cookie == nil && c.authMethod != cookieAuthMethod {
		return nil, ErrCookieRequired
	}

	if input == nil {
		return nil, ErrNilInputProvided
	}

	// deliberately not validating here because it requires settings awareness
	logger := c.logger.WithValue(keys.NameKey, input.Name)

	var apiClientResponse *types.APIClientCreationResponse

	req, err := c.requestBuilder.BuildCreateAPIClientRequest(ctx, cookie, input)
	if err != nil {
		return nil, errs.PrepareError(err, logger, span, "building create API client request")
	}

	if err = c.fetchAndUnmarshal(ctx, req, &apiClientResponse); err != nil {
		return nil, errs.PrepareError(err, logger, span, "creating api client")
	}

	return apiClientResponse, nil
}

// ArchiveAPIClient archives an API client.
func (c *Client) ArchiveAPIClient(ctx context.Context, apiClientDatabaseID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if apiClientDatabaseID == 0 {
		return ErrInvalidIDProvided
	}

	logger := c.logger.WithValue(keys.APIClientDatabaseIDKey, apiClientDatabaseID)

	req, err := c.requestBuilder.BuildArchiveAPIClientRequest(ctx, apiClientDatabaseID)
	if err != nil {
		return errs.PrepareError(err, logger, span, "building archive API client request")
	}

	if err = c.fetchAndUnmarshal(ctx, req, nil); err != nil {
		return errs.PrepareError(err, logger, span, "archiving api client")
	}

	return nil
}

// GetAuditLogForAPIClient retrieves a list of audit log entries pertaining to an API client.
func (c *Client) GetAuditLogForAPIClient(ctx context.Context, apiClientDatabaseID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if apiClientDatabaseID == 0 {
		return nil, ErrInvalidIDProvided
	}

	logger := c.logger.WithValue(keys.APIClientDatabaseIDKey, apiClientDatabaseID)

	req, err := c.requestBuilder.BuildGetAuditLogForAPIClientRequest(ctx, apiClientDatabaseID)
	if err != nil {
		return nil, errs.PrepareError(err, logger, span, "building retrieve audit log entries for API client request")
	}

	var entries []*types.AuditLogEntry
	if err = c.fetchAndUnmarshal(ctx, req, &entries); err != nil {
		return nil, errs.PrepareError(err, logger, span, "retrieving plan")
	}

	return entries, nil
}
