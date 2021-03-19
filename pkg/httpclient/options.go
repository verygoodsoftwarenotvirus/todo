package httpclient

import (
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
)

type option func(*Client) error

// SetOptions sets a new option on the client.
func (c *Client) SetOptions(opts ...option) error {
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return err
		}
	}

	return nil
}

// UsingJSON sets the url on the client.
func UsingJSON() func(*Client) error {
	return func(c *Client) error {
		c.contentType = "application/json"

		return nil
	}
}

// UsingXML sets the url on the client.
func UsingXML() func(*Client) error {
	return func(c *Client) error {
		c.contentType = "application/xml"

		return nil
	}
}

// UsingLogger sets the logger on the client.
func UsingLogger(logger logging.Logger) func(*Client) error {
	return func(c *Client) error {
		c.logger = logging.EnsureLogger(logger)
		c.encoderDecoder = encoding.ProvideHTTPResponseEncoder(c.logger)

		return nil
	}
}

// UsingHTTPClient sets the unauthenticatedClient value on the client.
func UsingHTTPClient(client *http.Client) func(*Client) error {
	return func(c *Client) error {
		if client == nil {
			return nil
		}

		if client.Timeout == 0 {
			client.Timeout = defaultTimeout
		}

		client.Transport = otelhttp.NewTransport(
			buildDefaultTransport(client.Timeout),
			otelhttp.WithSpanNameFormatter(tracing.FormatSpan),
		)

		c.unauthenticatedClient = client
		c.authedClient = client

		return nil
	}
}

// WithDebug sets the debug value on the client.
func WithDebug() func(*Client) error {
	return func(c *Client) error {
		c.debug = true
		c.logger.SetLevel(logging.DebugLevel)

		return nil
	}
}

// UsingTimeout sets the debug value on the client.
func UsingTimeout(timeout time.Duration) func(*Client) error {
	return func(c *Client) error {
		if timeout == 0 {
			timeout = defaultTimeout
		}

		c.authedClient.Timeout = timeout
		c.unauthenticatedClient.Timeout = timeout

		return nil
	}
}

// UsingCookie sets the authCookie value on the client.
func UsingCookie(cookie *http.Cookie) func(*Client) error {
	return func(c *Client) error {
		if cookie == nil {
			return ErrNilInputProvided
		}

		c.authMethod = cookieAuthMethod
		c.authedClient.Transport = newCookieRoundTripper(c, cookie)

		return nil
	}
}

// UsingPASETO sets the authCookie value on the client.
func UsingPASETO(clientID string, secretKey []byte) func(*Client) error {
	return func(c *Client) error {
		c.authMethod = pasetoAuthMethod
		c.authedClient.Transport = newPASETORoundTripper(c, clientID, secretKey)

		return nil
	}
}
