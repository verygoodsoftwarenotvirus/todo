// +build wireinject

package main

import (
	"context"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	auth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding"
	metrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics"
	server "gitlab.com/verygoodsoftwarenotvirus/todo/server/v1"
	httpserver "gitlab.com/verygoodsoftwarenotvirus/todo/server/v1/http"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/auth"
	frontendservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/frontend"
	itemsservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	oauth2clientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	usersservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/webhooks"

	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
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
		auth.Providers,
		// server things,
		server.Providers,
		encoding.Providers,
		httpserver.Providers,
		// metrics,
		metrics.Providers,
		// external libs,
		newsman.NewNewsman,
		ProvideReporter,
		// services,
		authservice.Providers,
		usersservice.Providers,
		itemsservice.Providers,
		frontendservice.Providers,
		webhooksservice.Providers,
		oauth2clientsservice.Providers,
	)
	return nil, nil
}
