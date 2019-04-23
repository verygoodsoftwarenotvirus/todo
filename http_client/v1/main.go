package client

import (
	"context"
	"fmt"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1/noop"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"

	"github.com/gorilla/websocket"
	"github.com/moul/http2curl"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	// ErrNotFound is a handy error to return when we receive a 404 response
	ErrNotFound = errors.New("404: not found")
)

// V1Client is a client for interacting with v1 of our API
type V1Client struct {
	plainClient       *http.Client
	authedClient      *http.Client
	currentUserCookie *http.Cookie

	logger logging.Logger
	tracer opentracing.Tracer
	Debug  bool
	URL    *url.URL

	// new hotness
	clientID     string
	clientSecret string
	Scopes       []string
	redirectURI  string
}

// AuthenticatedClient returns the authenticated *http.Client that we use to make most requests
func (c *V1Client) AuthenticatedClient() *http.Client {
	return c.authedClient
}

// PlainClient returns the unauthenticated *http.Client that we use to make certain requests
func (c *V1Client) PlainClient() *http.Client {
	return c.plainClient
}

func tokenEndpoint(baseURL *url.URL) oauth2.Endpoint {
	tu, au := *baseURL, *baseURL
	tu.Path, au.Path = "oauth2/token", "oauth2/authorize"

	return oauth2.Endpoint{
		TokenURL: tu.String(),
		AuthURL:  au.String(),
	}
}

func buildOAuthClient(uri *url.URL, clientID, clientSecret string) *http.Client {
	conf := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"*"}, // SECUREME
		EndpointParams: url.Values{
			"client_id":     []string{clientID},
			"client_secret": []string{clientSecret},
		},
		TokenURL: tokenEndpoint(uri).TokenURL,
	}

	client := &http.Client{
		Transport: &oauth2.Transport{
			Base:   &nethttp.Transport{},
			Source: oauth2.ReuseTokenSource(nil, conf.TokenSource(context.Background())),
		},
		Timeout: 5 * time.Second,
	}

	return client
}

// NewClient builds a new API client for us
func NewClient(
	clientID,
	clientSecret string,
	address *url.URL,
	logger logging.Logger,
	hclient *http.Client,
	tracer opentracing.Tracer,
	debug bool,
) (*V1Client, error) {
	var client = hclient
	if client == nil {
		client = &http.Client{
			Timeout: 5 * time.Second,
		}
	}

	if tracer == nil {
		tracer = &opentracing.NoopTracer{}
	}

	if debug {
		logger.SetLevel(logging.DebugLevel)
		logger.Debug("level set to debug!")
	}

	c := &V1Client{
		URL:          address,
		plainClient:  client,
		logger:       logger.WithName("v1_client"),
		tracer:       tracer,
		Debug:        debug,
		authedClient: buildOAuthClient(address, clientID, clientSecret),
	}

	logger.WithValue("url", address.String()).Debug("returning client")
	return c, nil
}

// NewSimpleClient is a client that is capable of much less than the normal client
// and has noops or empty values for most of its authentication and debug parts.
// Its purpose at the time of this writing is merely so I can make users (which
// is a route that doesn't require authentication)
func NewSimpleClient(address *url.URL, debug bool) (*V1Client, error) {
	l := noop.ProvideNoopLogger()
	h := &http.Client{Timeout: 5 * time.Second}
	t := tracing.ProvideNoopTracer()
	c, err := NewClient("", "", address, l, h, t, debug)
	return c, err
}

func (c *V1Client) executeRequest(ctx context.Context, client *http.Client, req *http.Request) (*http.Response, error) {
	var logger = c.logger
	if command, err := http2curl.GetCurlCommand(req); err == nil {
		logger = c.logger.WithValue("curl", command.String())
	}

	//// attach ClientTrace to the Context, and Context to request
	// err := c.tracer.Inject(
	// 	span.Context(),
	// 	opentracing.HTTPHeaders,
	// 	opentracing.HTTPHeadersCarrier(req.Header),
	// )
	// if err != nil {
	// 	return nil, errors.Wrap(err, "injecting span into headers")
	// }

	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "executing request")
	}

	if c.Debug {
		bdump, err := httputil.DumpResponse(res, true)
		if err == nil && req.Method != http.MethodGet {
			logger = logger.WithValue("response_body", string(bdump))
		}
	}

	logger.Debug("request executed")
	return res, nil
}

// BuildURL builds URLs
func (c *V1Client) BuildURL(qp url.Values, parts ...string) string {
	if qp != nil {
		return c.buildURL(qp, parts...).String()
	}
	return c.buildURL(nil, parts...).String()
}

func (c *V1Client) buildURL(queryParams url.Values, parts ...string) *url.URL {
	tu := *c.URL

	parts = append([]string{"api", "v1"}, parts...)
	u, _ := url.Parse(strings.Join(parts, "/"))

	if queryParams != nil {
		u.RawQuery = queryParams.Encode()
	}

	return tu.ResolveReference(u)
}

// BuildWebsocketURL builds websocket URLs
func (c *V1Client) BuildWebsocketURL(parts ...string) string {
	u := c.buildURL(nil, parts...)
	u.Scheme = "ws"

	return u.String()
}

// IsUp returns whether or not the service is healthy
func (c *V1Client) IsUp() bool {
	u := *c.URL
	uri := fmt.Sprintf("%s://%s:%s/_meta_/health", u.Scheme, u.Host, u.Port())

	logger := c.logger.WithValue("health_check_url", uri)

	req, _ := http.NewRequest(http.MethodGet, uri, nil)
	res, err := c.plainClient.Do(req)

	if err != nil {
		logger.Error(err, "health check")
		return false
	}

	logger.WithValue("status_code", res.StatusCode)

	return res.StatusCode == http.StatusOK
}

