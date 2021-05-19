// +build wireinject

package server

import (
	"context"
	"database/sql"

	"github.com/google/wire"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/passwords"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/chi"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search/bleve"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/server"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/accounts"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/admin"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/apiclients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"
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
) (*server.HTTPServer, error) {
	wire.Build(
		bleve.Providers,
		config.Providers,
		database.Providers,
		encoding.Providers,
		server.Providers,
		metrics.Providers,
		dbconfig.Providers,
		images.Providers,
		uploads.Providers,
		observability.Providers,
		storage.Providers,
		chi.Providers,
		authentication.Providers,
		users.Providers,
		accounts.Providers,
		apiclients.Providers,
		webhooks.Providers,
		audit.Providers,
		admin.Providers,
		frontend.Providers,
		items.Providers,
	)
	return nil, nil
}
