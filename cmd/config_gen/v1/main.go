package main

import (
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
)

type configFunc func(filepath string) error

var (
	files = map[string]configFunc{
		"config_files/coverage.toml":                   coverageConfig,
		"config_files/development.toml":                developmentConfig,
		"config_files/integration-tests-postgres.toml": buildIntegrationTestForDBImplementation("postgres", postgresDBConnDetails),
		"config_files/integration-tests-sqlite.toml":   buildIntegrationTestForDBImplementation("sqlite", "/tmp/db"),
		"config_files/integration-tests-mariadb.toml":  buildIntegrationTestForDBImplementation("mariadb", "dbuser:hunter2@tcp(database:3306)/todo"),
		"config_files/production.toml":                 productionConfig,
	}
)

func developmentConfig(filepath string) error {
	cfg := config.BuildConfig()

	cfg.Set(metaStartupDeadline, time.Minute)
	cfg.Set(serverHTTPPort, defaultPort)
	cfg.Set(serverDebug, true)

	cfg.Set(frontendDebug, true)
	cfg.Set(frontendStaticFilesDir, defaultFrontendFilepath)
	cfg.Set(frontendCacheStatics, false)

	cfg.Set(authDebug, true)
	cfg.Set(authCookieDomain, "")
	cfg.Set(authCookieSecret, debugCookieSecret)
	cfg.Set(authCookieLifetime, oneDay)
	cfg.Set(authSecureCookiesOnly, false)
	cfg.Set(authEnableUserSignup, true)

	cfg.Set(metricsProvider, "prometheus")
	cfg.Set(metricsTracer, "jaeger")
	cfg.Set(metricsDBCollectionInterval, time.Second)
	cfg.Set(metricsRuntimeCollectionInterval, time.Second)

	cfg.Set(dbDebug, true)
	cfg.Set(dbProvider, "postgres")
	cfg.Set(dbDeets, postgresDBConnDetails)

	return cfg.WriteConfigAs(filepath)
}

func coverageConfig(filepath string) error {
	cfg := config.BuildConfig()

	cfg.Set(serverHTTPPort, defaultPort)
	cfg.Set(serverDebug, true)

	cfg.Set(frontendDebug, true)
	cfg.Set(frontendStaticFilesDir, defaultFrontendFilepath)
	cfg.Set(frontendCacheStatics, false)

	cfg.Set(authDebug, false)
	cfg.Set(authCookieSecret, debugCookieSecret)

	cfg.Set(dbDebug, false)
	cfg.Set(dbProvider, "postgres")
	cfg.Set(dbDeets, postgresDBConnDetails)

	return cfg.WriteConfigAs(filepath)
}

func productionConfig(filepath string) error {
	cfg := config.BuildConfig()

	cfg.Set(metaDebug, false)
	cfg.Set(metaStartupDeadline, time.Minute)

	cfg.Set(serverHTTPPort, defaultPort)
	cfg.Set(serverDebug, false)

	cfg.Set(frontendDebug, false)
	cfg.Set(frontendStaticFilesDir, defaultFrontendFilepath)
	cfg.Set(frontendCacheStatics, false)

	cfg.Set(authDebug, false)
	cfg.Set(authCookieDomain, "")
	cfg.Set(authCookieSecret, debugCookieSecret)
	cfg.Set(authCookieLifetime, oneDay)
	cfg.Set(authSecureCookiesOnly, false)
	cfg.Set(authEnableUserSignup, true)

	cfg.Set(metricsProvider, "prometheus")
	cfg.Set(metricsTracer, "jaeger")
	cfg.Set(metricsDBCollectionInterval, time.Second)
	cfg.Set(metricsRuntimeCollectionInterval, time.Second)

	cfg.Set(dbDebug, false)
	cfg.Set(dbProvider, "postgres")
	cfg.Set(dbDeets, postgresDBConnDetails)

	return cfg.WriteConfigAs(filepath)
}

func buildIntegrationTestForDBImplementation(dbprov, dbDeet string) configFunc {
	return func(filepath string) error {
		cfg := config.BuildConfig()

		cfg.Set(metaDebug, false)
		cfg.Set(metaStartupDeadline, time.Minute)

		cfg.Set(serverHTTPPort, defaultPort)
		cfg.Set(serverDebug, true)

		cfg.Set(frontendStaticFilesDir, defaultFrontendFilepath)
		cfg.Set(authCookieSecret, debugCookieSecret)

		cfg.Set(metricsProvider, "prometheus")
		cfg.Set(metricsTracer, "jaeger")

		cfg.Set(dbDebug, false)
		cfg.Set(dbProvider, dbprov)
		cfg.Set(dbDeets, dbDeet)

		return cfg.WriteConfigAs(filepath)
	}
}

func main() {
	for filepath, fun := range files {
		if err := fun(filepath); err != nil {
			log.Fatal(err)
		}
	}
}
