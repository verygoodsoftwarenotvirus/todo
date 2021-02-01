package metrics

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"

	"github.com/google/wire"
)

var (
	// Providers represents what this library offers to external users in the form of dependencies.
	Providers = wire.NewSet(
		ProvideUnitCounter,
		ProvideUnitCounterProvider,
		ProvideMetricsInstrumentationHandlerForServer,
	)
)

// ProvideUnitCounterProvider provides UnitCounter providers.
func ProvideUnitCounterProvider() UnitCounterProvider {
	return ProvideUnitCounter
}

// ProvideMetricsInstrumentationHandlerForServer provides a metrics.InstrumentationHandler from a config for our server.
func ProvideMetricsInstrumentationHandlerForServer(cfg *Config, logger logging.Logger) InstrumentationHandler {
	return cfg.ProvideInstrumentationHandler(logger)
}
