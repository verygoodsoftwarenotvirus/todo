package http

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	userAgentHeader = "User-Agent"
	userAgent       = "TODO Service Client"

	maxRetryCount = 10
	minRetryWait  = 500 * time.Millisecond
	maxRetryWait  = 5 * time.Second

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

func buildRetryingClient(client *http.Client, logger logging.Logger) *http.Client {
	rc := &retryablehttp.Client{
		HTTPClient:   client,
		Logger:       logging.EnsureLogger(logger),
		RetryWaitMin: minRetryWait,
		RetryWaitMax: maxRetryWait,
		RetryMax:     maxRetryCount,
		RequestLogHook: func(_ retryablehttp.Logger, req *http.Request, numTries int) {
			if req != nil {
				logger.WithRequest(req).WithValue("attempt_number", numTries).Debug("making request")
			}
		},
		ResponseLogHook: func(_ retryablehttp.Logger, res *http.Response) {
			if res != nil {
				logger.WithResponse(res).Debug("received response")
			}
		},
		CheckRetry: func(ctx context.Context, res *http.Response, err error) (bool, error) {
			ctx, span := tracing.StartCustomSpan(ctx, "CheckRetry")
			defer span.End()

			tracing.AttachResponseToSpan(span, res)

			return retryablehttp.DefaultRetryPolicy(ctx, res, err)
		},
		Backoff: retryablehttp.DefaultBackoff,
		ErrorHandler: func(res *http.Response, err error, numTries int) (*http.Response, error) {
			logger.WithValue("try_number", numTries).WithResponse(res).Error(err, "executing request")
			return res, err
		},
	}

	c := rc.StandardClient()
	c.Timeout = defaultTimeout

	return c
}
