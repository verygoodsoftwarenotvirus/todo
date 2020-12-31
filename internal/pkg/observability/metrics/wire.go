package metrics

import (
	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
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
