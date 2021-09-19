package frontend

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Config configures the frontend service.
type Config struct {
	_ struct{}

	UseFakeData bool `json:"use_fake_data" mapstructure:"use_fake_data" toml:"use_fake_data"`
}

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg)
}
