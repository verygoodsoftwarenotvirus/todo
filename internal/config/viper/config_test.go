package viper

import (
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/capitalism"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/server"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/audit"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/storage"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/stretchr/testify/assert"
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

		exampleConfig := &config.InstanceConfig{
			Server: server.Config{
				HTTPPort:        1234,
				Debug:           false,
				StartupDeadline: time.Minute,
			},
			Meta: config.MetaSettings{
				RunMode: config.DevelopmentRunMode,
			},
			Encoding: encoding.Config{
				ContentType: "application/json",
			},
			Capitalism: capitalism.Config{
				Enabled:  false,
				Provider: capitalism.StripeProvider,
				Stripe: &capitalism.StripeConfig{
					APIKey:        "whatever",
					SuccessURL:    "whatever",
					CancelURL:     "whatever",
					WebhookSecret: "whatever",
				},
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
			Services: config.ServicesConfigurations{
				AuditLog: audit.Config{
					Enabled: true,
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
				Items: items.Config{
					SearchIndexPath: "/items_index_path",
					Logger: logging.Config{
						Name:     "items",
						Level:    logging.InfoLevel,
						Provider: logging.ProviderZerolog,
					},
				},
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

		exampleConfig := &config.InstanceConfig{}

		actual, err := FromConfig(exampleConfig)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}
