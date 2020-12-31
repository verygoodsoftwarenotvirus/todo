package httpclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"

	"github.com/moul/http2curl"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
	"golang.org/x/oauth2"
)

const (
	defaultTimeout = 30 * time.Second
	clientName     = "todo_client_v1"
)

var (
	// ErrNotFound is a handy error to return when we receive a 404 response.
	ErrNotFound = fmt.Errorf("%d: not found", http.StatusNotFound)

	// ErrUnauthorized is a handy error to return when we receive a 401 response.
	ErrUnauthorized = fmt.Errorf("%d: not authorized", http.StatusUnauthorized)

	// ErrInvalidTOTPToken is an error for when our TOTP validation request goes awry.
	ErrInvalidTOTPToken = errors.New("invalid TOTP token")

	// ErrNilInputProvided indicates nil input was provided in an unacceptable context.
	ErrNilInputProvided = errors.New("nil input provided")
)

// Client is a client for interacting with v1 of our HTTP API.
type Client struct {
	url          *url.URL
	plainClient  *http.Client
	authedClient *http.Client

	logger logging.Logger
	tracer tracing.Tracer

	debug     bool
	adminMode bool

	authMode    *authMode
	authCookie  *http.Cookie
	tokenSource oauth2.TokenSource
}

// EnableAdminMode enables admin mode.
func (c *Client) EnableAdminMode() {
	c.adminMode = true
}

// AuthenticatedClient returns the authenticated *http.Client that we use to make most requests.
func (c *Client) AuthenticatedClient() *http.Client {
	return c.authedClient
}

// PlainClient returns the unauthenticated *http.Client that we use to make certain requests.
func (c *Client) PlainClient() *http.Client {
	return c.plainClient
}

// TokenSource provides the client's token source.
func (c *Client) TokenSource() oauth2.TokenSource {
	return c.tokenSource
}

// TokenSource provides the client's token source.
func (c *Client) URL() *url.URL {
	return c.url
}

// tokenEndpoint provides the oauth2 Endpoint for a given host.
func tokenEndpoint(baseURL *url.URL) oauth2.Endpoint {
	tu, au := *baseURL, *baseURL
	tu.Path, au.Path = "oauth2/token", "oauth2/authorize"

	return oauth2.Endpoint{
		TokenURL: tu.String(),
		AuthURL:  au.String(),
	}
}

// NewClient builds a new API client for us.
func NewClient(options ...option) *Client {
	c := &Client{
		url:          MustParseURL(""),
		tokenSource:  nil,
		authedClient: http.DefaultClient,
		plainClient:  http.DefaultClient,
		debug:        false,
		logger:       noop.NewLogger(),
		tracer:       tracing.NewTracer(clientName),
	}

	for _, opt := range options {
		opt(c)
	}

	return c
}

// closeResponseBody takes a given HTTP response and closes its body, logging if an error occurs.
func (c *Client) closeResponseBody(res *http.Response) {
	if res != nil {
		if err := res.Body.Close(); err != nil {
			c.logger.Error(err, "closing response body")
		}
	}
}

// BuildURL builds standard service URLs.
func (c *Client) BuildURL(qp url.Values, parts ...string) string {
	if u := c.buildRawURL(qp, parts...); u != nil {
		return u.String()
	}

	return ""
}

// buildRawURL takes a given set of query parameters and url parts, and returns.
// a parsed url object from them.
func (c *Client) buildRawURL(queryParams url.Values, parts ...string) *url.URL {
	tu := *c.url

	parts = append([]string{"api", "v1"}, parts...)

	u, err := url.Parse(strings.Join(parts, "/"))
	if err != nil {
		c.logger.Error(err, "building url")
		return nil
	}

	if queryParams != nil {
		u.RawQuery = queryParams.Encode()
	}

	return tu.ResolveReference(u)
}

// buildVersionlessURL builds a url without the `/api/v1/` prefix. It should
// otherwise be identical to buildRawURL.
func (c *Client) buildVersionlessURL(qp url.Values, parts ...string) string {
	tu := *c.url

	u, err := url.Parse(path.Join(parts...))
	if err != nil {
		c.logger.Error(err, "building url")
		return ""
	}

	if qp != nil {
		u.RawQuery = qp.Encode()
	}

	return tu.ResolveReference(u).String()
}

// BuildWebsocketURL builds a standard url and then converts its scheme to the websocket protocol.
func (c *Client) BuildWebsocketURL(parts ...string) string {
	u := c.buildRawURL(nil, parts...)
	u.Scheme = "ws"

	return u.String()
}

