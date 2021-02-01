package main

import (
	"context"
	"fmt"
	"log"
	"time"

	httpserver "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/server/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/audit"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	frontendservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/frontend"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/webhooks"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config/viper"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password/bcrypt"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/storage"
)

const (
	defaultPort              = 8888
	defaultCookieDomain      = "localhost"
	debugCookieSecret        = "HEREISA32CHARSECRETWHICHISMADEUP"
	devPostgresDBConnDetails = "postgres://dbuser:hunter2@database:5432/todo?sslmode=disable"
	devSqliteConnDetails     = "/tmp/db"
	devMariaDBConnDetails    = "dbuser:hunter2@tcp(database:3306)/todo"
	defaultFrontendFilepath  = "/frontend"
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

	maxAttempts = 50
)

type configFunc func(filePath string) error

var files = map[string]configFunc{
	"environments/local/config.toml":                                    localDevelopmentConfig,
	"environments/testing/config_files/frontend-tests.toml":             frontendTestsConfig,
	"environments/testing/config_files/coverage.toml":                   coverageConfig,
	"environments/testing/config_files/integration-tests-postgres.toml": buildIntegrationTestForDBImplementation(postgres, devPostgresDBConnDetails),
	"environments/testing/config_files/integration-tests-sqlite.toml":   buildIntegrationTestForDBImplementation(sqlite, devSqliteConnDetails),
	"environments/testing/config_files/integration-tests-mariadb.toml":  buildIntegrationTestForDBImplementation(mariadb, devMariaDBConnDetails),
}

func mustHashPass(password string) string {
	hashed, err := bcrypt.ProvideAuthenticator(bcrypt.DefaultHashCost-1, logging.NewNonOperationalLogger()).
		HashPassword(context.Background(), password)

	if err != nil {
		panic(err)
	}

	return hashed
}

