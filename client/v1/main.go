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

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/gorilla/websocket"
	"github.com/moul/http2curl"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	defaultTimeout = 5 * time.Second
)

var (
	// ErrNotFound is a handy error to return when we receive a 404 response
	ErrNotFound = errors.New("404: not found")
)

// V1Client is a client for interacting with v1 of our API
type V1Client struct {
	plainClient  *http.Client
	authedClient *http.Client
	logger       *logrus.Logger
	Debug        bool
	URL          *url.URL

	// old and busted
	authCookie *http.Cookie
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

func endpoint(baseURL *url.URL) oauth2.Endpoint {
	tu, au := *baseURL, *baseURL
	tu.Path, au.Path = "oauth2/token", "oauth2/authorize"

	return oauth2.Endpoint{
		TokenURL: tu.String(),
		AuthURL:  au.String(),
	}
}

func (c *V1Client) enableOauth(clientID, clientSecret string) error {
	c.logger.Debugln("enabling OAuth")

	conf := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"*"},
		EndpointParams: url.Values{
			"client_id":     []string{clientID},
			"client_secret": []string{clientSecret},
		},
		TokenURL: endpoint(c.URL).TokenURL,
	}
	c.authedClient = conf.Client(context.TODO())

	return nil
}

// NewClient builds a new API client for us
func NewClient(address, clientID, clientSecret string, logger *logrus.Logger, client *http.Client, debug bool) (*V1Client, error) {
	c := &V1Client{Debug: debug}

	if clientID == "" || clientSecret == "" {
		return nil, errors.New("Client ID and Client Secret required")
	}

	u, err := url.Parse(address)
	if err != nil {
		return nil, errors.Wrap(err, "parsing URL")
	}
	c.URL = u

	if client != nil {
		c.plainClient = client
	} else {
		c.plainClient = &http.Client{Timeout: 5 * time.Second}
	}

	if logger != nil {
		c.logger = logger
	} else {
		c.logger = logrus.New()
		if debug {
			c.logger.SetLevel(logrus.DebugLevel)
		}
	}
	if err := c.enableOauth(clientID, clientSecret); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *V1Client) executeRequest(req *http.Request) (*http.Response, error) {
	command, err := http2curl.GetCurlCommand(req)
	if err == nil {
		c.logger.Debugln(command)
	}

	res, err := c.authedClient.Do(req)
	if err != nil {
		return nil, err
	}

	dump, err := httputil.DumpResponse(res, true)
	if err == nil {
		if req.Method != http.MethodGet {
			d := string(dump)
			c.logger.Debugln(d)
		}
	}

	return res, nil
}

// Do executes a raw request object
// TODO: find out why this was implemented
func (c *V1Client) Do(req *http.Request) (*http.Response, error) {
	if c.URL.Hostname() != req.URL.Hostname() {
		return nil, errors.New("request is destined for unknown server")
	}
	return c.executeRequest(req)
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

	uri := fmt.Sprintf("%s://%s:%s/_meta_/health", u.Scheme, u.Hostname(), u.Port())
	req, _ := http.NewRequest(http.MethodGet, uri, nil)
	res, err := c.plainClient.Do(req)

	if err != nil {
		return false
	}

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

	var (
		req *http.Request
		res *http.Response
		err error
	)

	if req, err = c.buildDataRequest(method, uri, in); err != nil {
		return errors.Wrap(err, "encountered error building request")
	}

	if res, err = c.executeRequest(req); err != nil {
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
		c.logger.Debugf("data request returned: %+v", out)
	}

	return nil
}

func (c *V1Client) exists(uri string) (bool, error) {
	req, _ := http.NewRequest(http.MethodHead, uri, nil)
	res, err := c.executeRequest(req)
	if err != nil {
		return false, errors.Wrap(err, "encountered error executing request")
	}

	return res.StatusCode == http.StatusOK, nil
}

func (c *V1Client) get(uri string, obj interface{}) error {
	ce := &Error{}

	if err := argIsNotPointerOrNil(obj); err != nil {
		ce.Err = errors.Wrap(err, "struct to load must be a pointer")
		return ce
	}

	req, _ := http.NewRequest(http.MethodGet, uri, nil)
	res, err := c.executeRequest(req)
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

func (c *V1Client) delete(uri string) error {
	req, _ := http.NewRequest(http.MethodDelete, uri, nil)
	res, err := c.executeRequest(req)
	if err != nil {
		return &Error{Err: err}
	} else if res.StatusCode != http.StatusOK {
		return &Error{Err: errors.New(fmt.Sprintf("status returned: %d", res.StatusCode))}
	}

	return nil
}

func (c *V1Client) post(uri string, in interface{}, out interface{}) error {
	return c.makeDataRequest(http.MethodPost, uri, in, out)
}

func (c *V1Client) put(uri string, in interface{}, out interface{}) error {
	return c.makeDataRequest(http.MethodPut, uri, in, out)
}

// DialWebsocket dials a websocket
func (c *V1Client) DialWebsocket(fq *FeedQuery) (*websocket.Conn, error) {
	if fq == nil {
		return nil, errors.New("Valid feed query required")
	}

	if !c.IsUp() {
		c.logger.Debugln("returning early from ItemsFeed because the service is down")
		return nil, errors.New("service is down")
	}

	u := c.buildURL(fq.Values(), "event_feed")
	u.Scheme = "wss"
	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	c.logger.Debugf("connecting to websocket at %q", u.String())
	conn, res, err := dialer.Dial(u.String(), nil)
	if err != nil {
		c.logger.Debugf("encountered error dialing %q: %v", u.String(), err)
		return nil, err
	}

	if res.StatusCode < http.StatusBadRequest {
		return conn, nil
	}
	return nil, fmt.Errorf("encountered status code: %d when trying to reach websocket", res.StatusCode)
}
