package httpclient

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type option func(*V1Client)

// SetOption sets a new option on the client.
func (c *V1Client) SetOption(opt option) {
	opt(c)
}

// WithURL sets the URL on the client.
func WithURL(u *url.URL) func(*V1Client) {
	return func(c *V1Client) {
		c.URL = u
	}
}

// WithLogger sets the logger on the client.
func WithLogger(logger logging.Logger) func(*V1Client) {
	return func(c *V1Client) {
		if logger == nil {
			return
		}

		c.logger = logger
	}
}

// WithPlainClient sets the logger on the client.
func WithPlainClient(client *http.Client) func(*V1Client) {
	return func(c *V1Client) {
		if client == nil {
			return
		}

		if client.Timeout == 0 {
			client.Timeout = defaultTimeout
		}

		client.Transport = otelhttp.NewTransport(buildDefaultTransport())
		c.plainClient = client
	}
}

// WithAuthenticatedClient sets the logger on the client.
func WithAuthenticatedClient(client *http.Client) func(*V1Client) {
	return func(c *V1Client) {
		if client == nil {
			return
		}

		if client.Timeout == 0 {
			client.Timeout = defaultTimeout
		}

		client.Transport = otelhttp.NewTransport(buildDefaultTransport())
		c.authedClient = client
	}
}

// WithDebugEnabled sets the debug value on the client.
func WithDebugEnabled() func(*V1Client) {
	return func(c *V1Client) {
		c.Debug = true
	}
}

// WithOAuth2ClientCredentials sets the debug value on the client.
func WithOAuth2ClientCredentials(
	conf *clientcredentials.Config,
	timeout time.Duration,
) func(*V1Client) {
	return func(c *V1Client) {
		if timeout == 0 {
			timeout = defaultTimeout
		}

		c.tokenSource = oauth2.ReuseTokenSource(nil, conf.TokenSource(context.Background()))
		c.authedClient = &http.Client{
			Transport: &oauth2.Transport{
				Base:   otelhttp.NewTransport(newDefaultRoundTripper()),
				Source: c.tokenSource,
			},
			Timeout: timeout,
		}
	}
}
