package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/secrets"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/server"
	auditservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/audit"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"
	frontendservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/frontend"
	itemsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/items"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/webhooks"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/storage"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

const (
	defaultPort              = 8888
	defaultCookieDomain      = "localhost"
	debugCookieSecret        = "HEREISA32CHARSECRETWHICHISMADEUP"
	devPostgresDBConnDetails = "postgres://dbuser:hunter2@database:5432/todo?sslmode=disable"
	devSqliteConnDetails     = "/tmp/db"
	devMariaDBConnDetails    = "dbuser:hunter2@tcp(database:3306)/todo"
	defaultCookieName        = authservice.DefaultCookieName

	// run modes.
	developmentEnv = "development"
	testingEnv     = "testing"

	// database providers.
	postgres = "postgres"
	sqlite   = "sqlite"
	mariadb  = "mariadb"

	// test user stuff.
	defaultPassword = "password"

	// search index paths.
	defaultItemsSearchIndexPath = "items.bleve"

	pasetoSecretSize      = 32
	maxAttempts           = 50
	defaultPASETOLifetime = 1 * time.Minute

	contentTypeJSON = "application/json"
)

var (
	examplePASETOKey = generatePASETOKey()

	noopTracingConfig = tracing.Config{
		Provider:                  "",
		SpanCollectionProbability: 1,
	}

	localServer = server.Config{
		Debug:           true,
		HTTPPort:        defaultPort,
		StartupDeadline: time.Minute,
	}

	localCookies = authservice.CookieConfig{
		Name:       defaultCookieName,
		Domain:     defaultCookieDomain,
		HashKey:    debugCookieSecret,
		SigningKey: debugCookieSecret,
		Lifetime:   authservice.DefaultCookieLifetime,
		SecureOnly: false,
	}

	localTracingConfig = tracing.Config{
		Provider:                  "jaeger",
		SpanCollectionProbability: 1,
		Jaeger: &tracing.JaegerConfig{
			CollectorEndpoint: "http://tracing-server:14268/api/traces",
			ServiceName:       "todo_service",
		},
	}
)

func initializeLocalSecretManager(ctx context.Context) secrets.SecretManager {
	logger := logging.NewNoopLogger()

	cfg := &secrets.Config{
		Provider: secrets.ProviderLocal,
		Key:      "SUFNQVdBUkVUSEFUVEhJU1NFQ1JFVElTVU5TRUNVUkU=",
	}

	k, err := secrets.ProvideSecretKeeper(ctx, cfg)
	if err != nil {
		panic(err)
	}

	sm, err := secrets.ProvideSecretManager(logger, k)
	if err != nil {
		panic(err)
	}

	return sm
}

func encryptAndSaveConfig(ctx context.Context, outputPath string, cfg *config.InstanceConfig) error {
	sm := initializeLocalSecretManager(ctx)
	output, err := sm.Encrypt(ctx, cfg)
	if err != nil {
		return fmt.Errorf("encrypting config: %v", err)
	}

	return os.WriteFile(outputPath, []byte(output), 0644)
}

type configFunc func(ctx context.Context, filePath string) error

var files = map[string]configFunc{
	"environments/local/service.config":                                   localDevelopmentConfig,
	"environments/testing/config_files/frontend-tests.config":             frontendTestsConfig,
	"environments/testing/config_files/integration-tests-postgres.config": buildIntegrationTestForDBImplementation(postgres, devPostgresDBConnDetails),
	"environments/testing/config_files/integration-tests-sqlite.config":   buildIntegrationTestForDBImplementation(sqlite, devSqliteConnDetails),
	"environments/testing/config_files/integration-tests-mariadb.config":  buildIntegrationTestForDBImplementation(mariadb, devMariaDBConnDetails),
}

func buildLocalFrontendServiceConfig() frontendservice.Config {
	return frontendservice.Config{
		UseFakeData: false,
	}
}

func mustHashPass(password string) string {
	hashed, err := authentication.ProvideArgon2Authenticator(logging.NewNoopLogger()).
		HashPassword(context.Background(), password)

	if err != nil {
		panic(err)
	}

	return hashed
}

