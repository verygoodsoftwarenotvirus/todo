package config

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	dbclient "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/client"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/queriers/postgres"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"

	"contrib.go.opencensus.io/integrations/ocsql"
	"github.com/pkg/errors"
)

// DatabaseSettings is a container struct for dealing with settings pertaining to
type DatabaseSettings struct {
	Type              string                     `mapstructure:"type"`
	Debug             bool                       `mapstructure:"debug"`
	ConnectionDetails database.ConnectionDetails `mapstructure:"connection_details"`
}

// ProvideDatabase provides a database implementation dependent on the configuration
func (cfg *ServerConfig) ProvideDatabase(logger logging.Logger) (database.Database, error) {
	var (
		debug             = cfg.Database.Debug || cfg.Meta.Debug
		connectionDetails = cfg.Database.ConnectionDetails
	)

	switch cfg.Database.Type {
	case "postgres":
		rawDB, err := postgres.ProvidePostgresDB(logger, connectionDetails)
		if err != nil {
			return nil, errors.Wrap(err, "establish postgres database connection")
		}
		ocsql.RegisterAllViews()
		ocsql.RecordStats(rawDB, cfg.Metrics.DBMetricsCollectionInterval)

		pg := postgres.ProvidePostgres(debug, rawDB, logger, connectionDetails)

		return dbclient.ProvideDatabaseClient(pg, debug, logger)
	default:
		return nil, errors.New("invalid database type selected")
	}
}
