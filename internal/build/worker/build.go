//go:build wireinject
// +build wireinject

package worker

import (
	"context"

	"github.com/google/wire"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/messagequeue/publishers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
)

// Build builds a worker.
func Build(
	ctx context.Context,
	cfg *config.InstanceConfig,
	logger logging.Logger,
) error {
	wire.Build(
		database.Providers,
		dbconfig.Providers,
		publishers.Providers,
		encoding.Providers,
		observability.Providers,
		authentication.Providers,
	)

	return nil
}