func localDevelopmentConfig(filePath string) error {
	cfg := &config.ServerConfig{
		Meta: config.MetaSettings{
			Debug:   true,
			RunMode: developmentEnv,
		},
		Server: httpserver.Config{
			Debug:           true,
			HTTPPort:        defaultPort,
			StartupDeadline: time.Minute,
		},
		Frontend: frontendservice.Config{
			StaticFilesDirectory: defaultFrontendFilepath,
			Debug:                true,
			LogStaticFiles:       false,
			CacheStaticFiles:     false,
		},
		Auth: authservice.Config{
			Cookies: authservice.CookieConfig{
				Name:       defaultCookieName,
				Domain:     defaultCookieDomain,
				SigningKey: debugCookieSecret,
				Lifetime:   authservice.DefaultCookieLifetime,
				SecureOnly: false,
			},
			Debug:                 true,
			EnableUserSignup:      true,
			MinimumUsernameLength: 4,
			MinimumPasswordLength: 8,
		},
		Database: dbconfig.Config{
			Debug:         true,
			RunMigrations: true,
			CreateTestUser: &types.TestUserCreationConfig{
				Username:       "username",
				Password:       defaultPassword,
				HashedPassword: mustHashPass(defaultPassword),
				IsSiteAdmin:    true,
			},
			Provider:                  postgres,
			ConnectionDetails:         devPostgresDBConnDetails,
			MetricsCollectionInterval: time.Second,
		},
		Observability: observability.Config{
			Metrics: metrics.Config{
				Provider: "prometheus",
				RouteAuth: metrics.AuthConfig{
					Method:    "",
					BasicAuth: nil,
				},
			},
			Tracing: tracing.Config{
				Provider:                  "jaeger",
				SpanCollectionProbability: 1,
			},
			RuntimeMetricsCollectionInterval: time.Second,
		},
		Uploads: uploads.Config{
			Debug:    true,
			Provider: "filesystem",
			Storage: storage.Config{
				Provider:    "filesystem",
				BucketName:  "avatars",
				AzureConfig: nil,
				GCSConfig:   nil,
				S3Config:    nil,
				FilesystemConfig: &storage.FilesystemConfig{
					RootDirectory: "artifacts",
				},
			},
		},
		Search: search.Config{
			Provider:       "bleve",
			ItemsIndexPath: "/search_indices/items.bleve",
		},
		Webhooks: webhooksservice.Config{
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
		Server: httpserver.Config{
			Debug:           true,
			HTTPPort:        defaultPort,
			StartupDeadline: time.Minute,
		},
		Frontend: frontendservice.Config{
			StaticFilesDirectory: defaultFrontendFilepath,
			Debug:                true,
			LogStaticFiles:       false,
			CacheStaticFiles:     false,
		},
		Auth: authservice.Config{
			Cookies: authservice.CookieConfig{
				Name:       defaultCookieName,
				Domain:     defaultCookieDomain,
				SigningKey: debugCookieSecret,
				Lifetime:   authservice.DefaultCookieLifetime,
				SecureOnly: false,
			},
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
			MetricsCollectionInterval: time.Second,
		},
		Observability: observability.Config{
			Metrics: metrics.Config{
				Provider: "prometheus",
				RouteAuth: metrics.AuthConfig{
					Method:    "",
					BasicAuth: nil,
				},
			},
			Tracing: tracing.Config{
				Provider:                  "jaeger",
				SpanCollectionProbability: 1,
			},
			RuntimeMetricsCollectionInterval: time.Second,
		},
		Uploads: uploads.Config{
			Debug:    true,
			Provider: "filesystem",
			Storage: storage.Config{
				Provider:    "filesystem",
				BucketName:  "avatars",
				AzureConfig: nil,
				GCSConfig:   nil,
				S3Config:    nil,
				FilesystemConfig: &storage.FilesystemConfig{
					RootDirectory: "artifacts",
				},
			},
		},
		Search: search.Config{
			Provider:       "bleve",
			ItemsIndexPath: defaultItemsSearchIndexPath,
		},
		Webhooks: webhooksservice.Config{
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
		Server: httpserver.Config{
			Debug:           true,
			HTTPPort:        defaultPort,
			StartupDeadline: time.Minute,
		},
		Frontend: frontendservice.Config{
			StaticFilesDirectory: defaultFrontendFilepath,
			Debug:                true,
			LogStaticFiles:       false,
			CacheStaticFiles:     false,
		},
		Auth: authservice.Config{
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
		Database: dbconfig.Config{
			Debug:                     false,
			RunMigrations:             true,
			Provider:                  postgres,
			ConnectionDetails:         devPostgresDBConnDetails,
			MetricsCollectionInterval: 2 * time.Second,
			CreateTestUser: &types.TestUserCreationConfig{
				Username:       "coverageUser",
				Password:       defaultPassword,
				HashedPassword: mustHashPass(defaultPassword),
				IsSiteAdmin:    false,
			},
		},
		Observability: observability.Config{
			Metrics: metrics.Config{
				Provider: "",
				RouteAuth: metrics.AuthConfig{
					Method:    "",
					BasicAuth: nil,
				},
			},
			Tracing: tracing.Config{
				Provider:                  "",
				SpanCollectionProbability: 1,
			},
			RuntimeMetricsCollectionInterval: time.Second,
		},
		Uploads: uploads.Config{
			Debug:    true,
			Provider: "filesystem",
			Storage: storage.Config{
				Provider:    "filesystem",
				BucketName:  "avatars",
				AzureConfig: nil,
				GCSConfig:   nil,
				S3Config:    nil,
				FilesystemConfig: &storage.FilesystemConfig{
					RootDirectory: "artifacts",
				},
			},
		},
		Search: search.Config{
			Provider:       "bleve",
			ItemsIndexPath: defaultItemsSearchIndexPath,
		},
		Webhooks: webhooksservice.Config{
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

func buildIntegrationTestForDBImplementation(dbVendor, dbDetails string) configFunc {
	return func(filePath string) error {
		cfg := &config.ServerConfig{
			Meta: config.MetaSettings{
				Debug:   false,
				RunMode: testingEnv,
			},
			Server: httpserver.Config{
				Debug:    false,
				HTTPPort: defaultPort,
				StartupDeadline: func() time.Duration {
					if dbVendor == mariadb {
						return 5 * time.Minute
					}
					return time.Minute
				}(),
			},
			Frontend: frontendservice.Config{
				StaticFilesDirectory: defaultFrontendFilepath,
				Debug:                false,
				LogStaticFiles:       false,
				CacheStaticFiles:     false,
			},
			Auth: authservice.Config{
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
			Database: dbconfig.Config{
				Debug:                     false,
				RunMigrations:             true,
				Provider:                  dbVendor,
				MaxPingAttempts:           maxAttempts,
				ConnectionDetails:         database.ConnectionDetails(dbDetails),
				MetricsCollectionInterval: 2 * time.Second,
				CreateTestUser: &types.TestUserCreationConfig{
					Username:       "exampleUser",
					Password:       "integration-tests-are-cool",
					HashedPassword: mustHashPass("integration-tests-are-cool"),
					IsSiteAdmin:    true,
				},
			},
			Observability: observability.Config{
				Metrics: metrics.Config{
					Provider: "",
					RouteAuth: metrics.AuthConfig{
						Method:    "",
						BasicAuth: nil,
					},
				},
				Tracing: tracing.Config{
					Provider:                  "",
					SpanCollectionProbability: 1,
				},
				RuntimeMetricsCollectionInterval: time.Second,
			},
			Uploads: uploads.Config{
				Debug:    false,
				Provider: "filesystem",
				Storage: storage.Config{
					Provider:    "filesystem",
					BucketName:  "avatars",
					AzureConfig: nil,
					GCSConfig:   nil,
					S3Config:    nil,
					FilesystemConfig: &storage.FilesystemConfig{
						RootDirectory: "artifacts",
					},
				},
			},
			Search: search.Config{
				Provider:       "bleve",
				ItemsIndexPath: defaultItemsSearchIndexPath,
			},
			Webhooks: webhooksservice.Config{
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
