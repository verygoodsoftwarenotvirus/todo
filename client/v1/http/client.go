package client

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

	"github.com/moul/http2curl"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/noop"
	"go.opencensus.io/plugin/ochttp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	defaultTimeout = 5 * time.Second
	clientName     = "v1_client"
)

var (
	// ErrNotFound is a handy error to return when we receive a 404 response
	ErrNotFound = errors.New("404: not found")

	// ErrUnauthorized is a handy error to return when we receive a 404 response
	ErrUnauthorized = errors.New("401: not authorized")
)

// V1Client is a client for interacting with v1 of our REST API
type V1Client struct {
	plainClient  *http.Client
	authedClient *http.Client
	logger       logging.Logger
	Debug        bool
	URL          *url.URL
	Scopes       []string
	tokenSource  oauth2.TokenSource
}

// AuthenticatedClient returns the authenticated *http.Client that we use to make most requests
func (c *V1Client) AuthenticatedClient() *http.Client {
	return c.authedClient
}

// PlainClient returns the unauthenticated *http.Client that we use to make certain requests
func (c *V1Client) PlainClient() *http.Client {
	return c.plainClient
}

// TokenSource provides the client's token source
func (c *V1Client) TokenSource() oauth2.TokenSource {
	return c.tokenSource
}

// NewClient builds a new API client for us
func NewClient(
	ctx context.Context,
	clientID,
	clientSecret string,
	address *url.URL,
	logger logging.Logger,
	hclient *http.Client,
	scopes []string,
	debug bool,
) (*V1Client, error) {
	var client = hclient
	if client == nil {
		client = &http.Client{
			Timeout: defaultTimeout,
		}
	}
	if client.Timeout == 0 {
		client.Timeout = defaultTimeout
	}

	if debug {
		logger.SetLevel(logging.DebugLevel)
		logger.Debug("log level set to debug!")
	}

	ac, ts := buildOAuthClient(ctx, address, clientID, clientSecret, scopes)

	c := &V1Client{
		URL:          address,
		plainClient:  client,
		logger:       logger.WithName(clientName),
		Debug:        debug,
		authedClient: ac,
		tokenSource:  ts,
	}

	logger.WithValue("url", address.String()).Debug("returning client")
	return c, nil
}

// buildOAuthClient does too much
func buildOAuthClient(
	ctx context.Context,
	uri *url.URL,
	clientID,
	clientSecret string,
	scopes []string,
) (*http.Client, oauth2.TokenSource) {
	conf := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		EndpointParams: url.Values{
			"client_id":     []string{clientID},
			"client_secret": []string{clientSecret},
		},
		TokenURL: tokenEndpoint(uri).TokenURL,
	}

	ts := oauth2.ReuseTokenSource(nil, conf.TokenSource(ctx))
	client := &http.Client{
		Transport: &oauth2.Transport{
			Base: &ochttp.Transport{
				Base: newDefaultRoundTripper(),
			},
			Source: ts,
		},
		Timeout: 5 * time.Second,
	}

	return client, ts
}

// tokenEndpoint provides the oauth2 Endpoint for a given host
func tokenEndpoint(baseURL *url.URL) oauth2.Endpoint {
	tu, au := *baseURL, *baseURL
	tu.Path, au.Path = "oauth2/token", "oauth2/authorize"

	return oauth2.Endpoint{
		TokenURL: tu.String(),
		AuthURL:  au.String(),
	}
}

// NewSimpleClient is a client that is capable of much less than the normal client
// and has noops or empty values for most of its authentication and debug parts.
// Its purpose at the time of this writing is merely so I can make users (which
// is a route that doesn't require authentication)
func NewSimpleClient(ctx context.Context, address *url.URL, debug bool) (*V1Client, error) {
	l := noop.ProvideNoopLogger()
	h := &http.Client{Timeout: 5 * time.Second}
	c, err := NewClient(ctx, "", "", address, l, h, []string{"*"}, debug)
	return c, err
}

// executeRawRequest takes a given *http.Request and executes it with the provided
// client, alongside some debugging logging.
func (c *V1Client) executeRawRequest(ctx context.Context, client *http.Client, req *http.Request) (*http.Response, error) {
	var logger = c.logger
	if command, err := http2curl.GetCurlCommand(req); err == nil && c.Debug {
		logger = c.logger.WithValue("curl", command.String())
	}

	res, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	if c.Debug {
		bdump, err := httputil.DumpResponse(res, true)
		if err == nil && req.Method != http.MethodGet {
			logger = logger.WithValue("response_body", string(bdump))
		}
		logger.Debug("request executed")
	}

	return res, nil
}

