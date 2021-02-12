package metrics

import (
	"log"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"

	"go.opentelemetry.io/otel/exporters/metric/prometheus"
)

const (
	// DefaultNamespace is the default namespace under which we register metrics.
	DefaultNamespace = "todo_server"

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
func (cfg *Config) ProvideInstrumentationHandler(logger logging.Logger) InstrumentationHandler {
	logger = logger.WithValue("metrics_provider", cfg.Provider)
	logger.Debug("setting metrics provider")

	switch strings.TrimSpace(strings.ToLower(cfg.Provider)) {
	case Prometheus:
		exporter, err := prometheus.InstallNewPipeline(prometheus.Config{})
		if err != nil {
			log.Panicf("failed to initialize prometheus exporter %v", err)
		}

		return exporter
	default:
		return nil
	}
}
