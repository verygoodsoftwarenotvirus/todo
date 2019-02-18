//+build wireinject

package main

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/client"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/encoding/v1"
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
	connectionDetails string,
	CookieName users.CookieName,
	metricsNamespace metrics.Namespace,
	CookieSecret []byte,
	Debug bool,
) (*server.Server, error) {

	wire.Build(
		auth.Providers,

		// Loggers
		zerolog.Providers,

		// Database things
		dbclient.Providers,

		// Server things
		server.Providers,
		encoding.Providers,

		// metrics
		prometheus.Providers,

		// Services
		users.Providers,
		items.Providers,
		oauth2clients.Providers,
	)
	return nil, nil
}
