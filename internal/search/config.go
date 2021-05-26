package search

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// BleveProvider represents the bleve search index provider.
	BleveProvider = "bleve"
)

// Config contains settings regarding search indices.
type Config struct {
	Provider string `json:"provider" mapstructure:"provider" toml:"provider,omitempty"`
}

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Provider, validation.In(BleveProvider)),
	)
}
