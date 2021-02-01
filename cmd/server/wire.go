// +build wireinject

package main

import (
	"context"
	"database/sql"

	server2 "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/server"
	httpserver "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/server/http"
	adminservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/admin"
	auditservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/audit"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	frontendservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/frontend"
	itemsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/items"
	oauth2clientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/oauth2clients"
	plansservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/plans"
	usersservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/users"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/webhooks"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	database2 "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search/bleve"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/images"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/storage"

	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/logging"
)

// BuildServer builds a server.
func BuildServer(
	ctx context.Context,
	cfg *config.ServerConfig,
	logger logging.Logger,
	dbm database2.DataManager,
	db *sql.DB,
	authenticator password.Authenticator,
) (*server2.Server, error) {
	wire.Build(
		// app
		server2.Providers,
		// packages,
		bleve.Providers,
		config.Providers,
		database2.Providers,
		encoding.Providers,
		httpserver.Providers,
		metrics.Providers,
		dbconfig.Providers,
		images.Providers,
		uploads.Providers,
		observability.Providers,
		storage.Providers,
		// services,
		adminservice.Providers,
		auditservice.Providers,
		authservice.Providers,
		plansservice.Providers,
		usersservice.Providers,
		itemsservice.Providers,
		frontendservice.Providers,
		webhooksservice.Providers,
		oauth2clientsservice.Providers,
	)
	return nil, nil
}
