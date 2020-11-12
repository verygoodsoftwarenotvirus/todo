// +build wireinject

package main

import (
	"context"
	"database/sql"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
	database2 "gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search/bleve"
	server2 "gitlab.com/verygoodsoftwarenotvirus/todo/internal/server"
	httpserver "gitlab.com/verygoodsoftwarenotvirus/todo/internal/server/http"
	auditservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/audit"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/auth"
	frontendservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/frontend"
	itemsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/items"
	oauth2clientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/oauth2clients"
	usersservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/users"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/webhooks"

	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

// BuildServer builds a server.
func BuildServer(
	ctx context.Context,
	cfg *config.ServerConfig,
	logger logging.Logger,
	dbm database2.DataManager,
	db *sql.DB,
	authenticator auth.Authenticator,
) (*server2.Server, error) {
	wire.Build(
		config.Providers,
		database2.Providers,
		// server things,
		bleve.Providers,
		server2.Providers,
		encoding.Providers,
		httpserver.Providers,
		// metrics,
		metrics.Providers,
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
