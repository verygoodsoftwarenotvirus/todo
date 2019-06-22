//+build wireinject

package main

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	libauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/auth/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1"
	metricsProvider "gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/server/v1"
	httpserver "gitlab.com/verygoodsoftwarenotvirus/todo/server/v1/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/frontend"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/webhooks"

	"gitlab.com/verygoodsoftwarenotvirus/newsman"

	"github.com/google/wire"
)

// ProvideReporter is an obligatory function that hopefully wire will eliminate for me one day
func ProvideReporter(n *newsman.Newsman) newsman.Reporter {
	return n
}

// BuildServer builds a server
func BuildServer(
	ctx context.Context,
	cfg *config.ServerConfig,
	logger logging.Logger,
	database database.Database,
) (*server.Server, error) {

	wire.Build(
		config.Providers,
		libauth.Providers,

		// Server things
		server.Providers,
		encoding.Providers,
		httpserver.Providers,

		// metrics
		metricsProvider.Providers,

		// external libs
		newsman.NewNewsman,
		ProvideReporter,

		auth.Providers,
		users.Providers,
		items.Providers,
		frontend.Providers,
		webhooks.Providers,
		oauth2clients.Providers,
	)
	return nil, nil
}
