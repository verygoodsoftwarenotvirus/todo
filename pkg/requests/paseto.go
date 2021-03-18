package requests

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

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
		return fmt.Errorf("writing hash content: %w", macWriteErr)
	}

	req.Header.Set(signatureHeaderKey, base64.RawURLEncoding.EncodeToString(mac.Sum(nil)))

	return nil
}

// BuildAPIClientAuthTokenRequest builds a request.
func (b *Builder) BuildAPIClientAuthTokenRequest(ctx context.Context, input *types.PASETOCreationInput, secretKey []byte) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		b.logger.Error(validationErr, "validating input")
		return nil, fmt.Errorf("validating input: %w", validationErr)
	}

	uri := b.buildVersionlessURL(ctx, nil, pasetoBasePath)

	tracing.AttachRequestURIToSpan(span, uri)

	req, requestBuildErr := b.buildDataRequest(ctx, http.MethodPost, uri, input)
	if requestBuildErr != nil {
		return nil, fmt.Errorf("building request: %w", requestBuildErr)
	}

	var buff bytes.Buffer

	if err := json.NewEncoder(&buff).Encode(input); err != nil {
		b.logger.Error(err, "marshaling to JSON: %v")
	}

	if signErr := setSignatureForRequest(req, buff.Bytes(), secretKey); signErr != nil {
		return nil, signErr
	}

	return req, nil
}
