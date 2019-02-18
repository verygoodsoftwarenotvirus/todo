package database

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/queriers/postgres"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
)

// ProvidePostgresDatabase provides a postgres database
func ProvidePostgresDatabase(
	debug bool,
	logger logging.Logger,
	connectionDetails string,
) (Database, error) {
	return postgres.ProvidePostgres(debug, logger, postgres.ConnectionDetails(connectionDetails))
}
