package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
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
)

const (
	defaultTimeout = 5 * time.Second
)

type Config struct {
	Client  *http.Client
	Debug   bool
	Logger  *logrus.Logger
	Address string

	UserCredentials *url.Userinfo
	AuthToken       *string
}

type V1Client struct {
	Client *http.Client
	logger *logrus.Logger
	Debug  bool
	URL    *url.URL

	// old and busted
	authCookie *http.Cookie
	// new hotness
	authToken string

	Items     <-chan *models.Item
	itemsChan chan *models.Item
}

func buildOauth2Client(info *models.Oauth2Client, baseURL *url.URL) (*http.Client, error) {
	if strings.ToLower(baseURL.Scheme) != "https" {
		return nil, errors.New("https is required")
	}

	conf := &oauth2.Config{
		ClientID:     info.ClientID,
		ClientSecret: info.ClientSecret,
		Scopes:       []string{"*"},
		Endpoint: oauth2.Endpoint{
			TokenURL: fmt.Sprintf("https://%s/oauth2/token", baseURL.Hostname()),
			AuthURL:  fmt.Sprintf("https://%s/oauth2/authorize", baseURL.Hostname()),
		},
	}

	// Use the custom HTTP client when requesting a token.
	httpClient := &http.Client{
		Timeout:   2 * time.Second,
		Transport: http.DefaultTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	httpClient.Transport.(*http.Transport).TLSClientConfig = &tls.Config{
		// WARNING: Never do this ordinarily, this is an application which will only ever run in a local context
		InsecureSkipVerify: true,
	}

	aurl := conf.AuthCodeURL(
		"",
		oauth2.SetAuthURLParam("client_id", conf.ClientID),
		oauth2.SetAuthURLParam("client_secret", conf.ClientSecret),
		oauth2.SetAuthURLParam("redirect_uri", "https://yourredirecturl.com"),
	)
	req, err := http.NewRequest(http.MethodPost, aurl, nil)
	if err != nil || req == nil {
		panic(err)
	}
	//req.AddCookie(cookie)

	res, err := httpClient.Do(req)
	if err != nil {
		log.Fatal("error trying to get authorized", err)
	}
	u, _ := url.Parse(res.Header.Get("Location"))
	tok := u.Query().Get("code")

	client := conf.Client(context.Background(), &oauth2.Token{
		AccessToken: tok,
		TokenType:   "Bearer",
		//RefreshToken
		Expiry: time.Now().Add(10 * time.Minute),
	})
	return client, nil
}

func NewClient(cfg *Config) (c *V1Client, err error) {
	c = &V1Client{
		Debug: cfg.Debug,
	}

	if c.URL, err = url.Parse(cfg.Address); err != nil {
		return nil, errors.Wrap(err, "parsing URL")
	}

	if cfg.AuthToken != nil {
		c.authToken = *cfg.AuthToken
	}

	if cfg.Client != nil {
		c.Client = cfg.Client
	} else {
		c.Client = &http.Client{Timeout: 5 * time.Second}
	}

	if cfg.Logger != nil {
		c.logger = cfg.Logger
	} else {
		c.logger = logrus.New()
		if cfg.Debug {
			c.logger.SetLevel(logrus.DebugLevel)
		}
	}

	return
}

func (c *V1Client) executeRequest(req *http.Request) (*http.Response, error) {
	command, err := http2curl.GetCurlCommand(req)
	if err == nil {
		c.logger.Debugln(command)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authToken))
	if c.authCookie != nil {
		req.AddCookie(c.authCookie)
	}

	res, err := c.Client.Do(req)
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

func (c *V1Client) Do(req *http.Request) (*http.Response, error) {
	if c.URL.Hostname() != req.URL.Hostname() {
		return nil, errors.New("request is destined for unknown server")
	}
	return c.executeRequest(req)
}

type Valuer interface {
	ToValues() url.Values
}

func (c *V1Client) BuildURL(qp Valuer, parts ...string) string {
	if qp != nil {
		return c.buildURL(qp.ToValues(), parts...).String()
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

func (c *V1Client) BuildWebsocketURL(parts ...string) string {
	u := c.buildURL(nil, parts...)
	u.Scheme = "ws"

	return u.String()
}

func (c *V1Client) IsUp() bool {
	u := *c.URL

	uri := fmt.Sprintf("%s://%s:%s/_meta_/health", u.Scheme, u.Hostname(), u.Port())
	req, _ := http.NewRequest(http.MethodGet, uri, nil)
	res, err := c.executeRequest(req)

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
	ce := &ClientError{}

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
		return &ClientError{Err: err}
	} else if res.StatusCode != http.StatusOK {
		return &ClientError{Err: errors.New(fmt.Sprintf("status returned: %d", res.StatusCode))}
	}

	return nil
}

func (c *V1Client) post(uri string, in interface{}, out interface{}) error {
	return c.makeDataRequest(http.MethodPost, uri, in, out)
}

func (c *V1Client) put(uri string, in interface{}, out interface{}) error {
	return c.makeDataRequest(http.MethodPut, uri, in, out)
}

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
