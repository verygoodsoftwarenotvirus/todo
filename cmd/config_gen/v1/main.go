package main

import (
	"fmt"
	"log"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"
)

const (
	defaultPort                      = 8888
	oneDay                           = 24 * time.Hour
	debugCookieSecret                = "HEREISA32CHARSECRETWHICHISMADEUP"
	defaultFrontendFilepath          = "/frontend"
	postgresDBConnDetails            = "postgres://dbuser:hunter2@database:5432/todo?sslmode=disable"
	metaDebug                        = "meta.debug"
	metaRunMode                      = "meta.run_mode"
	metaStartupDeadline              = "meta.startup_deadline"
	serverHTTPPort                   = "server.http_port"
	serverDebug                      = "server.debug"
	frontendDebug                    = "frontend.debug"
	frontendStaticFilesDir           = "frontend.static_files_directory"
	frontendCacheStatics             = "frontend.cache_static_files"
	authDebug                        = "auth.debug"
	authCookieDomain                 = "auth.cookie_domain"
	authCookieSecret                 = "auth.cookie_secret"
	authCookieLifetime               = "auth.cookie_lifetime"
	authSecureCookiesOnly            = "auth.secure_cookies_only"
	authEnableUserSignup             = "auth.enable_user_signup"
	metricsProvider                  = "metrics.metrics_provider"
	metricsTracer                    = "metrics.tracing_provider"
	metricsDBCollectionInterval      = "metrics.database_metrics_collection_interval"
	metricsRuntimeCollectionInterval = "metrics.runtime_metrics_collection_interval"
	dbDebug                          = "database.debug"
	dbProvider                       = "database.provider"
	dbDeets                          = "database.connection_details"
	dbCreateTestUser                 = "database.create_test_user"
	itemsSearchIndexPath             = "search.items_index_path"

	// run modes
	developmentEnv = "development"
	testingEnv     = "testing"

	// database providers
	postgres = "postgres"
	sqlite   = "sqlite"
	mariadb  = "mariadb"

	// search index paths
	defaultItemsSearchIndexPath = "items.bleve"
)

type configFunc func(filePath string) error

var (
	files = map[string]configFunc{
		"environments/local/config.toml":                                    localDevelopmentCOnfig,
		"environments/testing/config_files/frontend-tests.toml":             frontendTestsConfig,
		"environments/testing/config_files/coverage.toml":                   coverageConfig,
		"environments/testing/config_files/integration-tests-postgres.toml": buildIntegrationTestForDBImplementation(postgres, postgresDBConnDetails),
		"environments/testing/config_files/integration-tests-sqlite.toml":   buildIntegrationTestForDBImplementation(sqlite, "/tmp/db"),
		"environments/testing/config_files/integration-tests-mariadb.toml":  buildIntegrationTestForDBImplementation(mariadb, "dbuser:hunter2@tcp(database:3306)/todo"),
	}
)

func localDevelopmentCOnfig(filePath string) error {
	cfg := config.BuildConfig()

	cfg.Set(metaRunMode, developmentEnv)
	cfg.Set(metaDebug, true)
	cfg.Set(metaStartupDeadline, time.Minute)

	cfg.Set(serverHTTPPort, defaultPort)
	cfg.Set(serverDebug, true)

	cfg.Set(frontendDebug, true)
	cfg.Set(frontendStaticFilesDir, defaultFrontendFilepath)
	cfg.Set(frontendCacheStatics, false)

	cfg.Set(authDebug, true)
	cfg.Set(authCookieDomain, "localhost")
	cfg.Set(authCookieSecret, debugCookieSecret)
	cfg.Set(authCookieLifetime, oneDay)
	cfg.Set(authSecureCookiesOnly, false)
	cfg.Set(authEnableUserSignup, true)

	cfg.Set(metricsProvider, "prometheus")
	cfg.Set(metricsTracer, "jaeger")
	cfg.Set(metricsDBCollectionInterval, time.Second)
	cfg.Set(metricsRuntimeCollectionInterval, time.Second)

	cfg.Set(dbDebug, true)
	cfg.Set(dbProvider, postgres)
	cfg.Set(dbDeets, postgresDBConnDetails)
	cfg.Set(dbCreateTestUser, true)

	cfg.Set(itemsSearchIndexPath, "/search_indices/items.bleve")

	if writeErr := cfg.WriteConfigAs(filePath); writeErr != nil {
		return fmt.Errorf("error writing developmentEnv config: %w", writeErr)
	}

	return nil
}

