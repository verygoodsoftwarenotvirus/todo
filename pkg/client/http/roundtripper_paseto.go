package http

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
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
		base:      newDefaultRoundTripper(client.unauthenticatedClient.Timeout),
		client:    client,
	}
}

var pasetoRoundTripperClient = buildRetryingClient(&http.Client{Timeout: defaultTimeout}, logging.NewNonOperationalLogger(), tracing.NewTracer("PASETO_roundtripper"))

// RoundTrip authorizes and authenticates the request with a PASETO.
func (t *pasetoRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, span := t.tracer.StartSpan(req.Context())
	defer span.End()

	reqBodyClosed := false

	logger := t.logger.WithRequest(req)

	if req.Body != nil {
		defer func() {
			if !reqBodyClosed {
				if err := req.Body.Close(); err != nil {
					observability.AcknowledgeError(err, logger, span, "closing response body")
				}
			}
		}()
	}

	token, err := t.client.fetchAuthTokenForAPIClient(ctx, pasetoRoundTripperClient, t.clientID, t.secretKey)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "fetching prerequisite PASETO")
	}

	// req.Body is assumed to be closed by the base RoundTripper.
	reqBodyClosed = true

	req.Header.Add("Authorization", token)

	res, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "executing PASETO-authorized request")
	}

	return res, nil
}
