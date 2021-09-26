package httpclient

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/httpclient/requests"
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
		requestBuilder, err := requests.NewBuilder(c.url, c.logger, encoding.ProvideClientEncoder(c.logger, encoding.ContentTypeJSON))
		if err != nil {
			return err
		}

		c.requestBuilder = requestBuilder
		c.encoder = encoding.ProvideClientEncoder(c.logger, encoding.ContentTypeJSON)

		return nil
	}
}

// UsingXML sets the url on the client.
func UsingXML() func(*Client) error {
	return func(c *Client) error {
		requestBuilder, err := requests.NewBuilder(c.url, c.logger, encoding.ProvideClientEncoder(c.logger, encoding.ContentTypeXML))
		if err != nil {
			return err
		}

		c.requestBuilder = requestBuilder
		c.encoder = encoding.ProvideClientEncoder(c.logger, encoding.ContentTypeXML)

		return nil
	}
}

// UsingLogger sets the logger on the client.
func UsingLogger(logger logging.Logger) func(*Client) error {
	return func(c *Client) error {
		c.logger = logging.EnsureLogger(logger)

		return nil
	}
}

// UsingDebug sets the debug value on the client.
func UsingDebug(debug bool) func(*Client) error {
	return func(c *Client) error {
		c.debug = debug

		if debug {
			c.logger.SetLevel(logging.DebugLevel)
		}

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
			return ErrCookieRequired
		}

		crt := newCookieRoundTripper(c, cookie)
		c.authMethod = cookieAuthMethod
		c.authedClient.Transport = crt
		c.authHeaderBuilder = crt
		c.websocketDialer = websocket.DefaultDialer
		c.authedClient = buildRetryingClient(c.authedClient, c.logger, c.tracer)

		c.logger.Debug("set client auth cookie")

		return nil
	}
}

// UsingPASETO sets the authCookie value on the client.
func UsingPASETO(clientID string, secretKey []byte) func(*Client) error {
	return func(c *Client) error {
		prt := newPASETORoundTripper(c, clientID, secretKey)

		c.authMethod = pasetoAuthMethod
		c.authedClient.Transport = prt
		c.authHeaderBuilder = prt
		c.websocketDialer = websocket.DefaultDialer
		c.authedClient = buildRetryingClient(c.authedClient, c.logger, c.tracer)

		return nil
	}
}
