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
	oauth2BasePath        = "oauth2"
	oauth2ClientBasePath  = "client"
	oauth2ClientsBasePath = "clients"
)

// BuildGetOAuth2ClientRequest builds an HTTP request for fetching an OAuth2 client.
func (c *V1Client) BuildGetOAuth2ClientRequest(ctx context.Context, id uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(
		nil,
		oauth2BasePath,
		oauth2ClientsBasePath,
		strconv.FormatUint(id, 10),
	)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetOAuth2Client gets an OAuth2 client.
func (c *V1Client) GetOAuth2Client(ctx context.Context, id uint64) (oauth2Client *types.OAuth2Client, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildGetOAuth2ClientRequest(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.retrieve(ctx, req, &oauth2Client)

	return oauth2Client, err
}

// BuildGetOAuth2ClientsRequest builds an HTTP request for fetching a list of OAuth2 clients.
func (c *V1Client) BuildGetOAuth2ClientsRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(
		filter.ToValues(),
		oauth2BasePath,
		oauth2ClientsBasePath,
	)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetOAuth2Clients gets a list of OAuth2 clients.
func (c *V1Client) GetOAuth2Clients(ctx context.Context, filter *types.QueryFilter) (*types.OAuth2ClientList, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildGetOAuth2ClientsRequest(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	var oauth2Clients *types.OAuth2ClientList
	err = c.retrieve(ctx, req, &oauth2Clients)

	return oauth2Clients, err
}

// BuildCreateOAuth2ClientRequest builds an HTTP request for creating OAuth2 clients.
func (c *V1Client) BuildCreateOAuth2ClientRequest(
	ctx context.Context,
	cookie *http.Cookie,
	body *types.OAuth2ClientCreationInput,
) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.buildVersionlessURL(
		nil,
		oauth2BasePath,
		oauth2ClientBasePath,
	)

	req, err := c.buildDataRequest(ctx, http.MethodPost, uri, body)
	if err != nil {
		return nil, err
	}

	req.AddCookie(cookie)

	return req, nil
}

// CreateOAuth2Client creates an OAuth2 client. Note that cookie must not be nil
// in order to receive a valid response.
func (c *V1Client) CreateOAuth2Client(
	ctx context.Context,
	cookie *http.Cookie,
	input *types.OAuth2ClientCreationInput,
) (*types.OAuth2Client, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	var oauth2Client *types.OAuth2Client

	if cookie == nil {
		return nil, errors.New("cookie required for request")
	}

	req, err := c.BuildCreateOAuth2ClientRequest(ctx, cookie, input)
	if err != nil {
		return nil, err
	}

	if resErr := c.executeUnauthenticatedDataRequest(ctx, req, &oauth2Client); resErr != nil {
		return nil, fmt.Errorf("executing request: %w", resErr)
	}

	return oauth2Client, nil
}

// BuildArchiveOAuth2ClientRequest builds an HTTP request for archiving an oauth2 client.
func (c *V1Client) BuildArchiveOAuth2ClientRequest(ctx context.Context, id uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(
		nil,
		oauth2BasePath,
		oauth2ClientsBasePath,
		strconv.FormatUint(id, 10),
	)

	return http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
}

// ArchiveOAuth2Client archives an OAuth2 client.
func (c *V1Client) ArchiveOAuth2Client(ctx context.Context, id uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildArchiveOAuth2ClientRequest(ctx, id)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// BuildGetAuditLogForOAuth2ClientRequest builds an HTTP request for fetching a list of audit log entries for an oauth2 client.
func (c *V1Client) BuildGetAuditLogForOAuth2ClientRequest(ctx context.Context, clientID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(
		nil,
		oauth2BasePath,
		oauth2ClientsBasePath,
		strconv.FormatUint(clientID, 10),
		"audit",
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetAuditLogForOAuth2Client retrieves a list of audit log entries pertaining to an oauth2 client.
func (c *V1Client) GetAuditLogForOAuth2Client(ctx context.Context, clientID uint64) (entries []types.AuditLogEntry, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildGetAuditLogForOAuth2ClientRequest(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &entries); retrieveErr != nil {
		return nil, retrieveErr
	}

	return entries, nil
}