// BuildHealthCheckRequest builds a health check HTTP request.
func (c *Client) BuildHealthCheckRequest(ctx context.Context) (*http.Request, error) {
	u := *c.url
	uri := fmt.Sprintf("%s://%s/_meta_/ready", u.Scheme, u.Host)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// IsUp returns whether or not the service's health endpoint is returning 200s.
func (c *Client) IsUp(ctx context.Context) bool {
	req, err := c.BuildHealthCheckRequest(ctx)
	if err != nil {
		c.logger.Error(err, "building request")
		return false
	}

	res, err := c.plainClient.Do(req)
	if err != nil {
		c.logger.Error(err, "health check")
		return false
	}

	c.closeResponseBody(res)

	return res.StatusCode == http.StatusOK
}

// buildDataRequest builds an HTTP request for a given method, url, and body data.
func (c *Client) buildDataRequest(ctx context.Context, method, uri string, in interface{}) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	body, err := createBodyFromStruct(in)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, uri, body)
	if err != nil {
		return nil, err
	}

	if c.authMode == cookieAuthMode && c.authCookie != nil {
		req.AddCookie(c.authCookie)
	}

	req.Header.Set("Content-type", "application/json")

	return req, nil
}

// executeRequest takes a given request and executes it with the auth client. It returns some errors
// upon receiving certain status codes, but otherwise will return nil upon success.
func (c *Client) executeRequest(ctx context.Context, req *http.Request, out interface{}) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithRequest(req)
	logger.Debug("executing request")

	// ensure cookie is attached
	if c.authMode == cookieAuthMode && c.authCookie != nil {
		if _, err := req.Cookie(c.authCookie.Name); err == nil {
			req.AddCookie(c.authCookie)
		}
	}

	res, err := c.executeRawRequest(ctx, c.authedClient, req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}

	switch res.StatusCode {
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusUnauthorized:
		return ErrUnauthorized
	}

	logger.WithValue(keys.ResponseStatusKey, res.StatusCode).Debug("request executed")

	if out != nil {
		if resErr := c.unmarshalBody(ctx, res, out); resErr != nil {
			return fmt.Errorf("loading %s %d response from server: %w", res.Request.Method, res.StatusCode, resErr)
		}
	}

	return nil
}

// executeRawRequest takes a given *http.Request and executes it with the provided.
// client, alongside some debugging logging.
func (c *Client) executeRawRequest(ctx context.Context, client *http.Client, req *http.Request) (*http.Response, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithRequest(req)
	logger.Debug("executing request")

	// ensure cookie is attached
	if c.authMode == cookieAuthMode && c.authCookie != nil {
		if _, err := req.Cookie(c.authCookie.Name); err == nil {
			req.AddCookie(c.authCookie)
		}
	}

	if command, err := http2curl.GetCurlCommand(req); err == nil && c.debug {
		logger = c.logger.WithValue("curl", command.String())
	}

	res, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	if bdump, resDumpErr := httputil.DumpResponse(res, true); resDumpErr == nil {
		logger = logger.WithValue("response_body", string(bdump))
	}

	logger.WithValue(keys.ResponseStatusKey, res.StatusCode).Debug("request executed")

	return res, nil
}

// checkExistence executes an HTTP request and loads the response content into a bool.
func (c *Client) checkExistence(ctx context.Context, req *http.Request) (bool, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	res, err := c.executeRawRequest(ctx, c.authedClient, req)
	if err != nil {
		return false, err
	}

	c.closeResponseBody(res)

	return res.StatusCode == http.StatusOK, nil
}

// retrieve executes an HTTP request and loads the response content into a struct. In the event of a 404,
// the provided ErrNotFound is returned.
func (c *Client) retrieve(ctx context.Context, req *http.Request, obj interface{}) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if err := argIsNotPointerOrNil(obj); err != nil {
		return fmt.Errorf("struct to load must be a pointer: %w", err)
	}

	res, err := c.executeRawRequest(ctx, c.authedClient, req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}

	if resErr := errorFromResponse(res); resErr != nil {
		return resErr
	}

	return c.unmarshalBody(ctx, res, &obj)
}

// executeUnauthenticatedDataRequest takes a given request and loads the response into an interface value.
func (c *Client) executeUnauthenticatedDataRequest(ctx context.Context, req *http.Request, out interface{}) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	res, err := c.executeRawRequest(ctx, c.plainClient, req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}

	if resErr := errorFromResponse(res); resErr != nil {
		return resErr
	}

	if out != nil {
		if resErr := c.unmarshalBody(ctx, res, out); resErr != nil {
			return fmt.Errorf("loading %s %d response from server: %w", res.Request.Method, res.StatusCode, err)
		}
	}

	return nil
}
