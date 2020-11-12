package config

import (
	"errors"
	"fmt"
	"math"
	"os"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics"

	"contrib.go.opencensus.io/exporter/jaeger"
	"contrib.go.opencensus.io/exporter/prometheus"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

const (
	// MetricsNamespace is the namespace under which we register metrics.
	MetricsNamespace = "todo_server"

	// MinimumRuntimeCollectionInterval is the smallest interval we can collect metrics at
	// this value is used to guard against zero values.
	MinimumRuntimeCollectionInterval = time.Second

	// DefaultMetricsCollectionInterval is the default amount of time we wait between runtime metrics queries.
	DefaultMetricsCollectionInterval = 2 * time.Second

	// Prometheus represents the popular time series database.
	Prometheus metricsProvider = "prometheus"
	// DefaultMetricsProvider indicates what the preferred metrics provider is.
	DefaultMetricsProvider = Prometheus
	// Jaeger represents the popular distributed tracing server.
	Jaeger tracingProvider = "jaeger"
)

type (
	metricsProvider string
	tracingProvider string

	// MetricsSettings contains settings about how we report our metrics.
	MetricsSettings struct {
		// MetricsProvider indicates where our metrics should go.
		MetricsProvider metricsProvider `json:"metrics_provider" mapstructure:"metrics_provider" toml:"metrics_provider,omitempty"`
		// TracingProvider indicates where our traces should go.
		TracingProvider tracingProvider `json:"tracing_provider" mapstructure:"tracing_provider" toml:"tracing_provider,omitempty"`
		// DBMetricsCollectionInterval is the interval we collect database statistics at.
		DBMetricsCollectionInterval time.Duration `json:"database_metrics_collection_interval" mapstructure:"database_metrics_collection_interval" toml:"database_metrics_collection_interval,omitempty"`
		// RuntimeMetricsCollectionInterval  is the interval we collect runtime statistics at.
		RuntimeMetricsCollectionInterval time.Duration `json:"runtime_metrics_collection_interval" mapstructure:"runtime_metrics_collection_interval" toml:"runtime_metrics_collection_interval,omitempty"`
	}
)

var (
	// ErrInvalidTracingProvider is a sentinel error value.
	ErrInvalidTracingProvider = errors.New("invalid tracing provider")
)

// ProvideInstrumentationHandler provides an instrumentation handler.
func (cfg *ServerConfig) ProvideInstrumentationHandler(logger logging.Logger) metrics.InstrumentationHandler {
	logger = logger.WithValue("metrics_provider", cfg.Metrics.MetricsProvider)
	logger.Debug("setting metrics provider")

	switch cfg.Metrics.MetricsProvider {
	case Prometheus:
		p, err := prometheus.NewExporter(
			prometheus.Options{
				OnError: func(err error) {
					logger.Error(err, "setting up prometheus export")
				},
				Namespace: MetricsNamespace,
			},
		)

		if err != nil {
			logger.Error(err, "failed to create Prometheus exporter")
			return nil
		}

		view.RegisterExporter(p)
		logger.Debug("metrics provider registered")

		if err := metrics.RegisterDefaultViews(); err != nil {
			logger.Error(err, "registering default metric views")
			return nil
		}

		metrics.RecordRuntimeStats(time.Duration(
			math.Max(
				float64(MinimumRuntimeCollectionInterval),
				float64(cfg.Metrics.RuntimeMetricsCollectionInterval),
			),
		))

		return p
	default:
		return nil
	}
}

// ProvideTracing provides an instrumentation handler.
func (cfg *ServerConfig) ProvideTracing(logger logging.Logger) error {
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(1)})

	log := logger.WithValue("tracing_provider", cfg.Metrics.TracingProvider)
	log.Info("setting tracing provider")

	switch cfg.Metrics.TracingProvider {
	case Jaeger:
		ah := os.Getenv("JAEGER_AGENT_HOST")
		ap := os.Getenv("JAEGER_AGENT_PORT")
		sn := os.Getenv("JAEGER_SERVICE_NAME")

		if ah != "" && ap != "" && sn != "" {
			je, err := jaeger.NewExporter(jaeger.Options{
				AgentEndpoint: fmt.Sprintf("%s:%s", ah, ap),
				Process:       jaeger.Process{ServiceName: sn},
			})
			if err != nil {
				return fmt.Errorf("failed to create Jaeger exporter: %w", err)
			}

			trace.RegisterExporter(je)
			log.Debug("tracing provider registered")
		}
	}

	return nil
}
