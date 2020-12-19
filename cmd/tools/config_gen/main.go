package main

import (
	"context"
	"fmt"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password/bcrypt"
	"log"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config/viper"
)

const (
	defaultPort              = 8888
	defaultCookieDomain      = "localhost"
	debugCookieSecret        = "HEREISA32CHARSECRETWHICHISMADEUP"
	devPostgresDBConnDetails = "postgres://dbuser:hunter2@database:5432/todo?sslmode=disable"
	devSqliteConnDetails     = "/tmp/db"
	devMariaDBConnDetails    = "dbuser:hunter2@tcp(database:3306)/todo"
	defaultFrontendFilepath  = "/frontend"

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
	hashed, err := bcrypt.ProvideAuthenticator(bcrypt.DefaultHashCost-1, noop.NewLogger()).
		HashPassword(context.Background(), password)

	if err != nil {
		panic(err)
	}

	return hashed
}

func localDevelopmentConfig(filePath string) error {
	cfg := viper.BuildViperConfig()

	cfg.Set(viper.ConfigKeyMetaRunMode, developmentEnv)
	cfg.Set(viper.ConfigKeyMetaDebug, true)
	cfg.Set(viper.ConfigKeyMetaStartupDeadline, time.Minute)

	cfg.Set(viper.ConfigKeyServerHTTPPort, defaultPort)
	cfg.Set(viper.ConfigKeyServerDebug, true)

	cfg.Set(viper.ConfigKeyFrontendDebug, true)
	cfg.Set(viper.ConfigKeyFrontendStaticFilesDir, defaultFrontendFilepath)
	cfg.Set(viper.ConfigKeyFrontendCacheStatics, false)

	cfg.Set(viper.ConfigKeyAuthDebug, true)
	cfg.Set(viper.ConfigKeyAuthCookieDomain, "localhost")
	cfg.Set(viper.ConfigKeyAuthCookieSigningKey, debugCookieSecret)
	cfg.Set(viper.ConfigKeyAuthCookieDomain, defaultCookieDomain)
	cfg.Set(viper.ConfigKeyAuthCookieLifetime, auth.DefaultCookieLifetime)
	cfg.Set(viper.ConfigKeyAuthSecureCookiesOnly, false)
	cfg.Set(viper.ConfigKeyAuthEnableUserSignup, true)

	cfg.Set(viper.ConfigKeyMetricsProvider, "prometheus")
	cfg.Set(viper.ConfigKeyMetricsTracer, "jaeger")
	cfg.Set(viper.ConfigKeyDatabaseMetricsCollectionInterval, time.Second)
	cfg.Set(viper.ConfigKeyMetricsRuntimeCollectionInterval, time.Second)

	cfg.Set(viper.ConfigKeyDatabaseDebug, true)
	cfg.Set(viper.ConfigKeyDatabaseRunMigrations, true)
	cfg.Set(viper.ConfigKeyDatabaseProvider, postgres)
	cfg.Set(viper.ConfigKeyDatabaseConnectionDetails, devPostgresDBConnDetails)

	cfg.Set(viper.ConfigKeyDatabaseCreateTestUserUsername, "username")
	cfg.Set(viper.ConfigKeyDatabaseCreateTestUserPassword, defaultPassword)
	cfg.Set(viper.ConfigKeyDatabaseCreateTestUserHashedPassword, mustHashPass(defaultPassword))
	cfg.Set(viper.ConfigKeyDatabaseCreateTestUserIsAdmin, true)

	cfg.Set(viper.ConfigKeyItemsSearchIndexPath, "/search_indices/items.bleve")

	if writeErr := cfg.WriteConfigAs(filePath); writeErr != nil {
		return fmt.Errorf("error writing developmentEnv config: %w", writeErr)
	}

	return nil
}

func frontendTestsConfig(filePath string) error {
	cfg := viper.BuildViperConfig()

	cfg.Set(viper.ConfigKeyMetaRunMode, developmentEnv)
	cfg.Set(viper.ConfigKeyMetaStartupDeadline, time.Minute)

	cfg.Set(viper.ConfigKeyServerHTTPPort, defaultPort)
	cfg.Set(viper.ConfigKeyServerDebug, true)

	cfg.Set(viper.ConfigKeyFrontendDebug, true)
	cfg.Set(viper.ConfigKeyFrontendStaticFilesDir, defaultFrontendFilepath)
	cfg.Set(viper.ConfigKeyFrontendCacheStatics, false)

	cfg.Set(viper.ConfigKeyAuthDebug, true)
	cfg.Set(viper.ConfigKeyAuthCookieDomain, "localhost")
	cfg.Set(viper.ConfigKeyAuthCookieSigningKey, debugCookieSecret)
	cfg.Set(viper.ConfigKeyAuthCookieDomain, defaultCookieDomain)
	cfg.Set(viper.ConfigKeyAuthCookieLifetime, auth.DefaultCookieLifetime)
	cfg.Set(viper.ConfigKeyAuthSecureCookiesOnly, false)
	cfg.Set(viper.ConfigKeyAuthEnableUserSignup, true)

	cfg.Set(viper.ConfigKeyMetricsProvider, "prometheus")
	cfg.Set(viper.ConfigKeyMetricsTracer, "jaeger")
	cfg.Set(viper.ConfigKeyDatabaseMetricsCollectionInterval, time.Second)
	cfg.Set(viper.ConfigKeyMetricsRuntimeCollectionInterval, time.Second)

	cfg.Set(viper.ConfigKeyDatabaseDebug, true)
	cfg.Set(viper.ConfigKeyDatabaseProvider, postgres)
	cfg.Set(viper.ConfigKeyDatabaseRunMigrations, true)
	cfg.Set(viper.ConfigKeyDatabaseConnectionDetails, devPostgresDBConnDetails)

	cfg.Set(viper.ConfigKeyItemsSearchIndexPath, defaultItemsSearchIndexPath)

	if writeErr := cfg.WriteConfigAs(filePath); writeErr != nil {
		return fmt.Errorf("error writing developmentEnv config: %w", writeErr)
	}

	return nil
}

