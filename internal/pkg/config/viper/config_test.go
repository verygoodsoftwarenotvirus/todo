package viper

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"time"

	httpserver "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/server/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/audit"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/frontend"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func Test_RandString(t *testing.T) {
	t.Parallel()

	actual := config.RandString()
	assert.NotEmpty(t, actual)
	assert.Len(t, actual, 32)
}

func TestBuildConfig(t *testing.T) {
	t.Parallel()

	actual := BuildViperConfig()
	assert.NotNil(t, actual)
}

func TestParseConfigFile(T *testing.T) {
	T.Parallel()

	ctx := context.Background()
	logger := noop.NewLogger()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		tf, err := ioutil.TempFile(os.TempDir(), "*.json")
		require.NoError(t, err)
		filename := tf.Name()

		exampleConfig := &config.ServerConfig{
			Server: httpserver.Config{
				HTTPPort: 1234,
				Debug:    false,
			},
			AuditLog: audit.Config{
				Enabled: true,
			},
			Meta: config.MetaSettings{
				StartupDeadline: time.Minute,
				RunMode:         "development",
			},
			Auth: authservice.Config{
				CookieDomain:          "https://verygoodsoftwarenotvirus.ru",
				CookieLifetime:        time.Second,
				MinimumUsernameLength: 4,
				MinimumPasswordLength: 8,
				EnableUserSignup:      true,
			},
			Observability: observability.Config{
				RuntimeMetricsCollectionInterval: 2 * time.Second,
			},
			Frontend: frontend.Config{
				StaticFilesDirectory: "/static",
			},
			Search: search.Config{
				ItemsIndexPath: "/items_index_path",
			},
			Database: dbconfig.Config{
				Provider:                  "postgres",
				MetricsCollectionInterval: 2 * time.Second,
				Debug:                     true,
				RunMigrations:             true,
				ConnectionDetails:         database.ConnectionDetails("postgres://username:password@host/table"),
			},
		}

		require.NoError(t, exampleConfig.EncodeToFile(filename, json.Marshal))

		cfg, err := ParseConfigFile(ctx, logger, filename)
		require.NoError(t, err)

		assert.Equal(t, exampleConfig, cfg)

		assert.NoError(t, os.Remove(tf.Name()))
	})

	T.Run("unparseable garbage", func(t *testing.T) {
		t.Parallel()
		tf, err := ioutil.TempFile(os.TempDir(), "*.toml")
		require.NoError(t, err)

		_, err = tf.Write([]byte(`
[server]
http_port = "fart"
debug = ":banana:"
`))
		require.NoError(t, err)

		cfg, err := ParseConfigFile(ctx, logger, tf.Name())
		assert.Error(t, err)
		assert.Nil(t, cfg)

		assert.NoError(t, os.Remove(tf.Name()))
	})

	T.Run("with nonexistent file", func(t *testing.T) {
		t.Parallel()
		cfg, err := ParseConfigFile(ctx, logger, "/this/doesn't/even/exist/lol")
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})
}
