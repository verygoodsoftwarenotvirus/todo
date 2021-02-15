package httpclient

import (
	"net"
	"net/http"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
)

const (
	userAgentHeader = "User-Agent"
	userAgent       = "TODO Service Client"

	keepAlive             = 30 * time.Second
	tlsHandshakeTimeout   = 10 * time.Second
	expectContinueTimeout = 2 * defaultTimeout
	idleConnTimeout       = 3 * defaultTimeout
	maxIdleConns          = 100
)

type defaultRoundTripper struct {
	baseTransport *http.Transport
}

// newDefaultRoundTripper constructs a new http.RoundTripper.
func newDefaultRoundTripper(timeout time.Duration) *defaultRoundTripper {
	return &defaultRoundTripper{
		baseTransport: buildDefaultTransport(timeout),
	}
}

// RoundTrip implements the http.RoundTripper interface.
func (t *defaultRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set(userAgentHeader, userAgent)
	return t.baseTransport.RoundTrip(req)
}

// buildDefaultTransport constructs a new http.Transport.
func buildDefaultTransport(timeout time.Duration) *http.Transport {
	if timeout == 0 {
		timeout = defaultTimeout
	}

	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: keepAlive,
		}).DialContext,
		MaxIdleConns:          maxIdleConns,
		MaxIdleConnsPerHost:   maxIdleConns,
		TLSHandshakeTimeout:   tlsHandshakeTimeout,
		ExpectContinueTimeout: expectContinueTimeout,
		IdleConnTimeout:       idleConnTimeout,
	}
}

// cookieRoundtripper is a transport that uses a cookie.
type cookieRoundtripper struct {
	cookie *http.Cookie

	logger logging.Logger

	// base is the base RoundTripper used to make HTTP requests. If nil, http.DefaultTransport is used.
	base http.RoundTripper
}

// RoundTrip authorizes and authenticates the request with a cookie.
func (t *cookieRoundtripper) RoundTrip(req *http.Request) (*http.Response, error) {
	reqBodyClosed := false

	if req.Body != nil {
		defer func() {
			if !reqBodyClosed {
				if err := req.Body.Close(); err != nil {
					t.logger.Error(err, "closing response body")
				}
			}
		}()
	}

	if c, err := req.Cookie(t.cookie.Name); c == nil || err != nil {
		req.AddCookie(t.cookie)
	}

	// req.Body is assumed to be closed by the base RoundTripper.
	reqBodyClosed = true

	return t.base.RoundTrip(req)
}