func generatePASETOKey() []byte {
	b := make([]byte, pasetoSecretSize)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return b
}

func localDevelopmentConfig(ctx context.Context, filePath string) error {
	cfg := &config.InstanceConfig{
		Meta: config.MetaSettings{
			Debug:   true,
			RunMode: developmentEnv,
		},
		Encoding: encoding.Config{
			ContentType: contentTypeJSON,
		},
		Server: localServer,
		Database: dbconfig.Config{
			Debug:                     true,
			RunMigrations:             true,
			MaxPingAttempts:           maxAttempts,
			Provider:                  postgres,
			ConnectionDetails:         devPostgresDBConnDetails,
			MetricsCollectionInterval: time.Second,
			CreateTestUser: &types.TestUserCreationConfig{
				Username:       "username",
				Password:       defaultPassword,
				HashedPassword: mustHashPass(defaultPassword),
				IsServiceAdmin: true,
			},
		},
		Observability: observability.Config{
			Metrics: metrics.Config{
				Provider:                         "prometheus",
				RouteToken:                       "",
				RuntimeMetricsCollectionInterval: time.Second,
			},
			Tracing: localTracingConfig,
		},
		Uploads: uploads.Config{
			Debug: true,
			Storage: storage.Config{
				UploadFilenameKey: "avatar",
				Provider:          "filesystem",
				BucketName:        "avatars",
				AzureConfig:       nil,
				GCSConfig:         nil,
				S3Config:          nil,
				FilesystemConfig: &storage.FilesystemConfig{
					RootDirectory: "/avatars",
				},
			},
		},
		Search: search.Config{
			Provider: "bleve",
		},
		Services: config.ServicesConfigurations{
			AuditLog: auditservice.Config{
				Debug:   true,
				Enabled: true,
			},
			Auth: authservice.Config{
				PASETO: authservice.PASETOConfig{
					Issuer:       "todo_service",
					Lifetime:     defaultPASETOLifetime,
					LocalModeKey: examplePASETOKey,
				},
				Cookies:               localCookies,
				Debug:                 true,
				EnableUserSignup:      true,
				MinimumUsernameLength: 4,
				MinimumPasswordLength: 8,
			},
			Frontend: buildLocalFrontendServiceConfig(),
			Webhooks: webhooksservice.Config{
				Debug:   true,
				Enabled: false,
			},
			Items: itemsservice.Config{
				SearchIndexPath: fmt.Sprintf("/search_indices/%s", defaultItemsSearchIndexPath),
				Logger: logging.Config{
					Name:     "items",
					Level:    logging.InfoLevel,
					Provider: logging.ProviderZerolog,
				},
			},
		},
	}

	return encryptAndSaveConfig(ctx, filePath, cfg)
}

func frontendTestsConfig(ctx context.Context, filePath string) error {
	cfg := &config.InstanceConfig{
		Meta: config.MetaSettings{
			Debug:   false,
			RunMode: developmentEnv,
		},
		Encoding: encoding.Config{
			ContentType: contentTypeJSON,
		},
		Server: localServer,
		Database: dbconfig.Config{
			Debug:                     true,
			RunMigrations:             true,
			Provider:                  postgres,
			ConnectionDetails:         devPostgresDBConnDetails,
			MaxPingAttempts:           maxAttempts,
			MetricsCollectionInterval: time.Second,
		},
		Observability: observability.Config{
			Metrics: metrics.Config{
				Provider:                         "prometheus",
				RouteToken:                       "",
				RuntimeMetricsCollectionInterval: time.Second,
			},
			Tracing: noopTracingConfig,
		},
		Uploads: uploads.Config{
			Debug: true,
			Storage: storage.Config{
				UploadFilenameKey: "avatar",
				Provider:          "memory",
				BucketName:        "avatars",
			},
		},
		Search: search.Config{
			Provider: "bleve",
		},
		Services: config.ServicesConfigurations{
			AuditLog: auditservice.Config{
				Debug:   true,
				Enabled: true,
			},
			Auth: authservice.Config{
				PASETO: authservice.PASETOConfig{
					Issuer:       "todo_service",
					Lifetime:     defaultPASETOLifetime,
					LocalModeKey: examplePASETOKey,
				},
				Cookies:               localCookies,
				Debug:                 true,
				EnableUserSignup:      true,
				MinimumUsernameLength: 4,
				MinimumPasswordLength: 8,
			},
			Frontend: buildLocalFrontendServiceConfig(),
			Webhooks: webhooksservice.Config{
				Debug:   true,
				Enabled: false,
			},
			Items: itemsservice.Config{
				SearchIndexPath: fmt.Sprintf("/search_indices/%s", defaultItemsSearchIndexPath),
				Logger: logging.Config{
					Name:     "items",
					Level:    logging.InfoLevel,
					Provider: logging.ProviderZerolog,
				},
			},
		},
	}

	return encryptAndSaveConfig(ctx, filePath, cfg)
}

