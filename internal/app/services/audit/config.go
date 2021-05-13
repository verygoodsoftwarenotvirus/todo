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

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg)
}
