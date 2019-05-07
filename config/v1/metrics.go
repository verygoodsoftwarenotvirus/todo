package config

import (
	"fmt"
	"os"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1"

	"contrib.go.opencensus.io/exporter/jaeger"
	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/pkg/errors"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

type (
	metricsProvider string
	tracingProvider string
)

var (
	// Prometheus is one of our supported metrics providers
	Prometheus metricsProvider = "prometheus"
	// DefaultMetricsProvider indicates what the preferred metrics provider is
	DefaultMetricsProvider = Prometheus

	// Jaeger is one of our supported tracing providers
	Jaeger tracingProvider = "jaeger"
	// DefaultTracingProvider indicates what the preferred tracing provider is
	DefaultTracingProvider = Jaeger
)

// MetricsSettings contains settings about how we report our metrics
type MetricsSettings struct {
	Namespace                        metrics.Namespace `mapstructure:"metrics_namespace"`
	MetricsProvider                  metricsProvider   `mapstructure:"metrics_provider"`
	TracingProvider                  tracingProvider   `mapstructure:"tracing_provider"`
	DBMetricsCollectionInterval      time.Duration     `mapstructure:"database_metrics_collection_interval"`
	RuntimeMetricsCollectionInterval time.Duration     `mapstructure:"runtime_metrics_collection_interval"`
}

// ProvideInstrumentationHandler provides an instrumentation handler
func (cfg *ServerConfig) ProvideInstrumentationHandler(logger logging.Logger) (metrics.InstrumentationHandler, error) {
	if err := view.Register(ochttp.DefaultServerViews...); err != nil {
		return nil, errors.Wrap(err, "Failed to register server views for HTTP metrics")
	}

	metrics.RegisterDefaultViews()
	metrics.RecordRuntimeStats(cfg.Metrics.RuntimeMetricsCollectionInterval)

	logger.WithValue("metrics_provider", cfg.Metrics.MetricsProvider).Debug("setting metrics provider")

	switch cfg.Metrics.MetricsProvider {
	case Prometheus, DefaultMetricsProvider:
		p, err := prometheus.NewExporter(prometheus.Options{
			Namespace: string(cfg.Metrics.Namespace),
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create Prometheus exporter")
		}
		view.RegisterExporter(p)
	default:
		return nil, nil // fmt.Errorf("invalid metrics provider requested: %q", cfg.Metrics.MetricsProvider)
	}

	return nil, nil
}

// ProvideTracing provides an instrumentation handler
func (cfg *ServerConfig) ProvideTracing(logger logging.Logger) error {
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	switch cfg.Metrics.TracingProvider {
	case Jaeger, DefaultTracingProvider:
		logger.WithValue("tracing_provider", cfg.Metrics.TracingProvider).Info("setting tracing provider")
		je, err := jaeger.NewExporter(jaeger.Options{
			AgentEndpoint: fmt.Sprintf("%s:%s", os.Getenv("JAEGER_AGENT_HOST"), os.Getenv("JAEGER_AGENT_PORT")),
			Process: jaeger.Process{
				ServiceName: os.Getenv("JAEGER_SERVICE_NAME"),
			},
		})
		if err != nil {
			return errors.Wrap(err, "failed to create Jaeger exporter")
		}

		trace.RegisterExporter(je)
	default:
		return nil // fmt.Errorf("invalid tracing provider requested: %q", cfg.Metrics.TracingProvider)
	}

	return nil
}
