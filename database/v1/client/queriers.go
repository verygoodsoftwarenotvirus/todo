package dbclient

import (
	"database/sql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/queriers/postgres"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
)

// ProvidePostgresDatabase provides a postgres database
func ProvidePostgresDatabase(
	debug bool,
	db *sql.DB,
	logger logging.Logger,
	connectionDetails database.ConnectionDetails,
	tracer Tracer,
) (database.Database, error) {
	dbc, err := postgres.ProvidePostgres(debug, db, logger, connectionDetails)
	if err != nil {
		return nil, err
	}

	return ProvideDatabaseClient(
		dbc,
		debug,
		logger,
		tracer,
	)
}
