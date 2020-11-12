// +build wireinject

package main

import (
	"context"
	"database/sql"
	database2 "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database"
	server2 "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/server"
	httpserver "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/server/http"
	auditservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/audit"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	frontendservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/frontend"
	itemsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/items"
	oauth2clientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/oauth2clients"
	usersservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/users"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/webhooks"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search/bleve"

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
