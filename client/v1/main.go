package client

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	// "gitlab.com/verygoodsoftwarenotvirus/todo/models"

	"github.com/moul/http2curl"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const defaultTimeout = 5 * time.Second

type Config struct {
	Client    *http.Client
	Debug     bool
	Logger    *logrus.Logger
	BaseURL   string
	AuthToken string
}

type V1Client struct {
	*http.Client
	Debug bool
	URL   *url.URL
	Token string
}

func NewClient(cfg Config) (*V1Client, error) {
	var c *V1Client
	if cfg.Client != nil {
		c = &V1Client{Client: cfg.Client}
	}

	u, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, errors.Wrap(err, "Store URL is not valid")
	}
	c.URL = u

	p := fmt.Sprintf("%s://%s/login", u.Scheme, u.Host)
	body := strings.NewReader(fmt.Sprintf(`
		{
			"username": "%s",
			"password": "%s"
		}
	`, username, password))
	req, _ := http.NewRequest(http.MethodPost, p, body)
	res, err := dc.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Error encountered logging into store")
	}
	cookies := res.Cookies()
	if len(cookies) == 0 {
		return nil, errors.New("No cookies returned with login response")
	}

	dc.Client.Timeout = defaultTimeout

	return dc, nil
}

func NewV1ClientFromCookie(apiURL string, cookie *http.Cookie, client *http.Client) (*V1Client, error) {
	var dc *V1Client
	if client != nil {
		dc = &V1Client{Client: client}
	}

	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, errors.Wrap(err, "API URL is not valid")
	}
	dc.URL = u

	dc.Client.Timeout = defaultTimeout

	return dc, nil
}

func (dc *V1Client) executeRequest(req *http.Request) (*http.Response, error) {
	if dc.Debug {
		command, err := http2curl.GetCurlCommand(req)
		if err == nil {
			// FIXME: add a real logger
			fmt.Println(command)
		}
	}

	res, err := dc.Do(req)
	if err != nil {
		return nil, err
	}

	if dc.Debug {
		dump, err := httputil.DumpResponse(res, true)
		if err != nil {
			return res, err
		}
		// FIXME: add a real logger
		fmt.Printf("%s", dump)
	}

	return res, nil
}

func (dc *V1Client) Raw(req *http.Request) (*http.Response, error) {
	return dc.executeRequest(req)
}

func (dc *V1Client) buildURL(queryParams map[string]string, parts ...string) string {
	parts = append([]string{"v1"}, parts...)
	u, _ := url.Parse(strings.Join(parts, "/"))
	queryString := mapToQueryValues(queryParams)
	u.RawQuery = queryString.Encode()
	return dc.URL.ResolveReference(u).String()
}

// BuildURL is the same as the unexported build URL, except I trust myself to never call the
// unexported function with variables that could lead to an error being returned. This function
// returns the error in the event a user needs to build an API url, but tries to do so with an
// invalid value.
func (dc *V1Client) BuildURL(queryParams map[string]string, parts ...string) (string, error) {
	parts = append([]string{"v1"}, parts...)

	u, err := url.Parse(strings.Join(parts, "/"))
	if err != nil {
		return "", err
	}

	queryString := mapToQueryValues(queryParams)
	u.RawQuery = queryString.Encode()
	return dc.URL.ResolveReference(u).String(), nil
}

func (dc *V1Client) exists(uri string) (bool, error) {
	req, _ := http.NewRequest(http.MethodHead, uri, nil)
	res, err := dc.executeRequest(req)
	if err != nil {
		return false, errors.Wrap(err, "encountered error executing request")
	}

	return res.StatusCode == http.StatusOK, nil
}

func (dc *V1Client) get(uri string, obj interface{}) error {
	ce := &ClientError{}

	if err := argIsNotPointerOrNil(obj); err != nil {
		ce.Err = errors.Wrap(err, "struct to load must be a pointer")
		return ce
	}

	req, _ := http.NewRequest(http.MethodGet, uri, nil)
	res, err := dc.executeRequest(req)
	if err != nil {
		ce.Err = errors.Wrap(err, "encountered error executing request")
		return ce
	}

	return unmarshalBody(res, &obj)
}

func (dc *V1Client) delete(uri string, out interface{}) error {
	ce := &ClientError{}
	if err := argIsNotPointerOrNil(out); err != nil {
		ce.Err = errors.Wrap(err, "struct to load must be a pointer")
		return ce
	}

	req, _ := http.NewRequest(http.MethodDelete, uri, nil)
	res, err := dc.executeRequest(req)
	if err != nil {
		return &ClientError{Err: err}
	}

	return unmarshalBody(res, nil)
}

func (dc *V1Client) makeDataRequest(method string, uri string, in interface{}, out interface{}) error {
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
	res, err := dc.executeRequest(req)
	if err != nil {
		ce.Err = errors.Wrap(err, "encountered error executing request")
		return ce
	}

	resErr := unmarshalBody(res, &out)
	if resErr != nil {
		ce.Err = errors.Wrap(err, "encountered error loading response from server")
		return ce
	}

	return nil
}

func (dc *V1Client) post(uri string, in interface{}, out interface{}) error {
	return dc.makeDataRequest(http.MethodPost, uri, in, out)
}

func (dc *V1Client) patch(uri string, in interface{}, out interface{}) error {
	return dc.makeDataRequest(http.MethodPatch, uri, in, out)
}
