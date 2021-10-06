package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetaSettings_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("testing mode", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := MetaSettings{
			RunMode: TestingRunMode,
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("development mode", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := MetaSettings{
			RunMode: DevelopmentRunMode,
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("production mode", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := MetaSettings{
			RunMode: ProductionRunMode,
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("invalid mode", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := MetaSettings{
			RunMode: runMode(t.Name()),
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})
}
