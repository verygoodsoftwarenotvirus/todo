package metrics

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"

	"github.com/google/wire"
)

var (
	// Providers represents what this library offers to external users in the form of dependencies.
	Providers = wire.NewSet(
		ProvideUnitCounterProvider,
		ProvideMetricsInstrumentationHandlerForServer,
	)
)

// ProvideMetricsInstrumentationHandlerForServer provides a metrics.InstrumentationHandler from a config for our server.
func ProvideMetricsInstrumentationHandlerForServer(cfg *Config, logger logging.Logger) (InstrumentationHandler, error) {
	return cfg.ProvideInstrumentationHandler(logger)
}

// ProvideUnitCounterProvider provides a metrics.InstrumentationHandler from a config for our server.
func ProvideUnitCounterProvider(cfg *Config, logger logging.Logger) (UnitCounterProvider, error) {
	return cfg.ProvideUnitCounterProvider(logger)
}
