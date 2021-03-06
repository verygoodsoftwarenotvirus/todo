//go:build wireinject
// +build wireinject

package server

import (
	"context"

	"github.com/google/wire"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	msgconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/messagequeue/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/chi"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search/elasticsearch"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/server"
	accountsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/accounts"
	adminservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/admin"
	apiclientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/apiclients"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"
	frontendservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/frontend"
	itemsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/items"
	usersservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/users"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/webhooks"
	websocketsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/websockets"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/storage"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads/images"
)

// Build builds a server.
func Build(
	ctx context.Context,
	logger logging.Logger,
	cfg *config.InstanceConfig,
) (*server.HTTPServer, error) {
	wire.Build(
		elasticsearch.Providers,
		config.Providers,
		database.Providers,
		dbconfig.Providers,
		encoding.Providers,
		msgconfig.Providers,
		server.Providers,
		metrics.Providers,
		images.Providers,
		uploads.Providers,
		observability.Providers,
		storage.Providers,
		chi.Providers,
		authentication.Providers,
		authservice.Providers,
		usersservice.Providers,
		accountsservice.Providers,
		apiclientsservice.Providers,
		webhooksservice.Providers,
		websocketsservice.Providers,
		adminservice.Providers,
		frontendservice.Providers,
		itemsservice.Providers,
	)

	return nil, nil
}