func (c *V1Client) buildDataRequest(method, uri string, in interface{}) (*http.Request, error) {
	body, err := createBodyFromStruct(in)
	if err != nil {
		return nil, err
	}

	return http.NewRequest(method, uri, body)
}

func (c *V1Client) makeDataRequest(method string, uri string, in interface{}, out interface{}) error {
	// sometimes we want to make requests with data attached, but we don't really care about the response
	// so we give this function a nil `out` value. That said, if you provide us a value, it needs to be a pointer.
	if out != nil {
		if np, err := argIsNotPointer(out); np || err != nil {
			return errors.Wrap(err, "struct to load must be a pointer")
		}
	}

	req, err := c.buildDataRequest(method, uri, in)
	if err != nil {
		return errors.Wrap(err, "encountered error building request")
	}

	res, err := c.executeRequest(context.Background(), c.authedClient, req)
	if err != nil {
		return errors.Wrap(err, "encountered error executing request")
	}

	if res.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}

	if out != nil {
		resErr := unmarshalBody(res, &out)
		if resErr != nil {
			return errors.Wrap(err, "encountered error loading response from server")
		}
		c.logger.WithValue("loaded_value", out).Debug("data request returned")
	}

	return nil
}

func (c *V1Client) makeUnauthedDataRequest(method string, uri string, in interface{}, out interface{}) error {
	// sometimes we want to make requests with data attached, but we don't really care about the response
	// so we give this function a nil `out` value. That said, if you provide us a value, it needs to be a pointer.
	if out != nil {
		if np, err := argIsNotPointer(out); np || err != nil {
			return errors.Wrap(err, "struct to load must be a pointer")
		}
	}

	req, err := c.buildDataRequest(method, uri, in)
	if err != nil {
		return errors.Wrap(err, "encountered error building request")
	}

	res, err := c.executeRequest(context.Background(), c.plainClient, req)
	if err != nil {
		return errors.Wrap(err, "encountered error executing request")
	}

	if res.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}

	if out != nil {
		resErr := unmarshalBody(res, &out)
		if resErr != nil {
			return errors.Wrap(err, "encountered error loading response from server")
		}
		c.logger.WithValue("loaded_value", out).Debug("unauthenticated data request returned")
	}

	return nil
}

func (c *V1Client) exists(ctx context.Context, uri string) (bool, error) {
	req, _ := http.NewRequest(http.MethodHead, uri, nil)
	res, err := c.executeRequest(context.Background(), c.authedClient, req)
	if err != nil {
		return false, errors.Wrap(err, "encountered error executing request")
	}

	return res.StatusCode == http.StatusOK, nil
}

func (c *V1Client) get(ctx context.Context, uri string, obj interface{}) error {
	ce := &Error{}

	if err := argIsNotPointerOrNil(obj); err != nil {
		ce.Err = errors.Wrap(err, "struct to load must be a pointer")
		return ce
	}

	req, _ := http.NewRequest(http.MethodGet, uri, nil)
	res, err := c.executeRequest(context.Background(), c.authedClient, req)
	if err != nil {
		ce.Err = errors.Wrap(err, "encountered error executing request")
		return ce
	}

	if res.StatusCode == http.StatusNotFound {
		ce.Err = errors.New("404 Not Found")
		return ce
	}

	return unmarshalBody(res, &obj)
}

func (c *V1Client) delete(ctx context.Context, uri string) error {
	req, _ := http.NewRequest(http.MethodDelete, uri, nil)
	res, err := c.executeRequest(context.Background(), c.authedClient, req)
	if err != nil {
		return &Error{Err: err}
	} else if res.StatusCode != http.StatusNoContent {
		return &Error{Err: errors.New(fmt.Sprintf("status returned: %d", res.StatusCode))}
	}

	return nil
}

func (c *V1Client) post(ctx context.Context, uri string, in interface{}, out interface{}) error {
	return c.makeDataRequest(http.MethodPost, uri, in, out)
}

// This function is, at the time of this writing, exclusively for the use of creating users
func (c *V1Client) postPlain(ctx context.Context, uri string, in interface{}, out interface{}) error {
	return c.makeUnauthedDataRequest(http.MethodPost, uri, in, out)
}

func (c *V1Client) put(ctx context.Context, uri string, in interface{}, out interface{}) error {
	return c.makeDataRequest(http.MethodPut, uri, in, out)
}

// DialWebsocket dials a websocket
func (c *V1Client) DialWebsocket(ctx context.Context, fq *FeedQuery) (*websocket.Conn, error) {
	u := c.buildURL(fq.Values(), "event_feed")
	u.Scheme = "wss"

	logger := c.logger.WithValues(map[string]interface{}{
		"feed_query":    fq,
		"websocket_url": u.String(),
	})

	if fq == nil {
		return nil, errors.New("Valid feed query required")
	}

	if !c.IsUp() {
		logger.Debug("websocket service is down")
		return nil, errors.New("service is down")
	}

	dialer := websocket.DefaultDialer
	logger.Debug("connecting to websocket")

	conn, res, err := dialer.Dial(u.String(), nil)
	if err != nil {
		logger.Debug("encountered error dialing websocket")
		return nil, err
	}

	if res.StatusCode < http.StatusBadRequest {
		return conn, nil
	}
	return nil, fmt.Errorf("encountered status code: %d when trying to reach websocket", res.StatusCode)
}