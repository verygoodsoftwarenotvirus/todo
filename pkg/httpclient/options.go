package httpclient

import (
	"net/http"
	"net/url"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type option func(*Client)

// SetOption sets a new option on the client.
func (c *Client) SetOption(opt option) {
	opt(c)
}

// UsingJSON sets the url on the client.
func UsingJSON() func(*Client) {
	return func(c *Client) {
		c.contentType = "application/json"
	}
}

// UsingXML sets the url on the client.
func UsingXML() func(*Client) {
	return func(c *Client) {
		c.contentType = "application/xml"
	}
}

// UsingURI sets the url on the client.
func UsingURI(raw string) func(*Client) {
	return func(c *Client) {
		c.url = mustParseURL(raw)
	}
}

// UsingURL sets the url on the client.
func UsingURL(u *url.URL) func(*Client) {
	return func(c *Client) {
		c.url = u
	}
}

// UsingLogger sets the logger on the client.
func UsingLogger(logger logging.Logger) func(*Client) {
	return func(c *Client) {
		if logger == nil {
			return
		}

		c.logger = logger
		c.encoderDecoder = encoding.ProvideHTTPResponseEncoder(logger)
	}
}

// UsingHTTPClient sets the plainClient value on the client.
func UsingHTTPClient(client *http.Client) func(*Client) {
	return func(c *Client) {
		if client == nil {
			return
		}

		if client.Timeout == 0 {
			client.Timeout = defaultTimeout
		}

		client.Transport = otelhttp.NewTransport(buildDefaultTransport(client.Timeout))

		c.plainClient = client
		c.authedClient = client
	}
}

// WithDebug sets the debug value on the client.
func WithDebug() func(*Client) {
	return func(c *Client) {
		c.debug = true
		c.logger.SetLevel(logging.DebugLevel)
	}
}

// UsingTimeout sets the debug value on the client.
func UsingTimeout(timeout time.Duration) func(*Client) {
	return func(c *Client) {
		if timeout == 0 {
			timeout = defaultTimeout
		}

		c.authedClient.Timeout = timeout
		c.plainClient.Timeout = timeout
	}
}

// UsingCookie sets the authCookie value on the client.
func UsingCookie(cookie *http.Cookie) func(*Client) {
	return func(c *Client) {
		if cookie == nil {
			return
		}

		c.authMethod = cookieAuthMethod
		c.authedClient.Transport = newCookieRoundTripper(c, cookie)
	}
}

// UsingPASETO sets the authCookie value on the client.
func UsingPASETO(clientID string, secretKey []byte) func(*Client) {
	return func(c *Client) {
		c.authMethod = pasetoAuthMethod
		c.authedClient.Transport = newPASETORoundTripper(c, clientID, secretKey)
	}
}
