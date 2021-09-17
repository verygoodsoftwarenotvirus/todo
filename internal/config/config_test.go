package config

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/server"

	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"
	itemsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/items"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerConfig_EncodeToFile(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &InstanceConfig{
			Server: server.Config{
				HTTPPort:        1234,
				Debug:           false,
				StartupDeadline: time.Minute,
			},
			Meta: MetaSettings{
				RunMode: DevelopmentRunMode,
			},
			Encoding: encoding.Config{
				ContentType: "application/json",
			},
			Observability: observability.Config{
				Metrics: metrics.Config{
					Provider:                         "",
					RouteToken:                       "",
					RuntimeMetricsCollectionInterval: 2 * time.Second,
				},
			},
			Services: ServicesConfigurations{
				Auth: authservice.Config{
					Cookies: authservice.CookieConfig{
						Name:     "todo_cookie",
						Domain:   "https://verygoodsoftwarenotvirus.ru",
						Lifetime: time.Second,
					},
					MinimumUsernameLength: 4,
					MinimumPasswordLength: 8,
					EnableUserSignup:      true,
				},
				Items: itemsservice.Config{
					SearchIndexPath: "/items_index_path",
				},
			},
			Database: config.Config{
				Provider:          "postgres",
				Debug:             true,
				RunMigrations:     true,
				ConnectionDetails: database.ConnectionDetails("postgres://username:passwords@host/table"),
			},
		}

		f, err := ioutil.TempFile("", "")
		require.NoError(t, err)

		assert.NoError(t, cfg.EncodeToFile(f.Name(), json.Marshal))
	})

	T.Run("with error marshaling", func(t *testing.T) {
		t.Parallel()

		cfg := &InstanceConfig{}

		f, err := ioutil.TempFile("", "")
		require.NoError(t, err)

		assert.Error(t, cfg.EncodeToFile(f.Name(), func(interface{}) ([]byte, error) {
			return nil, errors.New("blah")
		}))
	})
}

func TestServerConfig_ProvideDatabaseClient(T *testing.T) {
	T.Run("supported providers", func(t *testing.T) {
		ctx := context.Background()
		logger := logging.NewNoopLogger()

		for _, provider := range []string{"postgres", "mysql"} {
			cfg := &InstanceConfig{
				Database: config.Config{
					Provider: provider,
				},
			}

			x, err := ProvideDatabaseClient(ctx, logger, cfg)
			assert.NotNil(t, x)
			assert.NoError(t, err)
		}
	})

	T.Run("with invalid provider", func(t *testing.T) {
		ctx := context.Background()
		logger := logging.NewNoopLogger()

		cfg := &InstanceConfig{
			Database: config.Config{
				Provider: "provider",
			},
		}

		x, err := ProvideDatabaseClient(ctx, logger, cfg)
		assert.Nil(t, x)
		assert.Error(t, err)
	})
}
