package audit

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Config represents our database configuration.
type Config struct {
	Debug   bool `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
	Enabled bool `json:"enabled" mapstructure:"enabled" toml:"enabled,omitempty"`
}

// Validate validates a Config struct.
func (cfg Config) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg)
}
