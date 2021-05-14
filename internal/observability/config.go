package observability

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	// Config contains settings about how we report our metrics.
	Config struct {
		Tracing tracing.Config `json:"tracing" mapstructure:"tracing" toml:"tracing,omitempty"`
		Metrics metrics.Config `json:"metrics" mapstructure:"metrics" toml:"metrics,omitempty"`
	}
)

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Metrics),
		validation.Field(&cfg.Tracing),
	)
}
