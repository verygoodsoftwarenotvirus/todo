package publishers

import (
	"fmt"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
)

type (
	// Provider is used to indicate what messaging provider we'll use.
	Provider string

	// Config is used to indicate how the messaging provider should be configured.
	Config struct {
		_ struct{}

		Provider     Provider
		QueueAddress MessageQueueAddress
	}
)

// ProvidePublisherProvider provides a PublisherProvider.
func ProvidePublisherProvider(logger logging.Logger, c *Config) (PublisherProvider, error) {
	p := strings.ToLower(strings.TrimSpace(string(c.Provider)))
	switch p {
	case "redis":
		return ProvideRedisProducerProvider(logger, string(c.QueueAddress)), nil
	default:
		return nil, fmt.Errorf("invalid provider: %q", c.Provider)
	}
}
