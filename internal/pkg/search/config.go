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
	// Provider indicates who provides the search functionality.
	Provider string `json:"provider" mapstructure:"provider" toml:"provider,omitempty"`
	// ItemsIndexPath indicates where our items search index files should go.
	ItemsIndexPath IndexPath `json:"items_index_path" mapstructure:"items_index_path" toml:"items_index_path,omitempty"`
}

// Validate validates a Config struct.
func (cfg *Config) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Provider, validation.In(BleveProvider)),
		validation.Field(&cfg.ItemsIndexPath, validation.Required),
	)
}
