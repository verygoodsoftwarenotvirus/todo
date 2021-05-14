// +build wireinject

package server

import (
	"context"
	"database/sql"

	"github.com/google/wire"

	config "gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	observability "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/passwords"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/chi"
	bleve2 "gitlab.com/verygoodsoftwarenotvirus/todo/internal/search/bleve"
	server2 "gitlab.com/verygoodsoftwarenotvirus/todo/internal/server"
	httpserver "gitlab.com/verygoodsoftwarenotvirus/todo/internal/server/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/accounts"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/accountsubscriptionplans"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/admin"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/apiclients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/frontend"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/users"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/webhooks"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads/images"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads/storage"
)

// Build builds a server.
func Build(
	ctx context.Context,
	cfg *config.ServerConfig,
	logger logging.Logger,
	dbm database.DataManager,
	db *sql.DB,
	authenticator passwords.Authenticator,
) (*server2.Server, error) {
	wire.Build(
		server2.Providers,
		bleve2.Providers,
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
		auth.Providers,
		users.Providers,
		accounts.Providers,
		accountsubscriptionplans.Providers,
		apiclients.Providers,
		webhooks.Providers,
		audit.Providers,
		admin.Providers,
		frontend.Providers,
		items.Providers,
	)
	return nil, nil
}
