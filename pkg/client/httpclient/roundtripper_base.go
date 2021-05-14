package httpclient

import (
	"context"
	"net"
	"net/http"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"github.com/hashicorp/go-retryablehttp"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	userAgentHeader = "User-Agent"
	userAgent       = "TODO Service Client"

	maxRetryCount = 5
	minRetryWait  = 100 * time.Millisecond
	maxRetryWait  = time.Second

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

func buildRequestLogHook(logger logging.Logger) func(retryablehttp.Logger, *http.Request, int) {
	logger = logging.EnsureLogger(logger)

	return func(_ retryablehttp.Logger, req *http.Request, numRetries int) {
		if req != nil {
			logger.WithRequest(req).WithValue("retry_count", numRetries).Debug("making request")
		}
	}
}

func buildResponseLogHook(logger logging.Logger) func(retryablehttp.Logger, *http.Response) {
	logger = logging.EnsureLogger(logger)

	return func(_ retryablehttp.Logger, res *http.Response) {
		if res != nil {
			logger.WithResponse(res).Debug("received response")
		}
	}
}

func buildCheckRetryFunc(tracer tracing.Tracer) func(context.Context, *http.Response, error) (bool, error) {
	return func(ctx context.Context, res *http.Response, err error) (bool, error) {
		ctx, span := tracer.StartCustomSpan(ctx, "CheckRetry")
		defer span.End()

		if res != nil {
			tracing.AttachResponseToSpan(span, res)
		}

		return retryablehttp.DefaultRetryPolicy(ctx, res, err)
	}
}

func buildErrorHandler(logger logging.Logger) func(res *http.Response, err error, numTries int) (*http.Response, error) {
	logger = logging.EnsureLogger(logger)

	return func(res *http.Response, err error, numTries int) (*http.Response, error) {
		logger.WithValue("try_number", numTries).WithResponse(res).Error(err, "executing request")

		return res, err
	}
}

func buildRetryingClient(client *http.Client, logger logging.Logger, tracer tracing.Tracer) *http.Client {
	rc := &retryablehttp.Client{
		HTTPClient:      client,
		Logger:          logging.EnsureLogger(logger),
		RetryWaitMin:    minRetryWait,
		RetryWaitMax:    maxRetryWait,
		RetryMax:        maxRetryCount,
		RequestLogHook:  buildRequestLogHook(logger),
		ResponseLogHook: buildResponseLogHook(logger),
		CheckRetry:      buildCheckRetryFunc(tracer),
		Backoff:         retryablehttp.DefaultBackoff,
		ErrorHandler:    buildErrorHandler(logger),
	}

	c := rc.StandardClient()
	c.Timeout = defaultTimeout

	return c
}
