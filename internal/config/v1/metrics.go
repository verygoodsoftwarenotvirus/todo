package config

import (
	"fmt"
	"os"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"

	"contrib.go.opencensus.io/exporter/jaeger"
	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/pkg/errors"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

const (
	// MetricsNamespace is the namespace under which we register metrics
	MetricsNamespace = "todo_server"
)

type (
	metricsProvider string
	tracingProvider string
)

var (
	// ErrInvalidMetricsProvider is a sentinel error value
	ErrInvalidMetricsProvider = errors.New("invalid metrics provider")
	// Prometheus is one of our supported metrics providers
	Prometheus metricsProvider = "prometheus"
	// DefaultMetricsProvider indicates what the preferred metrics provider is
	DefaultMetricsProvider = Prometheus

	// ErrInvalidTracingProvider is a sentinel error value
	ErrInvalidTracingProvider = errors.New("invalid tracing provider")
	// Jaeger is one of our supported tracing providers
	Jaeger tracingProvider = "jaeger"
	// DefaultTracingProvider indicates what the preferred tracing provider is
	DefaultTracingProvider = Jaeger
)

// MetricsSettings contains settings about how we report our metrics
type MetricsSettings struct {
	MetricsProvider                  metricsProvider `mapstructure:"metrics_provider"`
	TracingProvider                  tracingProvider `mapstructure:"tracing_provider"`
	DBMetricsCollectionInterval      time.Duration   `mapstructure:"database_metrics_collection_interval"`
	RuntimeMetricsCollectionInterval time.Duration   `mapstructure:"runtime_metrics_collection_interval"`
}

// ProvideInstrumentationHandler provides an instrumentation handler
func (cfg *ServerConfig) ProvideInstrumentationHandler(logger logging.Logger) (metrics.InstrumentationHandler, error) {
	if err := metrics.RegisterDefaultViews(); err != nil {
		return nil, errors.Wrap(err, "registering default metric views")
	}
	_ = metrics.RecordRuntimeStats(cfg.Metrics.RuntimeMetricsCollectionInterval)

	log := logger.WithValue("metrics_provider", cfg.Metrics.MetricsProvider)
	log.Debug("setting metrics provider")

	switch cfg.Metrics.MetricsProvider {
	case Prometheus, DefaultMetricsProvider:
		p, err := prometheus.NewExporter(prometheus.Options{
			OnError: func(err error) {
				logger.Error(err, "setting up prometheus export")
			},
			Namespace: string(MetricsNamespace),
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create Prometheus exporter")
		}
		view.RegisterExporter(p)
		log.Debug("metrics provider registered")
		return p, nil
	default:
		return nil, ErrInvalidMetricsProvider
	}
}

// ProvideTracing provides an instrumentation handler
func (cfg *ServerConfig) ProvideTracing(logger logging.Logger) error {
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(1)})

	log := logger.WithValue("tracing_provider", cfg.Metrics.TracingProvider)
	log.Info("setting tracing provider")

	switch cfg.Metrics.TracingProvider {
	case Jaeger, DefaultTracingProvider:
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
		log.Debug("tracing provider registered")
	default:
		return ErrInvalidTracingProvider
	}

	return nil
}
