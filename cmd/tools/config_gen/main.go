package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/server"

	config "gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
	viper "gitlab.com/verygoodsoftwarenotvirus/todo/internal/config/viper"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/passwords"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search"
	audit "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/frontend"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/webhooks"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads/storage"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

const (
	defaultPort              = 8888
	defaultCookieDomain      = "localhost"
	debugCookieSecret        = "HEREISA32CHARSECRETWHICHISMADEUP"
	devPostgresDBConnDetails = "postgres://dbuser:hunter2@database:5432/todo?sslmode=disable"
	devSqliteConnDetails     = "/tmp/db"
	devMariaDBConnDetails    = "dbuser:hunter2@tcp(database:3306)/todo"
	defaultCookieName        = auth.DefaultCookieName

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

	localCookies = auth.CookieConfig{
		Name:       defaultCookieName,
		Domain:     defaultCookieDomain,
		HashKey:    debugCookieSecret,
		SigningKey: debugCookieSecret,
		Lifetime:   auth.DefaultCookieLifetime,
		SecureOnly: false,
	}

	localTracingConfig = tracing.Config{
		Provider:                  "jaeger",
		SpanCollectionProbability: 1,
		Jaeger: &tracing.JaegerConfig{
			CollectorEndpoint: "http://tracing-server:14268/api/traces",
			ServiceName:       "todo-service",
		},
	}
)

type configFunc func(filePath string) error

var files = map[string]configFunc{
	"environments/local/config.toml":                                    localDevelopmentConfig,
	"environments/testing/config_files/frontend-tests.toml":             frontendTestsConfig,
	"environments/testing/config_files/integration-tests-postgres.toml": buildIntegrationTestForDBImplementation(postgres, devPostgresDBConnDetails),
	"environments/testing/config_files/integration-tests-sqlite.toml":   buildIntegrationTestForDBImplementation(sqlite, devSqliteConnDetails),
	"environments/testing/config_files/integration-tests-mariadb.toml":  buildIntegrationTestForDBImplementation(mariadb, devMariaDBConnDetails),
}

func buildLocalFrontendServiceConfig() frontend.Config {
	return frontend.Config{
		UseFakeData: false,
	}
}

func mustHashPass(password string) string {
	hashed, err := passwords.ProvideArgon2Authenticator(logging.NewNonOperationalLogger()).
		HashPassword(context.Background(), password)

	if err != nil {
		panic(err)
	}

	return hashed
}

func generatePASETOKey() []byte {
	b := make([]byte, pasetoSecretSize)

	// Note that err == nil only if we read len(b) bytes.
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return b
}

