package config

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	dbclient "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/client"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/queriers/postgres"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/queriers/sqlite"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1"

	"contrib.go.opencensus.io/integrations/ocsql"
	"github.com/pkg/errors"
)

// ProvideDatabase provides a database implementation dependent on the configuration
func (cfg *ServerConfig) ProvideDatabase(ctx context.Context, logger logging.Logger) (database.Database, error) {
	var (
		debug             = cfg.Database.Debug || cfg.Meta.Debug
		connectionDetails = cfg.Database.ConnectionDetails
	)

	switch cfg.Database.Provider {
	case "postgres":
		rawDB, err := postgres.ProvidePostgresDB(logger, connectionDetails)
		if err != nil {
			return nil, errors.Wrap(err, "establish postgres database connection")
		}
		ocsql.RegisterAllViews()
		ocsql.RecordStats(rawDB, cfg.Metrics.DBMetricsCollectionInterval)

		pg := postgres.ProvidePostgres(debug, rawDB, logger)

		return dbclient.ProvideDatabaseClient(ctx, rawDB, pg, debug, logger)
	case "sqlite":
		rawDB, err := sqlite.ProvideSqliteDB(logger, connectionDetails)
		if err != nil {
			return nil, errors.Wrap(err, "establish postgres database connection")
		}
		ocsql.RegisterAllViews()
		ocsql.RecordStats(rawDB, cfg.Metrics.DBMetricsCollectionInterval)

		pg := sqlite.ProvideSqlite(debug, rawDB, logger)

		return dbclient.ProvideDatabaseClient(ctx, rawDB, pg, debug, logger)
	default:
		return nil, errors.New("invalid database type selected")
	}
}
