package prometheus

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/metrics/v1"

	"github.com/google/wire"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Providers represents what this library offers to external users in the form of dependencies
	Providers = wire.NewSet(
		ProvideInstrumentationHandlerProvider,
		ProvideMetricsHandler,
	)
)

// ProvideInstrumentationHandlerProvider provides an instrumentation handler provider
func ProvideInstrumentationHandlerProvider() metrics.InstrumentationHandlerProvider {
	return func(handler http.Handler) metrics.InstrumentationHandler {
		return promhttp.InstrumentMetricHandler(prometheus.DefaultRegisterer, handler)
	}
}

// ProvideMetricsHandler provides a metrics handler
func ProvideMetricsHandler() metrics.Handler {
	return promhttp.Handler()
}
