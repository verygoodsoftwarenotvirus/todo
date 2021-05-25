package secrets

import (
	"context"
	"encoding/json"
	"errors"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"gocloud.dev/secrets"
	_ "gocloud.dev/secrets/gcpkms"
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
		Encrypt(ctx context.Context, value interface{}) ([]byte, error)
		Decrypt(ctx context.Context, content []byte, v interface{}) error
	}

	secretManager struct {
		logger logging.Logger
		tracer tracing.Tracer
		keeper *secrets.Keeper
	}
)

// ProvideSecretManager builds a new SecretManager
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

func (sm *secretManager) Encrypt(ctx context.Context, value interface{}) ([]byte, error) {
	ctx, span := sm.tracer.StartSpan(ctx)
	defer span.End()

	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	return sm.keeper.Encrypt(ctx, jsonBytes)
}

func (sm *secretManager) Decrypt(ctx context.Context, content []byte, v interface{}) error {
	ctx, span := sm.tracer.StartSpan(ctx)
	defer span.End()

	jsonBytes, err := sm.keeper.Decrypt(ctx, content)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(jsonBytes, &v); err != nil {
		return err
	}

	return nil
}
