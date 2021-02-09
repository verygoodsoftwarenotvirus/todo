package routing

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ChiProviderKey is the string we use to refer to chi.
	ChiProviderKey = "chi"
)

var (
	validProviders = []string{ChiProviderKey}
)

// Config configures our router.
type Config struct {
	// Provider indicates which library we'd like to use for routing
	Provider string `json:"provider" mapstructure:"provider" toml:"provider,omitempty"`
}

// Validate validates a router config struct.
func (cfg Config) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.Provider, validation.In(validProviders)),
	)
}
