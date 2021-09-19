package metrics

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	otelprom "go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/unit"
)

const (
	// defaultNamespace is the default namespace under which we register metrics.
	defaultNamespace = "todo_server"

	instrumentationVersion = "1.0.0"

	// minimumRuntimeCollectionInterval is the smallest interval we can collect metrics at
	// this value is used to guard against zero values.
	minimumRuntimeCollectionInterval = time.Second

	// DefaultMetricsCollectionInterval is the default amount of time we wait between runtime metrics queries.
	DefaultMetricsCollectionInterval = 2 * time.Second

	// Prometheus represents the popular time series database.
	Prometheus = "prometheus"
)

type (
	// Config contains settings related to .
	Config struct {
		_ struct{}

		// Provider indicates where our metrics should go.
		Provider string `json:"provider" mapstructure:"provider" toml:"provider,omitempty"`
		// RouteToken indicates how the metrics route should be authenticated.
		RouteToken string `json:"route_token" mapstructure:"route_token" toml:"route_token,omitempty"`
		// RuntimeMetricsCollectionInterval  is the interval we collect runtime statistics at.
		RuntimeMetricsCollectionInterval time.Duration `json:"runtime_metrics_collection_interval" mapstructure:"runtime_metrics_collection_interval" toml:"runtime_metrics_collection_interval,omitempty"`
	}
)

// ValidateWithContext validates the config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.RuntimeMetricsCollectionInterval, validation.Min(minimumRuntimeCollectionInterval)),
	)
}

var (
	// regretful state artifacts.
	prometheusExporterInitOnce sync.Once
	prometheusExporter         *otelprom.Exporter
)

func initiatePrometheusExporter() {
	prometheusExporterInitOnce.Do(func() {
		var err error
		if prometheusExporter, err = otelprom.InstallNewPipeline(otelprom.Config{}); err != nil {
			panic(err)
		}
	})
}

// ProvideInstrumentationHandler provides an instrumentation handler.
func (cfg *Config) ProvideInstrumentationHandler(logger logging.Logger) (InstrumentationHandler, error) {
	logger = logger.WithValue("metrics_provider", cfg.Provider)
	logger.Debug("setting metrics provider")

	if err := runtime.Start(runtime.WithMinimumReadMemStatsInterval(cfg.RuntimeMetricsCollectionInterval)); err != nil {
		return nil, fmt.Errorf("failed to start runtime metrics collection: %w", err)
	}

	switch strings.TrimSpace(strings.ToLower(cfg.Provider)) {
	case Prometheus:
		initiatePrometheusExporter()

		return prometheusExporter, nil
	default:
		// not a crime to not need metrics
		return nil, nil
	}
}

// ProvideUnitCounterProvider provides an instrumentation handler.
func (cfg *Config) ProvideUnitCounterProvider(logger logging.Logger) (UnitCounterProvider, error) {
	p := strings.TrimSpace(strings.ToLower(cfg.Provider))

	logger = logger.WithValue("metrics_provider", p)
	logger.Debug("setting up meter")

	switch p {
	case Prometheus:
		initiatePrometheusExporter()

		meterProvider := prometheusExporter.MeterProvider()
		mustMeter := metric.Must(meterProvider.Meter(defaultNamespace, metric.WithInstrumentationVersion(instrumentationVersion)))

		logger.Debug("meter initialized successfully")

		return func(name, description string) UnitCounter {
			l := logger.WithValue("name", name)

			counter := mustMeter.NewInt64Counter(
				fmt.Sprintf("%s_count", name),
				metric.WithUnit(unit.Dimensionless),
				metric.WithDescription(description),
				metric.WithInstrumentationName(name),
				metric.WithInstrumentationVersion(instrumentationVersion),
			)

			l.Debug("returning wrapped unit counter")

			return &unitCounter{counter: counter}
		}, nil
	default:
		logger.Debug("nil unit counter provider")
		// not a crime to not need metrics
		return nil, nil
	}
}
