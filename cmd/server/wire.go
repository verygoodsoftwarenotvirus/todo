// +build wireinject

package main

import (
	"context"
	"database/sql"

	server "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/server"
	httpserver "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/server/http"
	accountsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/accounts"
	plansservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/accountsubscriptionplans"
	adminservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/admin"
	apiclientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/apiclients"
	auditservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/audit"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	frontendservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/frontend"
	itemsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/items"
	usersservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/users"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/webhooks"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/passwords"
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
	authenticator passwords.Authenticator,
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
		chi.Providers,
		authservice.Providers,
		usersservice.Providers,
		accountsservice.Providers,
		plansservice.Providers,
		apiclientsservice.Providers,
		webhooksservice.Providers,
		auditservice.Providers,
		adminservice.Providers,
		frontendservice.Providers,
		itemsservice.Providers,
	)
	return nil, nil
}
