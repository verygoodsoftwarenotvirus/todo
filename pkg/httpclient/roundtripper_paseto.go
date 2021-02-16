package httpclient

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// pasetoRoundTripper is a transport that uses a cookie.
type pasetoRoundTripper struct {
	clientID  string
	secretKey []byte
	uri       string

	logger logging.Logger
	tracer tracing.Tracer

	// base is the base RoundTripper used to make HTTP requests. If nil, http.DefaultTransport is used.
	base       http.RoundTripper
	httpClient *http.Client
}

func newPASETORoundTripper(client *Client, clientID string, secretKey []byte) *pasetoRoundTripper {
	return &pasetoRoundTripper{
		clientID:   clientID,
		secretKey:  secretKey,
		uri:        client.buildVersionlessURL(nil, "paseto"),
		logger:     client.logger,
		tracer:     client.tracer,
		base:       otelhttp.NewTransport(newDefaultRoundTripper(client.plainClient.Timeout)),
		httpClient: client.plainClient,
	}
}

func (t *pasetoRoundTripper) buildPasetoRequest(ctx context.Context) (*http.Request, error) {
	ctx, span := t.tracer.StartSpan(ctx)
	defer span.End()

	reqBytes, jsonErr := json.Marshal(&types.PASETOCreationInput{
		ClientID:  t.clientID,
		NonceUUID: uuid.New().String(),
	})
	if jsonErr != nil {
		return nil, jsonErr
	}

	tracing.AttachRequestURIToSpan(span, t.uri)

	req, requestBuildErr := http.NewRequestWithContext(ctx, http.MethodPost, t.uri, bytes.NewReader(reqBytes))
	if requestBuildErr != nil {
		return nil, fmt.Errorf("error building request: %w", requestBuildErr)
	}

	mac := hmac.New(sha256.New, t.secretKey)
	if _, macWriteErr := mac.Write(reqBytes); macWriteErr != nil {
		return nil, fmt.Errorf("error writing hash content: %w", macWriteErr)
	}

	req.Header.Set("Signature", string(mac.Sum(nil)))

	return req, nil
}

func (t *pasetoRoundTripper) fetchPasetoToken(ctx context.Context) (string, error) {
	ctx, span := t.tracer.StartSpan(ctx)
	defer span.End()

	req, err := t.buildPasetoRequest(ctx)
	if err != nil {
		return "", err
	}

	res, err := t.httpClient.Do(req)
	if err != nil {
		return "", err
	}

	token, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	closeResponseBody(t.logger, res)

	return string(token), nil
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

	token, tokenRetrievalErr := t.fetchPasetoToken(ctx)
	if tokenRetrievalErr != nil {
		return nil, fmt.Errorf("error fetching prerequisite PASETO: %w", tokenRetrievalErr)
	}

	req.Header.Add("Authorization", token)

	// req.Body is assumed to be closed by the base RoundTripper.
	reqBodyClosed = true

	return t.base.RoundTrip(req)
}
