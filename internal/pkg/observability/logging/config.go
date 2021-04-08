package logging

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Config configures logging for a service.
type Config struct {
	Level    string `json:"level" mapstructure:"level" toml:"level,omitempty"`
	Provider string `json:"provider" mapstructure:"provider" toml:"provider,omitempty"`
	Enabled  bool   `json:"enabled" mapstructure:"enabled" toml:"enabled,omitempty"` // Level indicates what level the logger should operate on
	// Enabled determines if logging is enabled.
}

// Validate validates a Config struct.
func (cfg Config) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg)
}
