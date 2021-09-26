package secrets

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"gocloud.dev/secrets"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
)

const (
	tracerName = "secret_manager"
)

var (
	errInvalidKeeper = errors.New("invalid keeper")
)

type (
	// SecretManager manages secrets.
	SecretManager interface {
		Encrypt(ctx context.Context, value interface{}) (string, error)
		Decrypt(ctx context.Context, content string, v interface{}) error
	}

	secretManager struct {
		logger logging.Logger
		tracer tracing.Tracer
		keeper *secrets.Keeper
	}
)

// ProvideSecretManager builds a new SecretManager.
func ProvideSecretManager(logger logging.Logger, keeper *secrets.Keeper) (SecretManager, error) {
	if keeper == nil {
		return nil, errInvalidKeeper
	}

	sm := &secretManager{
		logger: logging.EnsureLogger(logger),
		tracer: tracing.NewTracer(tracerName),
		keeper: keeper,
	}

	return sm, nil
}

// Encrypt does the following:
//		1. JSON encodes a given value
//		2. encrypts that encoded data
//		3. base64 URL encodes that encrypted data
func (sm *secretManager) Encrypt(ctx context.Context, value interface{}) (string, error) {
	ctx, span := sm.tracer.StartSpan(ctx)
	defer span.End()

	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("encoding value to JSON: %w", err)
	}

	encrypted, err := sm.keeper.Encrypt(ctx, jsonBytes)
	if err != nil {
		// this can literally never occur in the local version, so it cannot be tested
		return "", fmt.Errorf("encrypting JSON encoded bytes: %w", err)
	}

	encoded := base64.URLEncoding.EncodeToString(encrypted)

	return encoded, nil
}

// Decrypt does the following:
//		1. base64 URL decodes the provided data
//		2. decrypts that encoded data
//		3. JSON decodes that decrypted data into the target variable.
func (sm *secretManager) Decrypt(ctx context.Context, content string, v interface{}) error {
	ctx, span := sm.tracer.StartSpan(ctx)
	defer span.End()

	decoded, err := base64.URLEncoding.DecodeString(content)
	if err != nil {
		return fmt.Errorf("decoding base64 encoded content: %w", err)
	}

	jsonBytes, err := sm.keeper.Decrypt(ctx, decoded)
	if err != nil {
		return fmt.Errorf("decrypting decoded bytes into JSON: %w", err)
	}

	err = json.Unmarshal(jsonBytes, &v)
	if err != nil {
		return fmt.Errorf("decoded JSON bytes into value: %w", err)
	}

	return nil
}
