package config

import (
	"context"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// MetaSettings is primarily used for development.
type MetaSettings struct {
	// Debug enables debug mode service-wide
	// NOTE: this debug should override all other debugs, which is to say, if this is enabled, all of them are enabled.
	Debug bool `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
	// StartupDeadline indicates how long the service can take to spin up. This includes database migrations, configuring services, etc.
	StartupDeadline time.Duration `json:"startup_deadline" mapstructure:"startup_deadline" toml:"startup_deadline,omitempty"`
	// RunMode indicates the current run mode
	RunMode runMode `json:"run_mode" mapstructure:"run_mode" toml:"run_mode,omitempty"`
}

// Validate validates an MetaSettings struct.
func (s MetaSettings) Validate(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	return validation.ValidateStructWithContext(ctx, &s,
		validation.Field(&s.RunMode, validation.In(TestingRunMode, DevelopmentRunMode, ProductionRunMode)),
		validation.Field(&s.StartupDeadline, validation.Required),
	)
}