func localDevelopmentConfig(filePath string) error {
	cfg := &config.ServerConfig{
		Meta: config.MetaSettings{
			Debug:   true,
			RunMode: developmentEnv,
		},
		Encoding: encoding.Config{
			ContentType: contentTypeJSON,
		},
		Server:   localServer,
		Frontend: buildLocalFrontendServiceConfig(),
		Auth: auth.Config{
			PASETO: auth.PASETOConfig{
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
			Provider:       "bleve",
			ItemsIndexPath: "/search_indices/items.bleve",
		},
		Webhooks: webhooks.Config{
			Debug:   true,
			Enabled: false,
		},
		AuditLog: audit.Config{
			Debug:   true,
			Enabled: true,
		},
	}

	vConfig, err := viper.FromConfig(cfg)
	if err != nil {
		return fmt.Errorf("converting config object: %w", err)
	}

	if writeErr := vConfig.WriteConfigAs(filePath); writeErr != nil {
		return fmt.Errorf("writing developmentEnv config: %w", writeErr)
	}

	return nil
}

func frontendTestsConfig(filePath string) error {
	cfg := &config.ServerConfig{
		Meta: config.MetaSettings{
			Debug:   false,
			RunMode: developmentEnv,
		},
		Encoding: encoding.Config{
			ContentType: contentTypeJSON,
		},
		Server:   localServer,
		Frontend: buildLocalFrontendServiceConfig(),
		Auth: auth.Config{
			PASETO: auth.PASETOConfig{
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
			Provider:       "bleve",
			ItemsIndexPath: defaultItemsSearchIndexPath,
		},
		Webhooks: webhooks.Config{
			Debug:   true,
			Enabled: false,
		},
		AuditLog: audit.Config{
			Debug:   true,
			Enabled: true,
		},
	}

	vConfig, err := viper.FromConfig(cfg)
	if err != nil {
		return fmt.Errorf("converting config object: %w", err)
	}

	if writeErr := vConfig.WriteConfigAs(filePath); writeErr != nil {
		return fmt.Errorf("writing developmentEnv config: %w", writeErr)
	}

	return nil
}

func coverageConfig(filePath string) error {
	cfg := &config.ServerConfig{
		Meta: config.MetaSettings{
			Debug:   true,
			RunMode: testingEnv,
		},
		Encoding: encoding.Config{
			ContentType: contentTypeJSON,
		},
		Server:   localServer,
		Frontend: buildLocalFrontendServiceConfig(),
		Auth: auth.Config{
			PASETO: auth.PASETOConfig{
				Issuer:       "todo_service",
				Lifetime:     defaultPASETOLifetime,
				LocalModeKey: examplePASETOKey,
			},
			Cookies:               localCookies,
			Debug:                 false,
			EnableUserSignup:      true,
			MinimumUsernameLength: 4,
			MinimumPasswordLength: 8,
		},
		Database: dbconfig.Config{
			Debug:                     false,
			RunMigrations:             true,
			Provider:                  postgres,
			ConnectionDetails:         devPostgresDBConnDetails,
			MetricsCollectionInterval: 2 * time.Second,
			MaxPingAttempts:           maxAttempts,
			CreateTestUser: &types.TestUserCreationConfig{
				Username:       "exampleUser",
				Password:       "integration-tests-are-cool",
				HashedPassword: mustHashPass("integration-tests-are-cool"),
				IsServiceAdmin: false,
			},
		},
		Observability: observability.Config{
			Metrics: metrics.Config{
				Provider:                         "",
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
			Provider:       "bleve",
			ItemsIndexPath: defaultItemsSearchIndexPath,
		},
		Webhooks: webhooks.Config{
			Debug:   false,
			Enabled: true,
		},
		AuditLog: audit.Config{
			Debug:   false,
			Enabled: true,
		},
	}

	vConfig, err := viper.FromConfig(cfg)
	if err != nil {
		return fmt.Errorf("converting config object: %w", err)
	}

	if writeErr := vConfig.WriteConfigAs(filePath); writeErr != nil {
		return fmt.Errorf("writing developmentEnv config: %w", writeErr)
	}

	return nil
}

func buildIntegrationTestForDBImplementation(dbVendor, dbDetails string) configFunc {
	return func(filePath string) error {
		startupDeadline := time.Minute
		if dbVendor == mariadb {
			startupDeadline = 5 * time.Minute
		}

		cfg := &config.ServerConfig{
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
			Frontend: buildLocalFrontendServiceConfig(),
			Auth: auth.Config{
				PASETO: auth.PASETOConfig{
					Issuer:       "todo_service",
					Lifetime:     defaultPASETOLifetime,
					LocalModeKey: examplePASETOKey,
				},
				Cookies: auth.CookieConfig{
					Name:       defaultCookieName,
					Domain:     defaultCookieDomain,
					SigningKey: debugCookieSecret,
					Lifetime:   auth.DefaultCookieLifetime,
					SecureOnly: false,
				},
				Debug:                 false,
				EnableUserSignup:      true,
				MinimumUsernameLength: 4,
				MinimumPasswordLength: 8,
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
				Provider:       "bleve",
				ItemsIndexPath: defaultItemsSearchIndexPath,
			},
			Webhooks: webhooks.Config{
				Debug:   false,
				Enabled: false,
			},
			AuditLog: audit.Config{
				Debug:   false,
				Enabled: true,
			},
		}

		vConfig, err := viper.FromConfig(cfg)
		if err != nil {
			return fmt.Errorf("converting config object: %w", err)
		}

		if writeErr := vConfig.WriteConfigAs(filePath); writeErr != nil {
			return fmt.Errorf("writing developmentEnv config: %w", writeErr)
		}

		return nil
	}
}

func main() {
	for filePath, fun := range files {
		if err := fun(filePath); err != nil {
			log.Fatalf("error rendering %s: %v", filePath, err)
		}
	}
}