func frontendTestsConfig(filePath string) error {
	cfg := config.BuildConfig()

	cfg.Set(metaRunMode, developmentEnv)
	cfg.Set(metaStartupDeadline, time.Minute)

	cfg.Set(serverHTTPPort, defaultPort)
	cfg.Set(serverDebug, true)

	cfg.Set(frontendDebug, true)
	cfg.Set(frontendStaticFilesDir, defaultFrontendFilepath)
	cfg.Set(frontendCacheStatics, false)

	cfg.Set(authDebug, true)
	cfg.Set(authCookieDomain, "localhost")
	cfg.Set(authCookieSecret, debugCookieSecret)
	cfg.Set(authCookieLifetime, oneDay)
	cfg.Set(authSecureCookiesOnly, false)
	cfg.Set(authEnableUserSignup, true)

	cfg.Set(metricsProvider, "prometheus")
	cfg.Set(metricsTracer, "jaeger")
	cfg.Set(metricsDBCollectionInterval, time.Second)
	cfg.Set(metricsRuntimeCollectionInterval, time.Second)

	cfg.Set(dbDebug, true)
	cfg.Set(dbProvider, postgres)
	cfg.Set(dbDeets, postgresDBConnDetails)
	cfg.Set(dbCreateTestUser, false)

	cfg.Set(itemsSearchIndexPath, defaultItemsSearchIndexPath)

	if writeErr := cfg.WriteConfigAs(filePath); writeErr != nil {
		return fmt.Errorf("error writing developmentEnv config: %w", writeErr)
	}

	return nil
}

func coverageConfig(filePath string) error {
	cfg := config.BuildConfig()

	cfg.Set(metaRunMode, testingEnv)
	cfg.Set(metaDebug, true)

	cfg.Set(serverHTTPPort, defaultPort)
	cfg.Set(serverDebug, true)

	cfg.Set(frontendDebug, true)
	cfg.Set(frontendStaticFilesDir, defaultFrontendFilepath)
	cfg.Set(frontendCacheStatics, false)

	cfg.Set(authDebug, false)
	cfg.Set(authCookieSecret, debugCookieSecret)

	cfg.Set(dbDebug, false)
	cfg.Set(dbProvider, postgres)
	cfg.Set(dbDeets, postgresDBConnDetails)
	cfg.Set(dbCreateTestUser, false)

	cfg.Set(itemsSearchIndexPath, defaultItemsSearchIndexPath)

	if writeErr := cfg.WriteConfigAs(filePath); writeErr != nil {
		return fmt.Errorf("error writing coverage config: %w", writeErr)
	}

	return nil
}

func buildIntegrationTestForDBImplementation(dbVendor, dbDetails string) configFunc {
	return func(filePath string) error {
		cfg := config.BuildConfig()

		cfg.Set(metaRunMode, testingEnv)
		cfg.Set(metaDebug, false)

		sd := time.Minute
		if dbVendor == mariadb {
			sd = 5 * time.Minute
		}
		cfg.Set(metaStartupDeadline, sd)

		cfg.Set(serverHTTPPort, defaultPort)
		cfg.Set(serverDebug, true)

		cfg.Set(frontendStaticFilesDir, defaultFrontendFilepath)
		cfg.Set(authCookieSecret, debugCookieSecret)

		cfg.Set(metricsProvider, "prometheus")
		cfg.Set(metricsTracer, "jaeger")

		cfg.Set(dbDebug, false)
		cfg.Set(dbProvider, dbVendor)
		cfg.Set(dbDeets, dbDetails)
		cfg.Set(dbCreateTestUser, false)

		cfg.Set(itemsSearchIndexPath, defaultItemsSearchIndexPath)

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
