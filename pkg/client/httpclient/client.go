package httpclient

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	panicking "gitlab.com/verygoodsoftwarenotvirus/todo/internal/panicking"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/httpclient/requests"

	"github.com/moul/http2curl"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

const (
	defaultTimeout = 30 * time.Second
	clientName     = "todo_client_v1"
)

type authMethod struct{}

var (
	cookieAuthMethod   = new(authMethod)
	pasetoAuthMethod   = new(authMethod)
	defaultContentType = encoding.ContentTypeJSON
)

// Client is a client for interacting with v1 of our HTTP API.
type Client struct {
	logger                logging.Logger
	tracer                tracing.Tracer
	panicker              panicking.Panicker
	url                   *url.URL
	requestBuilder        *requests.Builder
	encoder               encoding.ClientEncoder
	unauthenticatedClient *http.Client
	authedClient          *http.Client
	authMethod            *authMethod
	accountID             uint64
	debug                 bool
}

// AuthenticatedClient returns the authenticated *httpclient.Client that we use to make most requests.
func (c *Client) AuthenticatedClient() *http.Client {
	return c.authedClient
}

// PlainClient returns the unauthenticated *httpclient.Client that we use to make certain requests.
func (c *Client) PlainClient() *http.Client {
	return c.unauthenticatedClient
}

// URL provides the client's URL.
func (c *Client) URL() *url.URL {
	return c.url
}

// RequestBuilder provides the client's *requests.Builder.
func (c *Client) RequestBuilder() *requests.Builder {
	return c.requestBuilder
}

// NewClient builds a new API client for us.
func NewClient(u *url.URL, options ...option) (*Client, error) {
	l := logging.NewNonOperationalLogger()

	if u == nil {
		return nil, ErrNoURLProvided
	}

	c := &Client{
		url:                   u,
		logger:                l,
		debug:                 false,
		tracer:                tracing.NewTracer(clientName),
		panicker:              panicking.NewProductionPanicker(),
		encoder:               encoding.ProvideClientEncoder(l, encoding.ContentTypeJSON),
		authedClient:          &http.Client{Transport: buildWrappedTransport(defaultTimeout), Timeout: defaultTimeout},
		unauthenticatedClient: &http.Client{Transport: buildWrappedTransport(defaultTimeout), Timeout: defaultTimeout},
	}

	requestBuilder, err := requests.NewBuilder(c.url, c.logger, encoding.ProvideClientEncoder(l, defaultContentType))
	if err != nil {
		return nil, err
	}

	c.requestBuilder = requestBuilder

	for _, opt := range options {
		if optionSetErr := opt(c); optionSetErr != nil {
			return nil, optionSetErr
		}
	}

	return c, nil
}

// closeResponseBody takes a given HTTP response and closes its body, logging if an error occurs.
func (c *Client) closeResponseBody(ctx context.Context, res *http.Response) {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithResponse(res)

	if res != nil {
		if err := res.Body.Close(); err != nil {
			observability.AcknowledgeError(err, logger, span, "closing response body")
		}
	}
}

// loggerWithFilter prepares a logger from the Client logger that has relevant filter information.
func (c *Client) loggerWithFilter(filter *types.QueryFilter) logging.Logger {
	if filter == nil {
		return c.logger.WithValue(keys.FilterIsNilKey, true)
	}

	return c.logger.WithValue(keys.FilterLimitKey, filter.Limit).WithValue(keys.FilterPageKey, filter.Page)
}

// BuildURL builds standard service URLs.
func (c *Client) BuildURL(ctx context.Context, qp url.Values, parts ...string) string {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if u := c.buildRawURL(ctx, qp, parts...); u != nil {
		return u.String()
	}

	return ""
}

// buildRawURL takes a given set of query parameters and url parts, and returns a parsed url object from them.
func (c *Client) buildRawURL(ctx context.Context, queryParams url.Values, parts ...string) *url.URL {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tu := *c.url
	logger := c.logger.WithValue(keys.URLQueryKey, queryParams.Encode())

	u, err := url.Parse(path.Join(append([]string{"api", "v1"}, parts...)...))
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "building URL")
		return nil
	}

	if queryParams != nil {
		u.RawQuery = queryParams.Encode()
	}

	return tu.ResolveReference(u)
}