func buildIntegrationTestForDBImplementation(dbVendor, dbDetails string) configFunc {
	return func(ctx context.Context, filePath string) error {
		startupDeadline := time.Minute
		if dbVendor == mariadb {
			startupDeadline = 5 * time.Minute
		}

		cfg := &config.InstanceConfig{
			Meta: config.MetaSettings{
				Debug:   false,
				RunMode: testingEnv,
			},
			Encoding: encoding.Config{
				ContentType: contentTypeJSON,
			},
			Server: server.Config{
				Debug:           false,
				HTTPPort:        defaultPort,
				StartupDeadline: startupDeadline,
			},
			Database: dbconfig.Config{
				Debug:                     false,
				RunMigrations:             true,
				Provider:                  dbVendor,
				MaxPingAttempts:           maxAttempts,
				MetricsCollectionInterval: 2 * time.Second,
				ConnectionDetails:         database.ConnectionDetails(dbDetails),
				CreateTestUser: &types.TestUserCreationConfig{
					Username:       "exampleUser",
					Password:       "integration-tests-are-cool",
					HashedPassword: mustHashPass("integration-tests-are-cool"),
					IsServiceAdmin: true,
				},
			},
			Observability: observability.Config{
				Metrics: metrics.Config{
					Provider:                         "",
					RouteToken:                       "",
					RuntimeMetricsCollectionInterval: time.Second,
				},
				Tracing: localTracingConfig,
			},
			Uploads: uploads.Config{
				Debug: false,
				Storage: storage.Config{
					Provider:    "memory",
					BucketName:  "avatars",
					AzureConfig: nil,
					GCSConfig:   nil,
					S3Config:    nil,
				},
			},
			Search: search.Config{
				Provider: "bleve",
			},
			Services: config.ServicesConfigurations{
				AuditLog: auditservice.Config{
					Debug:   false,
					Enabled: true,
				},
				Auth: authservice.Config{
					PASETO: authservice.PASETOConfig{
						Issuer:       "todo_service",
						Lifetime:     defaultPASETOLifetime,
						LocalModeKey: examplePASETOKey,
					},
					Cookies: authservice.CookieConfig{
						Name:       defaultCookieName,
						Domain:     defaultCookieDomain,
						SigningKey: debugCookieSecret,
						Lifetime:   authservice.DefaultCookieLifetime,
						SecureOnly: false,
					},
					Debug:                 false,
					EnableUserSignup:      true,
					MinimumUsernameLength: 4,
					MinimumPasswordLength: 8,
				},
				Frontend: buildLocalFrontendServiceConfig(),
				Webhooks: webhooksservice.Config{
					Debug:   true,
					Enabled: false,
				},
				Items: itemsservice.Config{
					SearchIndexPath: fmt.Sprintf("/search_indices/%s", defaultItemsSearchIndexPath),
					Logger: logging.Config{
						Name:     "items",
						Level:    logging.InfoLevel,
						Provider: logging.ProviderZerolog,
					},
				},
			},
		}

		return encryptAndSaveConfig(ctx, filePath, cfg)
	}
}

func main() {
	ctx := context.Background()

	for filePath, fun := range files {
		if err := fun(ctx, filePath); err != nil {
			log.Fatalf("error rendering %s: %v", filePath, err)
		}
	}
}
