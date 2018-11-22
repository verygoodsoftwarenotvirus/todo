package client

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	"github.com/moul/http2curl"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	defaultTimeout = 5 * time.Second
)

type Config struct {
	Client    *http.Client
	Debug     bool
	Logger    *logrus.Logger
	Address   string
	AuthToken string
}

type V1Client struct {
	Client *http.Client
	logger *logrus.Logger
	Debug  bool
	URL    *url.URL
	Token  string

	Items     <-chan *models.Item
	itemsChan chan *models.Item
}

func NewClient(cfg *Config) (c *V1Client, err error) {
	c = &V1Client{
		Debug: cfg.Debug,
		Token: cfg.AuthToken,
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

	if c.URL, err = url.Parse(cfg.Address); err != nil {
		return nil, errors.Wrap(err, "Invalid URL is invalid")
	}

	return
}

func (c *V1Client) executeRequest(req *http.Request) (*http.Response, error) {
	command, err := http2curl.GetCurlCommand(req)
	if err == nil {
		c.logger.Debugln(command)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

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

func (c *V1Client) BuildURL(queryParams map[string]string, parts ...string) string {
	return c.buildURL(queryParams, parts...).String()
}

func (c *V1Client) buildURL(queryParams map[string]string, parts ...string) *url.URL {
	tu := *c.URL

	parts = append([]string{"api", "v1"}, parts...)
	u, _ := url.Parse(strings.Join(parts, "/"))

	if queryParams != nil {
		query := url.Values{}
		for k, v := range queryParams {
			query.Set(k, v)
		}
		u.RawQuery = query.Encode()
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

	uri := fmt.Sprintf("%s://%s:%s/_debug_/health", u.Scheme, u.Hostname(), u.Port())
	req, _ := http.NewRequest(http.MethodGet, uri, nil)
	res, err := c.executeRequest(req)

	if err != nil {
		return false
	}

	return res.StatusCode == http.StatusOK
}

func (c *V1Client) makeDataRequest(method string, uri string, in interface{}, out interface{}) error {
	ce := &ClientError{}

	if err := argIsNotPointerOrNil(out); err != nil {
		ce.Err = errors.Wrap(err, "struct to load must be a pointer")
		return ce
	}

	body, err := createBodyFromStruct(in)
	if err != nil {
		ce.Err = errors.Wrap(err, "encountered error marshaling data to JSON")
		return ce
	}

	req, _ := http.NewRequest(method, uri, body)
	res, err := c.executeRequest(req)
	if err != nil {
		ce.Err = errors.Wrap(err, "encountered error executing request")
		return ce
	}

	resErr := unmarshalBody(res, &out)
	if resErr != nil {
		ce.Err = errors.Wrap(err, "encountered error loading response from server")
		return ce
	}

	c.logger.Debugf("data request returned: %+v", out)

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
