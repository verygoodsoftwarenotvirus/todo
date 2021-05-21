package viper

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/server"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/frontend"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/storage"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildViperConfig(t *testing.T) {
	t.Parallel()

	actual := BuildViperConfig()
	assert.NotNil(t, actual)
}

func TestFromConfig(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleConfig := &config.ServerConfig{
			Server: server.Config{
				HTTPPort:        1234,
				Debug:           false,
				StartupDeadline: time.Minute,
			},
			AuditLog: audit.Config{
				Enabled: true,
			},
			Meta: config.MetaSettings{
				RunMode: config.DevelopmentRunMode,
			},
			Encoding: encoding.Config{
				ContentType: "application/json",
			},
			Auth: authentication.Config{
				Cookies: authentication.CookieConfig{
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
				Tracing: tracing.Config{
					Jaeger: &tracing.JaegerConfig{
						CollectorEndpoint: "things",
						ServiceName:       "stuff",
					},
					Provider:                  "blah",
					SpanCollectionProbability: 0,
				},
			},
			Uploads: uploads.Config{
				Storage: storage.Config{
					FilesystemConfig: &storage.FilesystemConfig{RootDirectory: "/blah"},
					AzureConfig: &storage.AzureConfig{
						BucketName: "blahs",
						Retrying:   &storage.AzureRetryConfig{},
					},
					GCSConfig: &storage.GCSConfig{
						ServiceAccountKeyFilepath: "/blah/blah",
						BucketName:                "blah",
						Scopes:                    nil,
					},
					S3Config:          &storage.S3Config{BucketName: "blahs"},
					BucketName:        "blahs",
					UploadFilenameKey: "blahs",
					Provider:          "blahs",
				},
				Debug: false,
			},
			Frontend: frontend.Config{
				//
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
				CreateTestUser: &types.TestUserCreationConfig{
					Username:       "username",
					Password:       "password",
					HashedPassword: "hashashashashash",
					IsServiceAdmin: false,
				}},
		}

		actual, err := FromConfig(exampleConfig)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		actual, err := FromConfig(nil)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid config", func(t *testing.T) {
		t.Parallel()

		exampleConfig := &config.ServerConfig{}

		actual, err := FromConfig(exampleConfig)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestParseConfigFile(T *testing.T) {
	T.Parallel()

	ctx := context.Background()
	logger := logging.NewNonOperationalLogger()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		tf, err := ioutil.TempFile(os.TempDir(), "*.json")
		require.NoError(t, err)
		filename := tf.Name()

		exampleConfig := &config.ServerConfig{
			Server: server.Config{
				HTTPPort:        1234,
				Debug:           false,
				StartupDeadline: time.Minute,
			},
			AuditLog: audit.Config{
				Enabled: true,
			},
			Meta: config.MetaSettings{
				RunMode: config.DevelopmentRunMode,
			},
			Encoding: encoding.Config{
				ContentType: "application/json",
			},
			Auth: authentication.Config{
				Cookies: authentication.CookieConfig{
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
				//
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
http_port = "blah"
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

	T.Run("with test user creation on production error", func(t *testing.T) {
		t.Parallel()

		tf, err := ioutil.TempFile(os.TempDir(), "*.json")
		require.NoError(t, err)
		filename := tf.Name()

		exampleConfig := &config.ServerConfig{
			Server: server.Config{
				HTTPPort:        1234,
				Debug:           false,
				StartupDeadline: time.Minute,
			},
			AuditLog: audit.Config{
				Enabled: true,
			},
			Meta: config.MetaSettings{
				RunMode: config.ProductionRunMode,
			},
			Encoding: encoding.Config{
				ContentType: "application/json",
			},
			Auth: authentication.Config{
				Cookies: authentication.CookieConfig{
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
				//
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
				CreateTestUser: &types.TestUserCreationConfig{
					Username:       "username",
					Password:       "password",
					HashedPassword: "blahblahblah",
					IsServiceAdmin: false,
				},
			},
		}

		require.NoError(t, exampleConfig.EncodeToFile(filename, json.Marshal))

		cfg, err := ParseConfigFile(ctx, logger, filename)
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	T.Run("with validation error", func(t *testing.T) {
		t.Parallel()

		tf, err := ioutil.TempFile(os.TempDir(), "*.toml")
		require.NoError(t, err)

		_, err = tf.Write([]byte(`
[server]
http_port = 8888
`))
		require.NoError(t, err)

		cfg, err := ParseConfigFile(ctx, logger, tf.Name())
		assert.Error(t, err)
		assert.Nil(t, cfg)

		assert.NoError(t, os.Remove(tf.Name()))
	})
}
