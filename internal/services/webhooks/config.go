package webhooks

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Config represents our database configuration.
type Config struct {
	_                    struct{}
	Logging              logging.Config `json:"logging" mapstructure:"logging" toml:"logging,omitempty"`
	PreWritesTopicName   string         `json:"pre_writes_topic_name" mapstructure:"pre_writes_topic_name" toml:"pre_writes_topic_name,omitempty"`
	PreArchivesTopicName string         `json:"pre_archives_topic_name" mapstructure:"pre_archives_topic_name" toml:"pre_archives_topic_name,omitempty"`
	Async                bool           `json:"async" mapstructure:"async" toml:"async,omitempty"`
	Debug                bool           `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
}

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg)
}
