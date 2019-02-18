package dbclient

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/queriers/postgres"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
)

// ProvidePostgresDatabase provides a postgres database
func ProvidePostgresDatabase(
	debug bool,
	logger logging.Logger,
	connectionDetails string,
	tracer Tracer,
) (database.Database, error) {
	db, err := postgres.ProvidePostgres(debug, logger, postgres.ConnectionDetails(connectionDetails))
	if err != nil {
		return nil, err
	}

	return ProvideDatabaseClient(
		db,
		debug,
		logger,
		tracer,
	)
}
