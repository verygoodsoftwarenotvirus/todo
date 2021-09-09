package items

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Config configures the service.
type Config struct {
	Logging                logging.Config `json:"logging" mapstructure:"logging" toml:"logging,omitempty"`
	PendingWritesTopicName string         `json:"pending_writes_topic" mapstructure:"pending_writes_topic" toml:"pending_writes_topic,omitempty"`
	SearchIndexPath        string         `json:"searchIndexPath" mapstructure:"search_index_path" toml:"search_index_path,omitempty"`
}

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		cfg,
		validation.Field(&cfg.Logging, validation.Required),
		validation.Field(&cfg.SearchIndexPath, validation.Required),
	)
}
