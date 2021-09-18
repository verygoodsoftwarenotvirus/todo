package config

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"

	"github.com/alexedwards/scs/v2/memstore"
	"github.com/stretchr/testify/assert"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			Provider:          PostgresProvider,
			ConnectionDetails: "example_connection_string",
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}

func TestProvideSessionManager(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cookieConfig := authservice.CookieConfig{}
		mdm := &database.MockDatabase{}
		store := memstore.New()

		mdm.On("ProvideSessionStore").Return(store)

		sessionManager, err := ProvideSessionManager(cookieConfig, mdm)
		assert.NotNil(t, sessionManager)
		assert.NoError(t, err)
	})
}
