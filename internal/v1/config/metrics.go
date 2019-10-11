package config

import (
	"errors"
	"fmt"
	"math"
	"os"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics"

	"contrib.go.opencensus.io/exporter/jaeger"
	"contrib.go.opencensus.io/exporter/prometheus"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

const (
	// MetricsNamespace is the namespace under which we register metrics
	MetricsNamespace = "todo_server"

	// MinimumRuntimeCollectionInterval is the smallest interval we can collect metrics at
	// this value is used to guard against zero values
	MinimumRuntimeCollectionInterval = time.Second
)

type (
	metricsProvider string
	tracingProvider string
)

var (
	// ErrInvalidMetricsProvider is a sentinel error value
	ErrInvalidMetricsProvider = errors.New("invalid metrics provider")
	// Prometheus represents the popular time series database
	Prometheus metricsProvider = "prometheus"
	// DefaultMetricsProvider indicates what the preferred metrics provider is
	DefaultMetricsProvider = Prometheus

	// ErrInvalidTracingProvider is a sentinel error value
	ErrInvalidTracingProvider = errors.New("invalid tracing provider")
	// Jaeger represents the popular distributed tracing server
	Jaeger tracingProvider = "jaeger"
	// DefaultTracingProvider indicates what the preferred tracing provider is
	DefaultTracingProvider = Jaeger
)

// ProvideInstrumentationHandler provides an instrumentation handler
func (cfg *ServerConfig) ProvideInstrumentationHandler(logger logging.Logger) (metrics.InstrumentationHandler, error) {
	if err := metrics.RegisterDefaultViews(); err != nil {
		return nil, fmt.Errorf("registering default metric views: %w", err)
	}
	_ = metrics.RecordRuntimeStats(time.Duration(
		math.Max(
			float64(MinimumRuntimeCollectionInterval),
			float64(cfg.Metrics.RuntimeMetricsCollectionInterval),
		),
	))

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
			return nil, fmt.Errorf("failed to create Prometheus exporter: %w", err)
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
