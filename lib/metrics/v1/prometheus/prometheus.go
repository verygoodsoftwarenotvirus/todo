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
		ProvideMetricsHandler,
	)
)

// Collector is our metric collector that exports to Prometheus
type Collector struct {
}

// ProvideInstrumentationHandler provides an instrumentation handler
func ProvideInstrumentationHandler(name metrics.Namespace, handler http.Handler) metrics.InstrumentationHandler {
	return prometheus.InstrumentHandlerFunc(string(name), handler.ServeHTTP)
}

// ProvideMetricsHandler provides a metrics handler
func ProvideMetricsHandler() metrics.Handler {
	return promhttp.Handler()
}
