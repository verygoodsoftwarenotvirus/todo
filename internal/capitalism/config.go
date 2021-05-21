package capitalism

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	StripeProvider = "stripe"
)

type (
	Config struct {
		Enabled  bool          `json:"enabled" mapstructure:"enabled" toml:"enabled"`
		Provider string        `json:"provider" mapstructure:"provider" toml:"provider"`
		Stripe   *StripeConfig `json:"stripe" mapstructure:"stripe" toml:"stripe"`
	}

	// StripeConfig configures our Stripe interface.
	StripeConfig struct {
		APIKey        string `json:"api_key" mapstructure:"api_key" toml:"api_key"`
		SuccessURL    string `json:"success_url" mapstructure:"success_url" toml:"success_url"`
		CancelURL     string `json:"cancel_url" mapstructure:"cancel_url" toml:"cancel_url"`
		WebhookSecret string `json:"webhook_secret" mapstructure:"webhook_secret" toml:"webhook_secret"`
	}
)

var _ validation.ValidatableWithContext = (*StripeConfig)(nil)

// ValidateWithContext validates a StripeConfig struct.
func (cfg *StripeConfig) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.APIKey, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a StripeConfig struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	if !cfg.Enabled {
		return nil
	}

	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Provider, validation.Required),
		validation.Field(&cfg.Stripe, validation.When(cfg.Provider == StripeProvider)),
	)
}
