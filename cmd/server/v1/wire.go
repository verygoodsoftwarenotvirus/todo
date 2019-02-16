//+build wireinject

package main

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/postgres"
	// "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/sqlite"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/metrics/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/metrics/v1/prometheus"
	"gitlab.com/verygoodsoftwarenotvirus/todo/server/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"

	"github.com/google/wire"
)

// BuildServer builds a server
func BuildServer(
	connectionDetails database.ConnectionDetails,
	CertPair server.CertPair,
	CookieName users.CookieName,
	metricsNamespace metrics.Namespace,
	CookieSecret []byte,
	Debug bool,
) (*server.Server, error) {

	wire.Build(
		auth.Providers,

		// Databases
		// sqlite.Providers,
		postgres.Providers,

		// Loggers
		zerolog.Providers,

		// Server things
		server.Providers,

		// metrics
		prometheus.Providers,

		// Services
		users.Providers,
		items.Providers,
		oauth2clients.Providers,
	)
	return nil, nil
}
