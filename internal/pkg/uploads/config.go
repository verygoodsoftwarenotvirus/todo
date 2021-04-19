package uploads

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/storage"
)

// Config contains settings regarding search indices.
type Config struct {
	Storage storage.Config `json:"storage_config" mapstructure:"storage_config" toml:"storage_config,omitempty"`
	Debug   bool           `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
}

// ValidateWithContext validates an Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Storage),
	)
}
