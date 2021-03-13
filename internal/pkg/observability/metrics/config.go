package metrics

import (
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel/unit"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"

	otelprom "go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/metric"
)

const (
	// defaultNamespace is the default namespace under which we register metrics.
	defaultNamespace = "todo_server"

	instrumentationVersion = "1.0.0"

	// MinimumRuntimeCollectionInterval is the smallest interval we can collect metrics at
	// this value is used to guard against zero values.
	MinimumRuntimeCollectionInterval = float64(time.Second)

	// DefaultMetricsCollectionInterval is the default amount of time we wait between runtime metrics queries.
	DefaultMetricsCollectionInterval = 2 * time.Second

	// Prometheus represents the popular time series database.
	Prometheus = "prometheus"
)

type (
	// Config contains settings related to .
	Config struct {
		// Provider indicates where our metrics should go.
		Provider string `json:"provider" mapstructure:"provider" toml:"provider,omitempty"`
		// RouteToken indicates how the metrics route should be authenticated.
		RouteToken string `json:"route_token" mapstructure:"route_token" toml:"route_token,omitempty"`
	}
)

// ProvideInstrumentationHandler provides an instrumentation handler.
func (cfg *Config) ProvideInstrumentationHandler(logger logging.Logger) (InstrumentationHandler, error) {
	logger = logger.WithValue("metrics_provider", cfg.Provider)
	logger.Debug("setting metrics provider")

	switch strings.TrimSpace(strings.ToLower(cfg.Provider)) {
	case Prometheus:
		exporter, err := otelprom.InstallNewPipeline(otelprom.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to initialize prometheus exporter: %w", err)
		}

		return exporter, nil
	default:
		// not a crime to not need metrics
		return nil, nil
	}
}

// ProvideUnitCounterProvider provides an instrumentation handler.
func (cfg *Config) ProvideUnitCounterProvider(logger logging.Logger) (UnitCounterProvider, error) {
	logger = logger.WithValue("metrics_provider", cfg.Provider)
	logger.Debug("setting up meter")

	switch strings.TrimSpace(strings.ToLower(cfg.Provider)) {
	case Prometheus:
		exporter, err := otelprom.NewExportPipeline(otelprom.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to initialize prometheus exporter: %w", err)
		}

		meterProvider := exporter.MeterProvider()
		meter := metric.Must(meterProvider.Meter(defaultNamespace))

		return func(name, description string) UnitCounter {
			counter := meter.NewInt64UpDownCounter(
				name,
				metric.WithInstrumentationVersion(instrumentationVersion),
				metric.WithInstrumentationName(name),
				metric.WithUnit(unit.Dimensionless),
				metric.WithDescription(description),
			)

			return &unitCounter{counter: counter}
		}, nil
	default:
		// not a crime to not need metrics
		return nil, nil
	}
}
