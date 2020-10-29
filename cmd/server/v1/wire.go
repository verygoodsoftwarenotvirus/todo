// +build wireinject

package main

import (
	"context"
	"database/sql"
	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/search/bleve"
	server "gitlab.com/verygoodsoftwarenotvirus/todo/server/v1"
	httpserver "gitlab.com/verygoodsoftwarenotvirus/todo/server/v1/http"
	auditservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/audit"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/auth"
	frontendservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/frontend"
	itemsservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	oauth2clientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	usersservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/webhooks"

	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
)

// ProvideReporter is an obligatory function that hopefully wire will eliminate for me one day.
func ProvideReporter(n *newsman.Newsman) newsman.Reporter {
	return n
}

// BuildServer builds a server.
func BuildServer(
	ctx context.Context,
	cfg *config.ServerConfig,
	logger logging.Logger,
	dbm database.DataManager,
	db *sql.DB,
	authenticator auth.Authenticator,
) (*server.Server, error) {
	wire.Build(
		config.Providers,
		database.Providers,
		// server things,
		bleve.Providers,
		server.Providers,
		encoding.Providers,
		httpserver.Providers,
		// metrics,
		metrics.Providers,
		// external libs,
		newsman.NewNewsman,
		ProvideReporter,
		// services,
		auditservice.Providers,
		authservice.Providers,
		usersservice.Providers,
		itemsservice.Providers,
		frontendservice.Providers,
		webhooksservice.Providers,
		oauth2clientsservice.Providers,
	)
	return nil, nil
}
