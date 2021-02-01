package httpclient

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type option func(*Client)

// SetOption sets a new option on the client.
func (c *Client) SetOption(opt option) {
	opt(c)
}

// WithRawURL sets the url on the client.
func WithRawURL(raw string) func(*Client) {
	return func(c *Client) {
		c.url = MustParseURL(raw)
	}
}

// WithURL sets the url on the client.
func WithURL(u *url.URL) func(*Client) {
	return func(c *Client) {
		c.url = u
	}
}

// WithLogger sets the logger on the client.
func WithLogger(logger logging.Logger) func(*Client) {
	return func(c *Client) {
		if logger == nil {
			return
		}

		c.logger = logger
	}
}

// WithHTTPClient sets the plainClient value on the client.
func WithHTTPClient(client *http.Client) func(*Client) {
	return func(c *Client) {
		if client == nil {
			return
		}

		if client.Timeout == 0 {
			client.Timeout = defaultTimeout
		}

		client.Transport = otelhttp.NewTransport(buildDefaultTransport())
		c.plainClient = client
		c.authedClient = client
	}
}

// WithDebugEnabled sets the debug value on the client.
func WithDebugEnabled() func(*Client) {
	return func(c *Client) {
		c.debug = true
		c.logger.SetLevel(logging.DebugLevel)
	}
}

// WithTimeout sets the debug value on the client.
func WithTimeout(timeout time.Duration) func(*Client) {
	return func(c *Client) {
		if timeout == 0 {
			timeout = defaultTimeout
		}

		c.authedClient.Timeout = timeout
		c.plainClient.Timeout = timeout
	}
}

// WithCookieCredentials sets the authCookie value on the client.
func WithCookieCredentials(cookie *http.Cookie) func(*Client) {
	return func(c *Client) {
		if cookie == nil {
			return
		}

		c.authMode = cookieAuthMode
		c.authCookie = cookie
	}
}

// WithOAuth2ClientCredentials sets the oauth2 credentials for the client.
func WithOAuth2ClientCredentials(conf *clientcredentials.Config) func(*Client) {
	return func(c *Client) {
		c.tokenSource = oauth2.ReuseTokenSource(nil, conf.TokenSource(context.Background()))
		c.authedClient = &http.Client{
			Transport: &oauth2.Transport{
				Base:   otelhttp.NewTransport(newDefaultRoundTripper()),
				Source: c.tokenSource,
			},
			Timeout: defaultTimeout,
		}

		c.authMode = oauth2AuthMode
	}
}
