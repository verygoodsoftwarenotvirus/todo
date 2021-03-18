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
	logger    logging.Logger
	tracer    tracing.Tracer
	base      http.RoundTripper
	client    *Client
	clientID  string
	secretKey []byte // base is the base RoundTripper used to make HTTP requests. If nil, http.DefaultTransport is used.

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
					tracing.AttachErrorToSpan(span, err)
					t.logger.Error(err, "closing response body")
				}
			}
		}()
	}

	token, tokenRetrievalErr := t.client.fetchAuthTokenForAPIClient(ctx, http.DefaultClient, t.clientID, t.secretKey)
	if tokenRetrievalErr != nil {
		tracing.AttachErrorToSpan(span, tokenRetrievalErr)
		return nil, fmt.Errorf("fetching prerequisite PASETO: %w", tokenRetrievalErr)
	}

	req.Header.Add("Authorization", token)

	// req.Body is assumed to be closed by the base RoundTripper.
	reqBodyClosed = true

	res, err := t.base.RoundTrip(req)
	if err != nil {
		tracing.AttachErrorToSpan(span, err)
		return nil, fmt.Errorf("executing PASETO-authorized request: %w", tokenRetrievalErr)
	}

	return res, nil
}
