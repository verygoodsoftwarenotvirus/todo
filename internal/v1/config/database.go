package config

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	dbclient "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/client"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/queriers/mariadb"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/queriers/postgres"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/queriers/sqlite"

	"contrib.go.opencensus.io/integrations/ocsql"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
)

const (
	postgresProviderKey = "postgres"
	mariaDBProviderKey  = "mariadb"
	sqliteProviderKey   = "sqlite"
)

// ProvideDatabase provides a database implementation dependent on the configuration
func (cfg *ServerConfig) ProvideDatabase(ctx context.Context, logger logging.Logger) (database.Database, error) {
	var (
		debug             = cfg.Database.Debug || cfg.Meta.Debug
		connectionDetails = cfg.Database.ConnectionDetails
	)

	switch cfg.Database.Provider {
	case postgresProviderKey:
		rawDB, err := postgres.ProvidePostgresDB(logger, connectionDetails)
		if err != nil {
			return nil, fmt.Errorf("establish postgres database connection: %w", err)
		}
		ocsql.RegisterAllViews()
		ocsql.RecordStats(rawDB, cfg.Metrics.DBMetricsCollectionInterval)

		pg := postgres.ProvidePostgres(debug, rawDB, logger)

		return dbclient.ProvideDatabaseClient(ctx, rawDB, pg, debug, logger)
	case mariaDBProviderKey:
		rawDB, err := mariadb.ProvideMariaDBConnection(logger, connectionDetails)
		if err != nil {
			return nil, fmt.Errorf("establish mariadb database connection: %w", err)
		}
		ocsql.RegisterAllViews()
		ocsql.RecordStats(rawDB, cfg.Metrics.DBMetricsCollectionInterval)

		pg := mariadb.ProvideMariaDB(debug, rawDB, logger)

		return dbclient.ProvideDatabaseClient(ctx, rawDB, pg, debug, logger)
	case sqliteProviderKey:
		rawDB, err := sqlite.ProvideSqliteDB(logger, connectionDetails)
		if err != nil {
			return nil, fmt.Errorf("establish sqlite database connection: %w", err)
		}
		ocsql.RegisterAllViews()
		ocsql.RecordStats(rawDB, cfg.Metrics.DBMetricsCollectionInterval)

		pg := sqlite.ProvideSqlite(debug, rawDB, logger)

		return dbclient.ProvideDatabaseClient(ctx, rawDB, pg, debug, logger)
	default:
		return nil, errors.New("invalid database type selected")
	}
}
