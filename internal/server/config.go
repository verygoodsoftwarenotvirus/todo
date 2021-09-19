package server

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	// Config describes the settings pertinent to the HTTP serving portion of the service.
	Config struct {
		_ struct{}

		StartupDeadline time.Duration `json:"startup_deadline" mapstructure:"startup_deadline" toml:"startup_deadline,omitempty"`
		HTTPPort        uint16        `json:"http_port" mapstructure:"http_port" toml:"http_port,omitempty"`
		Debug           bool          `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
	}
)

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.HTTPPort, validation.Required),
		validation.Field(&cfg.StartupDeadline, validation.Required),
	)
}
