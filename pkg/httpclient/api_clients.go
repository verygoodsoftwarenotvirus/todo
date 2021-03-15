package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	apiClientsBasePath = "api_clients"
)

// BuildGetAPIClientRequest builds an HTTP request for fetching an OAuth2 client.
func (c *Client) BuildGetAPIClientRequest(ctx context.Context, id uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if id == 0 {
		return nil, ErrZeroIDProvided
	}

	uri := c.BuildURL(
		ctx,
		nil,
		apiClientsBasePath,
		strconv.FormatUint(id, 10),
	)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetAPIClient gets an OAuth2 client.
func (c *Client) GetAPIClient(ctx context.Context, id uint64) (apiClient *types.APIClient, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if id == 0 {
		return nil, ErrZeroIDProvided
	}

	req, err := c.BuildGetAPIClientRequest(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.retrieve(ctx, req, &apiClient)

	return apiClient, err
}

// BuildGetAPIClientsRequest builds an HTTP request for fetching a list of OAuth2 clients.
func (c *Client) BuildGetAPIClientsRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(ctx, filter.ToValues(), apiClientsBasePath)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetAPIClients gets a list of OAuth2 clients.
func (c *Client) GetAPIClients(ctx context.Context, filter *types.QueryFilter) (*types.APIClientList, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildGetAPIClientsRequest(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	var apiClients *types.APIClientList
	err = c.retrieve(ctx, req, &apiClients)

	return apiClients, err
}

// BuildCreateAPIClientRequest builds an HTTP request for creating OAuth2 clients.
func (c *Client) BuildCreateAPIClientRequest(
	ctx context.Context,
	cookie *http.Cookie,
	input *types.APICientCreationInput,
) (*http.Request, error) {
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

// CreateAPIClient creates an OAuth2 client. Note that cookie must not be nil
// in order to receive a valid response.
func (c *Client) CreateAPIClient(
	ctx context.Context,
	cookie *http.Cookie,
	input *types.APICientCreationInput,
) (*types.APIClientCreationResponse, error) {
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

	req, err := c.BuildCreateAPIClientRequest(ctx, cookie, input)
	if err != nil {
		return nil, err
	}

	if resErr := c.executeRequest(ctx, req, &apiClientResponse); resErr != nil {
		return nil, fmt.Errorf("executing request: %w", resErr)
	}

	return apiClientResponse, nil
}

// BuildArchiveAPIClientRequest builds an HTTP request for archiving an oauth2 client.
func (c *Client) BuildArchiveAPIClientRequest(ctx context.Context, id uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if id == 0 {
		return nil, ErrZeroIDProvided
	}

	uri := c.BuildURL(
		ctx,
		nil,
		apiClientsBasePath,
		strconv.FormatUint(id, 10),
	)

	return http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
}

// ArchiveAPIClient archives an OAuth2 client.
func (c *Client) ArchiveAPIClient(ctx context.Context, id uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if id == 0 {
		return ErrZeroIDProvided
	}

	req, err := c.BuildArchiveAPIClientRequest(ctx, id)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// BuildGetAuditLogForAPIClientRequest builds an HTTP request for fetching a list of audit log entries for an oauth2 client.
func (c *Client) BuildGetAuditLogForAPIClientRequest(ctx context.Context, clientID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if clientID == 0 {
		return nil, ErrZeroIDProvided
	}

	uri := c.BuildURL(ctx, nil, apiClientsBasePath, strconv.FormatUint(clientID, 10), "audit")
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetAuditLogForAPIClient retrieves a list of audit log entries pertaining to an oauth2 client.
func (c *Client) GetAuditLogForAPIClient(ctx context.Context, clientID uint64) (entries []*types.AuditLogEntry, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if clientID == 0 {
		return nil, ErrZeroIDProvided
	}

	req, err := c.BuildGetAuditLogForAPIClientRequest(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &entries); retrieveErr != nil {
		return nil, retrieveErr
	}

	return entries, nil
}
