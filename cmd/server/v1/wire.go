//+build wireinject

package main

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/config/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/auth/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/encoding/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/metrics/v1/prometheus"
	"gitlab.com/verygoodsoftwarenotvirus/todo/server/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"

	"github.com/google/wire"
)

// BuildServer builds a server
func BuildServer(
	cfg *config.ServerConfig,
	logger logging.Logger,
	database database.Database,
) (*server.Server, error) {

	wire.Build(
		config.Providers,
		auth.Providers,

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
