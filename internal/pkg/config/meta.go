package config

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// MetaSettings is primarily used for development.
type MetaSettings struct {
	RunMode runMode `json:"run_mode" mapstructure:"run_mode" toml:"run_mode,omitempty"`
	Debug   bool    `json:"debug" mapstructure:"debug" toml:"debug,omitempty"` // RunMode indicates the current run mode
	// Debug enables debug mode service-wide
	// NOTE: this debug should override all other debugs, which is to say, if this is enabled, all of them are enabled.
}

// Validate validates an MetaSettings struct.
func (s MetaSettings) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &s,
		validation.Field(&s.RunMode, validation.In(TestingRunMode, DevelopmentRunMode, ProductionRunMode)),
	)
}
