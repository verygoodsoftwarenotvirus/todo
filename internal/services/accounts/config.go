package accounts

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
)

// Config configures the service.
type Config struct {
	_                  struct{}
	Logging            logging.Config `json:"logging" mapstructure:"logging" toml:"logging,omitempty"`
	PreWritesTopicName string         `json:"pre_writes_topic_name" mapstructure:"pre_writes_topic_name" toml:"pre_writes_topic_name,omitempty"`
}

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		cfg,
		validation.Field(&cfg.Logging, validation.Required),
		validation.Field(&cfg.PreWritesTopicName, validation.Required),
	)
}
