package frontend

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"

	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

// Providers is our collection of what we provide to other services.
var Providers = wire.NewSet(
	ProvideService,
	ProvideMetricsInstrumentationHandlerForServer,
)

// ProvideMetricsInstrumentationHandlerForServer provides a metrics.InstrumentationHandler from a config for our server.
func ProvideMetricsInstrumentationHandlerForServer(cfg *observability.Config, logger logging.Logger) metrics.InstrumentationHandler {
	return cfg.ProvideInstrumentationHandler(logger)
}
