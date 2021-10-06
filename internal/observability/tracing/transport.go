package tracing

import (
	"net"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	defaultTimeout = 30 * time.Second
)

// BuildTracedHTTPTransport constructs a new http.Transport.
func BuildTracedHTTPTransport(timeout time.Duration) http.RoundTripper {
	if timeout == 0 {
		timeout = defaultTimeout
	}

	keepAlive := 30 * time.Second
	tlsHandshakeTimeout := 10 * time.Second
	expectContinueTimeout := 2 * timeout
	idleConnTimeout := 3 * timeout
	maxIdleConns := 100

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

	return otelhttp.NewTransport(t, otelhttp.WithSpanNameFormatter(FormatSpan))
}
