package config

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	httpserver "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/server/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/audit"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/frontend"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerConfig_EncodeToFile(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		t.Parallel()

		cfg := &ServerConfig{
			Server: httpserver.Config{
				HTTPPort:        1234,
				Debug:           false,
				StartupDeadline: time.Minute,
			},
			AuditLog: audit.Config{
				Enabled: true,
			},
			Meta: MetaSettings{
				RunMode: DevelopmentRunMode,
			},
			Encoding: encoding.Config{
				ContentType: "application/json",
			},
			Auth: authservice.Config{
				Cookies: authservice.CookieConfig{
					Name:     "todocookie",
					Domain:   "https://verygoodsoftwarenotvirus.ru",
					Lifetime: time.Second,
				},
				MinimumUsernameLength: 4,
				MinimumPasswordLength: 8,
				EnableUserSignup:      true,
			},
			Observability: observability.Config{
				Metrics: metrics.Config{
					Provider:                         "",
					RouteToken:                       "",
					RuntimeMetricsCollectionInterval: 2 * time.Second,
				},
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
				ConnectionDetails:         database.ConnectionDetails("postgres://username:passwords@host/table"),
			},
		}

		f, err := ioutil.TempFile("", "")
		require.NoError(t, err)

		assert.NoError(t, cfg.EncodeToFile(f.Name(), json.Marshal))
	})
}