func coverageConfig(filePath string) error {
	cfg := viper.BuildViperConfig()

	cfg.Set(viper.ConfigKeyMetaRunMode, testingEnv)
	cfg.Set(viper.ConfigKeyMetaDebug, true)

	cfg.Set(viper.ConfigKeyServerHTTPPort, defaultPort)
	cfg.Set(viper.ConfigKeyServerDebug, true)

	cfg.Set(viper.ConfigKeyFrontendDebug, true)
	cfg.Set(viper.ConfigKeyFrontendStaticFilesDir, defaultFrontendFilepath)
	cfg.Set(viper.ConfigKeyFrontendCacheStatics, false)

	cfg.Set(viper.ConfigKeyAuthDebug, false)
	cfg.Set(viper.ConfigKeyAuthCookieSigningKey, debugCookieSecret)
	cfg.Set(viper.ConfigKeyAuthCookieDomain, defaultCookieDomain)

	cfg.Set(viper.ConfigKeyDatabaseDebug, false)
	cfg.Set(viper.ConfigKeyDatabaseProvider, postgres)
	cfg.Set(viper.ConfigKeyDatabaseRunMigrations, true)
	cfg.Set(viper.ConfigKeyDatabaseConnectionDetails, devPostgresDBConnDetails)

	cfg.Set(viper.ConfigKeyDatabaseCreateTestUserUsername, "coverageUser")
	cfg.Set(viper.ConfigKeyDatabaseCreateTestUserPassword, defaultPassword)
	cfg.Set(viper.ConfigKeyDatabaseCreateTestUserHashedPassword, mustHashPass(defaultPassword))
	cfg.Set(viper.ConfigKeyDatabaseCreateTestUserIsAdmin, false)

	cfg.Set(viper.ConfigKeyItemsSearchIndexPath, defaultItemsSearchIndexPath)

	if writeErr := cfg.WriteConfigAs(filePath); writeErr != nil {
		return fmt.Errorf("error writing coverage config: %w", writeErr)
	}

	return nil
}

func buildIntegrationTestForDBImplementation(dbVendor, dbDetails string) configFunc {
	return func(filePath string) error {
		cfg := viper.BuildViperConfig()

		cfg.Set(viper.ConfigKeyMetaRunMode, testingEnv)
		cfg.Set(viper.ConfigKeyMetaDebug, false)

		sd := time.Minute
		if dbVendor == mariadb {
			sd = 5 * time.Minute
		}

		cfg.Set(viper.ConfigKeyMetaStartupDeadline, sd)

		cfg.Set(viper.ConfigKeyServerHTTPPort, defaultPort)
		cfg.Set(viper.ConfigKeyServerDebug, false)

		cfg.Set(viper.ConfigKeyFrontendStaticFilesDir, defaultFrontendFilepath)
		cfg.Set(viper.ConfigKeyAuthCookieSigningKey, debugCookieSecret)
		cfg.Set(viper.ConfigKeyAuthCookieDomain, defaultCookieDomain)

		cfg.Set(viper.ConfigKeyMetricsProvider, "prometheus")
		cfg.Set(viper.ConfigKeyMetricsTracer, "jaeger")

		cfg.Set(viper.ConfigKeyDatabaseDebug, false)
		cfg.Set(viper.ConfigKeyDatabaseProvider, dbVendor)
		cfg.Set(viper.ConfigKeyDatabaseRunMigrations, true)
		cfg.Set(viper.ConfigKeyDatabaseConnectionDetails, dbDetails)

		pw := "integration-tests-are-cool"

		cfg.Set(viper.ConfigKeyDatabaseCreateTestUserUsername, "exampleUser")
		cfg.Set(viper.ConfigKeyDatabaseCreateTestUserPassword, pw)
		cfg.Set(viper.ConfigKeyDatabaseCreateTestUserHashedPassword, mustHashPass(pw))
		cfg.Set(viper.ConfigKeyDatabaseCreateTestUserIsAdmin, true)

		cfg.Set(viper.ConfigKeyItemsSearchIndexPath, defaultItemsSearchIndexPath)

		if writeErr := cfg.WriteConfigAs(filePath); writeErr != nil {
			return fmt.Errorf("error writing integration test config for %s: %w", dbVendor, writeErr)
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
