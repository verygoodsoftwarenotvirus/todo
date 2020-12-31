package observability

import (
	"context"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	// Config contains settings about how we report our metrics.
	Config struct {
		// Metrics contains metrics settings.
		Metrics metrics.Config `json:"metrics" mapstructure:"metrics" toml:"metrics,omitempty"`
		// Tracing contains tracing settings.
		Tracing tracing.Config `json:"tracing" mapstructure:"tracing" toml:"tracing,omitempty"`
		// RuntimeMetricsCollectionInterval  is the interval we collect runtime statistics at.
		RuntimeMetricsCollectionInterval time.Duration `json:"runtime_metrics_collection_interval" mapstructure:"runtime_metrics_collection_interval" toml:"runtime_metrics_collection_interval,omitempty"`
	}
)

// Validate validates a Config struct.
func (cfg Config) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.Metrics, validation.Required),
		validation.Field(&cfg.Tracing, validation.Required),
		validation.Field(&cfg.RuntimeMetricsCollectionInterval, validation.Required),
	)
}
