package http

import (
	"net"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
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
	baseRoundTripper http.RoundTripper
}

// newDefaultRoundTripper constructs a new http.RoundTripper.
func newDefaultRoundTripper(timeout time.Duration) http.RoundTripper {
	return &defaultRoundTripper{
		baseRoundTripper: buildWrappedTransport(timeout),
	}
}

// RoundTrip implements the http.RoundTripper interface.
func (t *defaultRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set(userAgentHeader, userAgent)

	return t.baseRoundTripper.RoundTrip(req)
}

// buildWrappedTransport constructs a new http.Transport.
func buildWrappedTransport(timeout time.Duration) http.RoundTripper {
	if timeout == 0 {
		timeout = defaultTimeout
	}

	t := &http.Transport{
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

	return otelhttp.NewTransport(t, otelhttp.WithSpanNameFormatter(tracing.FormatSpan))
}
