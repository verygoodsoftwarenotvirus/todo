package events

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
)

func buildTestPublisher(t *testing.T) Publisher {
	t.Helper()

	ctx := context.Background()
	logger := logging.NewNoopLogger()
	cfg := &Config{
		Provider: ProviderMemory,
		Topic:    t.Name(),
		Enabled:  true,
	}

	p, err := ProvidePublisher(ctx, logger, cfg)
	require.NoError(t, err)
	require.NotNil(t, p)

	return p
}

func TestProvidePublisher(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()
		cfg := &Config{
			Provider: ProviderMemory,
			Topic:    t.Name(),
			Enabled:  true,
		}

		p, err := ProvidePublisher(ctx, logger, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, p)
	})

	T.Run("with nil config", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()

		p, err := ProvidePublisher(ctx, logger, nil)
		assert.Error(t, err)
		assert.Nil(t, p)
	})

	T.Run("with disabled config", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()
		cfg := &Config{
			Provider: ProviderMemory,
			Topic:    t.Name(),
			Enabled:  false,
		}

		p, err := ProvidePublisher(ctx, logger, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, p)
	})

	T.Run("with error initializing topic", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNoopLogger()
		cfg := &Config{
			Provider: "",
			Topic:    t.Name(),
			Enabled:  true,
		}

		p, err := ProvidePublisher(ctx, logger, cfg)
		assert.Error(t, err)
		assert.Nil(t, p)
	})
}

func Test_publisher_PublishEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		p := buildTestPublisher(t)
		x := struct {
			Name string `json:"name"`
		}{
			Name: t.Name(),
		}

		err := p.PublishEvent(ctx, &x, nil)
		assert.NoError(t, err)
	})

	T.Run("with invalid data", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		p := buildTestPublisher(t)
		x := struct {
			Name json.Number `json:"name"`
		}{
			Name: json.Number(t.Name()),
		}

		err := p.PublishEvent(ctx, &x, nil)
		assert.Error(t, err)
	})
}

func TestNoopEventPublisher_PublishEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		//
	})
}
