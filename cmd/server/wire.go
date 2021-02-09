// +build wireinject

package main

import (
	"context"
	"database/sql"

	server "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/server"
	httpserver "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/server/http"
	plansservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/accountsubscriptionplans"
	adminservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/admin"
	auditservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/audit"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	frontendservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/frontend"
	itemsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/items"
	oauth2clientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/oauth2clients"
	usersservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/users"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/webhooks"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing/chi"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search/bleve"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/images"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/storage"

	"github.com/google/wire"
)

// BuildServer builds a server.
func BuildServer(
	ctx context.Context,
	cfg *config.ServerConfig,
	logger logging.Logger,
	dbm database.DataManager,
	db *sql.DB,
	authenticator authentication.Authenticator,
) (*server.Server, error) {
	wire.Build(
		server.Providers,
		bleve.Providers,
		config.Providers,
		database.Providers,
		encoding.Providers,
		httpserver.Providers,
		metrics.Providers,
		dbconfig.Providers,
		images.Providers,
		uploads.Providers,
		observability.Providers,
		storage.Providers,
		adminservice.Providers,
		auditservice.Providers,
		authservice.Providers,
		plansservice.Providers,
		usersservice.Providers,
		itemsservice.Providers,
		chi.Providers,
		frontendservice.Providers,
		webhooksservice.Providers,
		oauth2clientsservice.Providers,
	)
	return nil, nil
}
