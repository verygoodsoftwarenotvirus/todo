package config

import (
	"fmt"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/messagequeue/consumers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/messagequeue/publishers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
)

const (
	// ProviderRedis is used to refer to redis.
	ProviderRedis = "redis"
)

type (
	// Provider is used to indicate what messaging provider we'll use.
	Provider string

	// MessageQueueAddress is a simple string alias for the location of our event queue server.
	MessageQueueAddress string

	// RedisConfig configures a Redis-backed consumer.
	RedisConfig struct {
		QueueAddress MessageQueueAddress `json:"message_queue_address" mapstructure:"message_queue_address" toml:"message_queue_address,omitempty"`
	}

	// Config is used to indicate how the messaging provider should be configured.
	Config struct {
		_ struct{}

		Provider    Provider    `json:"provider" mapstructure:"provider" toml:"provider,omitempty"`
		RedisConfig RedisConfig `json:"redis" mapstructure:"redis" toml:"redis,omitempty"`
	}
)

func cleanString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// ProvideConsumerProvider provides a PublisherProvider.
func ProvideConsumerProvider(logger logging.Logger, c *Config) (consumers.ConsumerProvider, error) {
	switch cleanString(string(c.Provider)) {
	case ProviderRedis:
		return consumers.ProvideRedisConsumerProvider(logger, string(c.RedisConfig.QueueAddress)), nil
	default:
		return nil, fmt.Errorf("invalid provider: %q", c.Provider)
	}
}

// ProvidePublisherProvider provides a PublisherProvider.
func ProvidePublisherProvider(logger logging.Logger, c *Config) (publishers.PublisherProvider, error) {
	switch cleanString(string(c.Provider)) {
	case ProviderRedis:
		return publishers.ProvideRedisPublisherProvider(logger, string(c.RedisConfig.QueueAddress)), nil
	default:
		return nil, fmt.Errorf("invalid provider: %q", c.Provider)
	}
}
