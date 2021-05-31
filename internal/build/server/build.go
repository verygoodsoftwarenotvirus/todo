// +build wireinject

package server

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/capitalism"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/capitalism/stripe"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/chi"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search/bleve"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/server"
	accountsservice "gitlab.com/verygoodsoftwarenotvirus/naff/example_output/internal/services/accounts"
	adminservice "gitlab.com/verygoodsoftwarenotvirus/naff/example_output/internal/services/admin"
	apiclientsservice "gitlab.com/verygoodsoftwarenotvirus/naff/example_output/internal/services/apiclients"
	auditservice "gitlab.com/verygoodsoftwarenotvirus/naff/example_output/internal/services/audit"
	authservice "gitlab.com/verygoodsoftwarenotvirus/naff/example_output/internal/services/authentication"
	frontendservice "gitlab.com/verygoodsoftwarenotvirus/naff/example_output/internal/services/frontend"
	itemsservice "gitlab.com/verygoodsoftwarenotvirus/naff/example_output/internal/services/items"
	usersservice "gitlab.com/verygoodsoftwarenotvirus/naff/example_output/internal/services/users"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/naff/example_output/internal/services/webhooks"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/storage"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads/images"

	"github.com/google/wire"
)

// Build builds a server.
func Build(
	ctx context.Context,
	cfg *config.InstanceConfig,
	logger logging.Logger,
) (*server.HTTPServer, error) {
	wire.Build(
		bleve.Providers,
		config.Providers,
		database.Providers,
		dbconfig.Providers,
		encoding.Providers,
		server.Providers,
		metrics.Providers,
		images.Providers,
		uploads.Providers,
		observability.Providers,
		storage.Providers,
		capitalism.Providers,
		stripe.Providers,
		chi.Providers,
		authentication.Providers,
		authservice.Providers,
		usersservice.Providers,
		accountsservice.Providers,
		apiclientsservice.Providers,
		webhooksservice.Providers,
		auditservice.Providers,
		adminservice.Providers,
		frontendservice.Providers,
		itemsservice.Providers,
	)

	return nil, nil
}
