package httpclient

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	pasetoBasePath        = "paseto"
	signatureHeaderKey    = "Signature"
	validClientSecretSize = 128
)

func setSignatureForRequest(req *http.Request, body, secretKey []byte) error {
	if len(secretKey) < validClientSecretSize {
		return fmt.Errorf("invalid secret key length: %d", len(secretKey))
	}

	mac := hmac.New(sha256.New, secretKey)
	if _, macWriteErr := mac.Write(body); macWriteErr != nil {
		return fmt.Errorf("error writing hash content: %w", macWriteErr)
	}

	req.Header.Set(signatureHeaderKey, base64.RawURLEncoding.EncodeToString(mac.Sum(nil)))

	return nil
}

// BuildDelegatedClientAuthTokenRequest builds a request.
func (c *Client) BuildDelegatedClientAuthTokenRequest(ctx context.Context, input *types.PASETOCreationInput, secretKey []byte) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.buildVersionlessURL(nil, pasetoBasePath)

	tracing.AttachRequestURIToSpan(span, uri)

	req, requestBuildErr := c.buildDataRequest(ctx, http.MethodPost, uri, input)
	if requestBuildErr != nil {
		return nil, fmt.Errorf("error building request: %w", requestBuildErr)
	}

	if signErr := setSignatureForRequest(req, c.encoderDecoder.MustJSON(input), secretKey); signErr != nil {
		return nil, signErr
	}

	return req, nil
}

func (c *Client) fetchDelegatedClientAuthToken(ctx context.Context, httpclient *http.Client, clientID string, secretKey []byte) (string, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	input := &types.PASETOCreationInput{
		ClientID:    clientID,
		RequestTime: time.Now().UTC().UnixNano(),
	}

	req, err := c.BuildDelegatedClientAuthTokenRequest(ctx, input, secretKey)
	if err != nil {
		return "", err
	}

	// use the default client here because we want a transport that doesn't worry about cookies or tokens.
	res, err := c.executeRawRequest(ctx, httpclient, req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}

	if resErr := errorFromResponse(res); resErr != nil {
		return "", resErr
	}

	var tokenRes types.PASETOResponse

	if unmarshalErr := c.unmarshalBody(ctx, res, &tokenRes); unmarshalErr != nil {
		return "", unmarshalErr
	}

	return tokenRes.Token, nil
}