// buildVersionlessURL builds a url without the `/api/v1/` prefix. It should otherwise be identical to buildRawURL.
func (c *Client) buildVersionlessURL(ctx context.Context, qp url.Values, parts ...string) string {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tu := *c.url

	u, err := url.Parse(path.Join(parts...))
	if err != nil {
		tracing.AttachErrorToSpan(span, "building url", err)
		return ""
	}

	if qp != nil {
		u.RawQuery = qp.Encode()
	}

	return tu.ResolveReference(u).String()
}

// BuildWebsocketURL builds a standard url and then converts its scheme to the websocket protocol.
func (c *Client) BuildWebsocketURL(ctx context.Context, qp url.Values, parts ...string) string {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	u := c.buildRawURL(ctx, qp, parts...)
	u.Scheme = "ws"

	return u.String()
}

// IsUp returns whether the service's health endpoint is returning 200s.
func (c *Client) IsUp(ctx context.Context) bool {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger

	req, err := c.requestBuilder.BuildHealthCheckRequest(ctx)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "building health check request")

		return false
	}

	res, err := c.fetchResponseToRequest(ctx, c.unauthenticatedClient, req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "performing health check")

		return false
	}

	c.closeResponseBody(ctx, res)

	return res.StatusCode == http.StatusOK
}

// fetchResponseToRequest takes a given *http.Request and executes it with the provided.
// client, alongside some debugging logging.
func (c *Client) fetchResponseToRequest(ctx context.Context, client *http.Client, req *http.Request) (*http.Response, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithRequest(req)

	if command, err := http2curl.GetCurlCommand(req); err == nil && c.debug {
		logger = c.logger.WithValue("curl", command.String())
	}

	// this should be the only use of .Do in this package
	res, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "executing request")
	}

	var bdump []byte
	if bdump, err = httputil.DumpResponse(res, true); err == nil {
		logger = logger.WithValue("response_body", string(bdump))
	}

	logger.WithValue(keys.ResponseStatusKey, res.StatusCode).Debug("request executed")

	return res, nil
}

// executeAndUnmarshal executes a request and unmarshalls it to the provided interface.
func (c *Client) executeAndUnmarshal(ctx context.Context, req *http.Request, httpClient *http.Client, out interface{}) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithRequest(req)
	logger.Debug("executing request")

	res, err := c.fetchResponseToRequest(ctx, httpClient, req)
	if err != nil {
		return observability.PrepareError(err, logger, span, "executing request")
	}

	logger.WithValue(keys.ResponseStatusKey, res.StatusCode).Debug("request executed")

	if err = errorFromResponse(res); err != nil {
		return observability.PrepareError(err, logger, span, "executing request")
	}

	if out != nil {
		if err = c.unmarshalBody(ctx, res, out); err != nil {
			return observability.PrepareError(err, logger, span, "loading %s %d response from server", res.Request.Method, res.StatusCode)
		}
	}

	return nil
}

// fetchAndUnmarshal takes a given request and executes it with the auth client.
func (c *Client) fetchAndUnmarshal(ctx context.Context, req *http.Request, out interface{}) error {
	return c.executeAndUnmarshal(ctx, req, c.authedClient, out)
}

// fetchAndUnmarshalWithoutAuthentication takes a given request and executes it with the plain client.
func (c *Client) fetchAndUnmarshalWithoutAuthentication(ctx context.Context, req *http.Request, out interface{}) error {
	return c.executeAndUnmarshal(ctx, req, c.unauthenticatedClient, out)
}

// responseIsOK executes an HTTP request and loads the response content into a bool.
func (c *Client) responseIsOK(ctx context.Context, req *http.Request) (bool, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithRequest(req)

	res, err := c.fetchResponseToRequest(ctx, c.authedClient, req)
	if err != nil {
		return false, observability.PrepareError(err, logger, span, "executing existence request")
	}

	c.closeResponseBody(ctx, res)

	return res.StatusCode == http.StatusOK, nil
}
