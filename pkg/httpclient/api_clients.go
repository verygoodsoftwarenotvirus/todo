package httpclient

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// GetAPIClient gets an OAuth2 client.
func (c *Client) GetAPIClient(ctx context.Context, id uint64) (apiClient *types.APIClient, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if id == 0 {
		return nil, ErrInvalidIDProvided
	}

	req, err := c.requestBuilder.BuildGetAPIClientRequest(ctx, id)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.retrieve(ctx, req, &apiClient)

	return apiClient, err
}

// GetAPIClients gets a list of OAuth2 clients.
func (c *Client) GetAPIClients(ctx context.Context, filter *types.QueryFilter) (*types.APIClientList, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.requestBuilder.BuildGetAPIClientsRequest(ctx, filter)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return nil, fmt.Errorf("building request: %w", err)
	}

	var apiClients *types.APIClientList
	err = c.retrieve(ctx, req, &apiClients)

	return apiClients, err
}

// CreateAPIClient creates an OAuth2 client. Note that cookie must not be nil in order to receive a valid response.
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

	var apiClientResponse *types.APIClientCreationResponse

	req, err := c.requestBuilder.BuildCreateAPIClientRequest(ctx, cookie, input)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return nil, err
	}

	if resErr := c.executeRequest(ctx, req, &apiClientResponse); resErr != nil {
		tracing.AttachErrorToSpan(span, resErr)
		return nil, fmt.Errorf("executing request: %w", resErr)
	}

	return apiClientResponse, nil
}

// ArchiveAPIClient archives an OAuth2 client.
func (c *Client) ArchiveAPIClient(ctx context.Context, id uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if id == 0 {
		return ErrInvalidIDProvided
	}

	req, err := c.requestBuilder.BuildArchiveAPIClientRequest(ctx, id)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// GetAuditLogForAPIClient retrieves a list of audit log entries pertaining to an oauth2 client.
func (c *Client) GetAuditLogForAPIClient(ctx context.Context, clientID uint64) (entries []*types.AuditLogEntry, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if clientID == 0 {
		return nil, ErrInvalidIDProvided
	}

	req, err := c.requestBuilder.BuildGetAuditLogForAPIClientRequest(ctx, clientID)
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
