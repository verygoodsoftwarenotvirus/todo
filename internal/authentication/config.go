package authentication

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	argon2Provider = "argon2"
)

// Config configures the authentication portion of the service.
type Config struct {
	Provider string `json:"provider" mapstructure:"provider" xml:"provider"`
}

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates the Config.
func (c *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.Provider, validation.In(argon2Provider)),
	)
}
