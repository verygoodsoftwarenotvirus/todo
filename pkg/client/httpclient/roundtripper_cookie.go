package httpclient

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
)

// cookieRoundtripper is a transport that uses a cookie.
type cookieRoundtripper struct {
	cookie *http.Cookie

	logger logging.Logger
	tracer tracing.Tracer

	// base is the base RoundTripper used to make HTTP requests. If nil, http.DefaultTransport is used.
	base http.RoundTripper
}

func newCookieRoundTripper(client *Client, cookie *http.Cookie) *cookieRoundtripper {
	return &cookieRoundtripper{
		cookie: cookie,
		logger: client.logger,
		tracer: client.tracer,
		base:   newDefaultRoundTripper(client.unauthenticatedClient.Timeout),
	}
}

// RoundTrip authorizes and authenticates the request with a cookie.
func (t *cookieRoundtripper) RoundTrip(req *http.Request) (*http.Response, error) {
	_, span := t.tracer.StartSpan(req.Context())
	defer span.End()

	reqBodyClosed := false
	logger := t.logger.WithRequest(req)

	if req.Body != nil {
		defer func() {
			if !reqBodyClosed {
				if err := req.Body.Close(); err != nil {
					observability.AcknowledgeError(err, t.logger, span, "closing response body")
				}
			}
		}()
	}

	if c, err := req.Cookie(t.cookie.Name); c == nil || err != nil {
		req.AddCookie(t.cookie)
	}

	// req.Body is assumed to be closed by the base RoundTripper.
	reqBodyClosed = true

	res, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "executing RoundTrip")
	}

	if responseCookies := res.Cookies(); len(responseCookies) >= 1 {
		t.cookie = responseCookies[0]
	}

	return res, nil
}
