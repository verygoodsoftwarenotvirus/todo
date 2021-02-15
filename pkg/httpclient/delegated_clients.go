package httpclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	delegatedClientBasePath  = "delegated_client"
	delegatedClientsBasePath = "delegated_clients"
)

// BuildGetDelegatedClientRequest builds an HTTP request for fetching an OAuth2 client.
func (c *Client) BuildGetDelegatedClientRequest(ctx context.Context, id uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(
		nil,
		delegatedClientsBasePath,
		strconv.FormatUint(id, 10),
	)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetDelegatedClient gets an OAuth2 client.
func (c *Client) GetDelegatedClient(ctx context.Context, id uint64) (delegatedClient *types.DelegatedClient, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildGetDelegatedClientRequest(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.retrieve(ctx, req, &delegatedClient)

	return delegatedClient, err
}

// BuildGetDelegatedClientsRequest builds an HTTP request for fetching a list of OAuth2 clients.
func (c *Client) BuildGetDelegatedClientsRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(
		filter.ToValues(),
		delegatedClientsBasePath,
	)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetDelegatedClients gets a list of OAuth2 clients.
func (c *Client) GetDelegatedClients(ctx context.Context, filter *types.QueryFilter) (*types.DelegatedClientList, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildGetDelegatedClientsRequest(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	var delegatedClients *types.DelegatedClientList
	err = c.retrieve(ctx, req, &delegatedClients)

	return delegatedClients, err
}

// BuildCreateDelegatedClientRequest builds an HTTP request for creating OAuth2 clients.
func (c *Client) BuildCreateDelegatedClientRequest(
	ctx context.Context,
	cookie *http.Cookie,
	body *types.DelegatedClientCreationInput,
) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.buildVersionlessURL(
		nil,
		delegatedClientBasePath,
	)

	req, err := c.buildDataRequest(ctx, http.MethodPost, uri, body)
	if err != nil {
		return nil, err
	}

	req.AddCookie(cookie)

	return req, nil
}

// CreateDelegatedClient creates an OAuth2 client. Note that cookie must not be nil
// in order to receive a valid response.
func (c *Client) CreateDelegatedClient(
	ctx context.Context,
	cookie *http.Cookie,
	input *types.DelegatedClientCreationInput,
) (*types.DelegatedClient, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	var delegatedClient *types.DelegatedClient

	if cookie == nil {
		return nil, errors.New("cookie required for request")
	}

	req, err := c.BuildCreateDelegatedClientRequest(ctx, cookie, input)
	if err != nil {
		return nil, err
	}

	if resErr := c.executeUnauthenticatedDataRequest(ctx, req, &delegatedClient); resErr != nil {
		return nil, fmt.Errorf("executing request: %w", resErr)
	}

	return delegatedClient, nil
}

// BuildArchiveDelegatedClientRequest builds an HTTP request for archiving an oauth2 client.
func (c *Client) BuildArchiveDelegatedClientRequest(ctx context.Context, id uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(
		nil,
		delegatedClientsBasePath,
		strconv.FormatUint(id, 10),
	)

	return http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
}

// ArchiveDelegatedClient archives an OAuth2 client.
func (c *Client) ArchiveDelegatedClient(ctx context.Context, id uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildArchiveDelegatedClientRequest(ctx, id)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// BuildGetAuditLogForDelegatedClientRequest builds an HTTP request for fetching a list of audit log entries for an oauth2 client.
func (c *Client) BuildGetAuditLogForDelegatedClientRequest(ctx context.Context, clientID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(
		nil,
		delegatedClientsBasePath,
		strconv.FormatUint(clientID, 10),
		"audit",
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetAuditLogForDelegatedClient retrieves a list of audit log entries pertaining to an oauth2 client.
func (c *Client) GetAuditLogForDelegatedClient(ctx context.Context, clientID uint64) (entries []*types.AuditLogEntry, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildGetAuditLogForDelegatedClientRequest(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &entries); retrieveErr != nil {
		return nil, retrieveErr
	}

	return entries, nil
}
