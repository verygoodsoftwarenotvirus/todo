package config

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/storage/gocloud"
)

// Config contains settings regarding search indices.
type Config struct {
	// Debug determines if debug logging or other development conditions are active.
	Debug bool `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
	// Provider indicates what database we'll connect to (postgres, mysql, etc.)
	Provider string `json:"provider" mapstructure:"provider" toml:"provider,omitempty"`
	// StorageConfig configures our storage provider
	StorageConfig *gocloud.UploaderConfig `json:"storage_config" mapstructure:"storage_config" toml:"storage_config,omitempty"`
}

// Validate validates an Config struct.
func (cfg Config) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.StorageConfig),
	)
}