// BuildURL builds standard service URLs
func (c *V1Client) BuildURL(qp url.Values, parts ...string) string {
	if qp != nil {
		return c.buildURL(qp, parts...).String()
	}
	return c.buildURL(nil, parts...).String()
}

// buildURL takes a given set of query parameters and URL parts, and returns
// a parsed URL object from them.
func (c *V1Client) buildURL(queryParams url.Values, parts ...string) *url.URL {
	tu := *c.URL

	parts = append([]string{"api", "v1"}, parts...)
	u, err := url.Parse(strings.Join(parts, "/"))
	if err != nil {
		panic(fmt.Sprintf("was asked to build an invalid URL: %v", err))
	}

	if queryParams != nil {
		u.RawQuery = queryParams.Encode()
	}

	return tu.ResolveReference(u)
}

// buildVersionlessURL builds a URL without the `/api/v1/` prefix. It should
// otherwise be identical to buildURL
func (c *V1Client) buildVersionlessURL(qp url.Values, parts ...string) string {
	tu := *c.URL

	u, err := url.Parse(path.Join(parts...))
	if err != nil {
		panic(fmt.Sprintf("user tried to build an invalid URL: %v", err))
	}

	if qp != nil {
		u.RawQuery = qp.Encode()
	}

	return tu.ResolveReference(u).String()
}

// BuildWebsocketURL builds a standard URL and then converts its scheme to the websocket protocol
func (c *V1Client) BuildWebsocketURL(parts ...string) string {
	u := c.buildURL(nil, parts...)
	u.Scheme = "ws"

	return u.String()
}

// BuildHealthCheckRequest builds a health check HTTP Request
func (c *V1Client) BuildHealthCheckRequest() (*http.Request, error) {
	u := *c.URL
	uri := fmt.Sprintf("%s://%s/_meta_/ready", u.Scheme, u.Host)

	return http.NewRequest(http.MethodGet, uri, nil)
}

// IsUp returns whether or not the service's health endpoint is returning 200s
func (c *V1Client) IsUp() bool {
	req, err := c.BuildHealthCheckRequest()
	if err != nil {
		c.logger.Error(err, "building request")
		return false
	}

	res, err := c.plainClient.Do(req)
	if err != nil {
		c.logger.Error(err, "health check")
		return false
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			c.logger.Error(err, "closing response body")
		}
	}()

	return res.StatusCode == http.StatusOK
}

// buildDataRequest builds an HTTP request for a given method, URL, and body data.
func (c *V1Client) buildDataRequest(method, uri string, in interface{}) (*http.Request, error) {
	body, err := createBodyFromStruct(in)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-type", "application/json")
	return req, nil
}

// retrieve executes an HTTP request and loads the response content into a struct
func (c *V1Client) retrieve(ctx context.Context, req *http.Request, obj interface{}) error {
	if err := argIsNotPointerOrNil(obj); err != nil {
		return fmt.Errorf("struct to load must be a pointer: %w", err)
	}

	res, err := c.executeRawRequest(ctx, c.authedClient, req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}

	if res.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}

	return unmarshalBody(res, &obj)
}

// executeRequest takes a given request and executes it with the auth client. It returns some errors
// upon receiving certain status codes, but otherwise will return nil upon success.
func (c *V1Client) executeRequest(ctx context.Context, req *http.Request, out interface{}) error {
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

	if out != nil {
		resErr := unmarshalBody(res, &out)
		if resErr != nil {
			return fmt.Errorf("loading response from server: %w", err)
		}
	}

	return nil
}

// executeUnathenticatedDataRequest takes a given request and loads the response into an interface value.
func (c *V1Client) executeUnathenticatedDataRequest(ctx context.Context, req *http.Request, out interface{}) error {
	// sometimes we want to make requests with data attached, but we don't really care about the response
	// so we give this function a nil `out` value. That said, if you provide us a value, it needs to be a pointer.
	if out != nil {
		if np, err := argIsNotPointer(out); np || err != nil {
			return fmt.Errorf("struct to load must be a pointer: %w", err)
		}
	}

	res, err := c.executeRawRequest(ctx, c.plainClient, req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}

	if res.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}

	if out != nil {
		resErr := unmarshalBody(res, &out)
		if resErr != nil {
			return fmt.Errorf("loading response from server: %w", err)
		}
	}

	return nil
}
