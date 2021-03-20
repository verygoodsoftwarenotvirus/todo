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

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
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
	if _, err := mac.Write(body); err != nil {
		return fmt.Errorf("writing hash content: %w", err)
	}

	req.Header.Set(signatureHeaderKey, base64.RawURLEncoding.EncodeToString(mac.Sum(nil)))

	return nil
}

// BuildAPIClientAuthTokenRequest builds a request that fetches a PASETO from the service.
func (b *Builder) BuildAPIClientAuthTokenRequest(ctx context.Context, input *types.PASETOCreationInput, secretKey []byte) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil || len(secretKey) == 0 {
		return nil, ErrNilInputProvided
	}

	uri := b.buildVersionlessURL(ctx, nil, pasetoBasePath)
	logger := b.logger.WithValue(keys.AccountIDKey, input.AccountID).
		WithValue(keys.APIClientClientIDKey, input.ClientID)

	tracing.AttachRequestURIToSpan(span, uri)

	if err := input.Validate(ctx); err != nil {
		return nil, prepareError(err, logger, span, "validating input")
	}

	req, err := b.buildDataRequest(ctx, http.MethodPost, uri, input)
	if err != nil {
		return nil, prepareError(err, logger, span, "building request")
	}

	var buffer bytes.Buffer
	if err = json.NewEncoder(&buffer).Encode(input); err != nil {
		return nil, prepareError(err, logger, span, "encoding body")
	}

	if err = setSignatureForRequest(req, buffer.Bytes(), secretKey); err != nil {
		return nil, prepareError(err, logger, span, "signing request")
	}

	logger.Debug("PASETO request built")

	return req, nil
}
