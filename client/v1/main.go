package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/gorilla/websocket"
	"github.com/moul/http2curl"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type contextKey int

const (
	keyTracer      contextKey = iota
	defaultTimeout            = 5 * time.Second
)

var (
	// ErrNotFound is a handy error to return when we receive a 404 response
	ErrNotFound = errors.New("404: not found")
)

// V1Client is a client for interacting with v1 of our API
type V1Client struct {
	plainClient  *http.Client
	authedClient *http.Client
	logger       logging.Logger
	tracer       opentracing.Tracer
	Debug        bool
	URL          *url.URL

	// new hotness
	clientID     string
	clientSecret string
	Scopes       []string
	redirectURI  string

	Items     <-chan *models.Item
	itemsChan chan *models.Item
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

func buildOAuthClient(ctx context.Context, uri *url.URL, clientID, clientSecret string) *http.Client {
	conf := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"*"},
		EndpointParams: url.Values{
			"client_id":     []string{clientID},
			"client_secret": []string{clientSecret},
		},
		TokenURL: tokenEndpoint(uri).TokenURL,
	}

	client := &http.Client{
		Transport: &oauth2.Transport{
			Base:   &nethttp.Transport{},
			Source: oauth2.ReuseTokenSource(nil, conf.TokenSource(ctx)),
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
	if clientID == "" || clientSecret == "" {
		return nil, errors.New("Client ID and Client Secret required")
	}

	var client = hclient
	if client == nil {
		client = &http.Client{
			Timeout: 5 * time.Second,
		}
	}

	if debug {
		logger.SetLevel(logging.DebugLevel)
		logger.Debug("level set to debug!")
	}

	c := &V1Client{
		URL:          address,
		plainClient:  client,
		logger:       logger,
		tracer:       tracer,
		Debug:        debug,
		authedClient: buildOAuthClient(context.Background(), address, clientID, clientSecret),
	}

	logger.WithValue("url", address.String()).Debug("returning client")
	return c, nil
}

func (c *V1Client) executeRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	var logger = c.logger
	if command, err := http2curl.GetCurlCommand(req); err == nil {
		logger = c.logger.WithValue("curl", command.String())
	}

	// attach ClientTrace to the Context, and Context to request
	span := tracing.FetchSpanFromContext(ctx, c.tracer, "executeRequest")
	// trace := NewClientTrace(span)
	// ctx = httptrace.WithClientTrace(ctx, trace)
	// req = req.WithContext(ctx)

	// wrap the request in nethttp.TraceRequest
	// req, _ = nethttp.TraceRequest(c.tracer, req, nethttp.ClientTrace(true))

	c.tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))

	res, err := c.authedClient.Do(req)
	if err != nil {
		return nil, err
	}

	if bdump, err := httputil.DumpResponse(res, true); err == nil && req.Method != http.MethodGet {
		logger = logger.WithValue("response_body", string(bdump))
	}

	logger.Debug("executeRequest called")
	return res, nil
}

// Do executes a raw request object
// TODO: find out why this was implemented
func (c *V1Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if c.URL.Hostname() != req.URL.Hostname() {
		return nil, errors.New("request is destined for unknown server")
	}
	return c.executeRequest(context.Background(), req)
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

	res, err := c.executeRequest(context.Background(), req)
	if err != nil {
		return errors.Wrap(err, "encountered error executing request")
	}

	if res.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}

	rb, err := httputil.DumpResponse(res, true)
	c.logger.WithValue("response_body", rb).Debug("")

	if out != nil {
		resErr := unmarshalBody(res, &out)
		if resErr != nil {
			return errors.Wrap(err, "encountered error loading response from server")
		}
		c.logger.WithValue("loaded_value", out).Debug("data request returned")
	}

	return nil
}

func (c *V1Client) exists(ctx context.Context, uri string) (bool, error) {
	req, _ := http.NewRequest(http.MethodHead, uri, nil)
	res, err := c.executeRequest(context.Background(), req)
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
	res, err := c.executeRequest(context.Background(), req)
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
	res, err := c.executeRequest(context.Background(), req)
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
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

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
