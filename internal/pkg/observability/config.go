package observability

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"

	"contrib.go.opencensus.io/exporter/jaeger"
	"contrib.go.opencensus.io/exporter/prometheus"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

const (
	// MetricsNamespace is the namespace under which we register metrics.
	MetricsNamespace = "todo_server"

	// MinimumRuntimeCollectionInterval is the smallest interval we can collect metrics at
	// this value is used to guard against zero values.
	MinimumRuntimeCollectionInterval = float64(time.Second)

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

	// Config contains settings about how we report our metrics.
	Config struct {
		// MetricsProvider indicates where our metrics should go.
		MetricsProvider metricsProvider `json:"metrics_provider" mapstructure:"metrics_provider" toml:"metrics_provider,omitempty"`
		// TracingProvider indicates where our traces should go.
		TracingProvider tracingProvider `json:"tracing_provider" mapstructure:"tracing_provider" toml:"tracing_provider,omitempty"`
		// RuntimeMetricsCollectionInterval  is the interval we collect runtime statistics at.
		RuntimeMetricsCollectionInterval time.Duration `json:"runtime_metrics_collection_interval" mapstructure:"runtime_metrics_collection_interval" toml:"runtime_metrics_collection_interval,omitempty"`
	}
)

// Validate validates a Config struct.
func (cfg Config) Validate(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.MetricsProvider, validation.In(Prometheus)),
		validation.Field(&cfg.TracingProvider, validation.In(Jaeger)),
		validation.Field(&cfg.RuntimeMetricsCollectionInterval, validation.Required),
	)
}

// ProvideInstrumentationHandler provides an instrumentation handler.
func (cfg *Config) ProvideInstrumentationHandler(logger logging.Logger) metrics.InstrumentationHandler {
	logger = logger.WithValue("metrics_provider", cfg.MetricsProvider)
	logger.Debug("setting metrics provider")

	switch cfg.MetricsProvider {
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
				MinimumRuntimeCollectionInterval,
				float64(cfg.RuntimeMetricsCollectionInterval),
			),
		))

		return p
	default:
		return nil
	}
}

// InitializeTracer provides an instrumentation handler.
func (cfg *Config) InitializeTracer(logger logging.Logger) error {
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(1)})

	log := logger.WithValue("tracing_provider", cfg.TracingProvider)
	log.Info("setting tracing provider")

	switch cfg.TracingProvider {
	case Jaeger:
		agentHost := os.Getenv("JAEGER_AGENT_HOST")
		agentPort := os.Getenv("JAEGER_AGENT_PORT")
		serviceName := os.Getenv("JAEGER_SERVICE_NAME")

		if agentHost != "" && agentPort != "" && serviceName != "" {
			je, err := jaeger.NewExporter(jaeger.Options{
				AgentEndpoint: fmt.Sprintf("%s:%s", agentHost, agentPort),
				Process:       jaeger.Process{ServiceName: serviceName},
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
