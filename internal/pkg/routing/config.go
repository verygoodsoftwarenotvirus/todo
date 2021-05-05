package routing

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ChiProvider is the string we use to refer to chi.
	ChiProvider = "chi"
)

// Config configures our router.
type Config struct {
	Provider       string   `json:"provider" mapstructure:"provider" toml:"provider,omitempty"`
	AllowedOrigins []string `json:"allowed_origins" mapstructure:"allowed_origins" toml:"allowed_origins,omitempty"`
	AllowedHeaders []string `json:"allowed_heade" mapstructure:"allowed_headers" toml:"allowed_headers,omitempty"`
}

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a router config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Provider, validation.In(ChiProvider)),
	)
}
