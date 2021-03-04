package httpclient

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type option func(*Client) error

// SetOption sets a new option on the client.
func (c *Client) SetOption(opt option) error {
	if err := opt(c); err != nil {
		return err
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

// UsingURI sets the url on the client.
func UsingURI(raw string) func(*Client) error {
	return func(c *Client) error {
		c.url = mustParseURL(raw)

		return nil
	}
}

// UsingURL sets the url on the client.
func UsingURL(u *url.URL) func(*Client) error {
	return func(c *Client) error {
		c.url = u

		return nil
	}
}

// UsingAccount sets the url on the client.
func UsingAccount(accountID uint64) func(*Client) error {
	return func(c *Client) error {
		c.accountID = accountID

		if c.authMethod == cookieAuthMethod {
			cookie, err := c.SwitchActiveAccount(context.Background(), &types.ChangeActiveAccountInput{AccountID: accountID})
			if err != nil {
				c.logger.Error(err, "fetching cookie for new account")
				return err
			}
			return c.SetOption(UsingCookie(cookie))
		}

		return nil
	}
}

// UsingLogger sets the logger on the client.
func UsingLogger(logger logging.Logger) func(*Client) error {
	return func(c *Client) error {
		if logger == nil {
			return nil
		}

		c.logger = logger
		c.encoderDecoder = encoding.ProvideHTTPResponseEncoder(logger)

		return nil
	}
}

// UsingHTTPClient sets the plainClient value on the client.
func UsingHTTPClient(client *http.Client) func(*Client) error {
	return func(c *Client) error {
		if client == nil {
			return nil
		}

		if client.Timeout == 0 {
			client.Timeout = defaultTimeout
		}

		client.Transport = otelhttp.NewTransport(buildDefaultTransport(client.Timeout))

		c.plainClient = client
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
		c.plainClient.Timeout = timeout

		return nil
	}
}

// UsingCookie sets the authCookie value on the client.
func UsingCookie(cookie *http.Cookie) func(*Client) error {
	return func(c *Client) error {
		if cookie == nil {
			return nil
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
