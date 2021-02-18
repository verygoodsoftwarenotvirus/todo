package httpclient

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
)

// pasetoRoundTripper is a transport that uses a cookie.
type pasetoRoundTripper struct {
	clientID  string
	secretKey []byte

	logger logging.Logger
	tracer tracing.Tracer

	// base is the base RoundTripper used to make HTTP requests. If nil, http.DefaultTransport is used.
	base   http.RoundTripper
	client *Client
}

func newPASETORoundTripper(client *Client, clientID string, secretKey []byte) *pasetoRoundTripper {
	return &pasetoRoundTripper{
		clientID:  clientID,
		secretKey: secretKey,
		logger:    client.logger,
		tracer:    client.tracer,
		base:      otelhttp.NewTransport(newDefaultRoundTripper(client.plainClient.Timeout)),
		client:    client,
	}
}

// RoundTrip authorizes and authenticates the request with a cookie.
func (t *pasetoRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, span := t.tracer.StartSpan(req.Context())
	defer span.End()

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

	token, tokenRetrievalErr := t.client.fetchDelegatedClientAuthToken(ctx, t.clientID, t.secretKey)
	if tokenRetrievalErr != nil {
		return nil, fmt.Errorf("error fetching prerequisite PASETO: %w", tokenRetrievalErr)
	}

	req.Header.Add("Authorization", token)

	// req.Body is assumed to be closed by the base RoundTripper.
	reqBodyClosed = true

	return t.base.RoundTrip(req)
}
