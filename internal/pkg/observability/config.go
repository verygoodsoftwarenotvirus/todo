package observability

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	// Config contains settings about how we report our metrics.
	Config struct {
		// Tracing contains tracing settings.
		Tracing tracing.Config `json:"tracing" mapstructure:"tracing" toml:"tracing,omitempty"`
		// Metrics contains metrics settings.
		Metrics metrics.Config `json:"metrics" mapstructure:"metrics" toml:"metrics,omitempty"`
	}
)

// Validate validates a Config struct.
func (cfg *Config) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Metrics),
		validation.Field(&cfg.Tracing),
	)
}
