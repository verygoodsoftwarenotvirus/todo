package publishers

import (
	"context"
)

type (
	// Publisher produces events onto a queue.
	Publisher interface {
		Publish(ctx context.Context, data interface{}) error
	}

	// MessageQueueAddress is a simple string alias for the location of our event queue server.
	MessageQueueAddress string

	// Topic is a simple string alias for the location of our event queue server.
	Topic string

	// PublisherProvider is a function that provides a Publisher for a given topic.
	PublisherProvider interface {
		ProviderPublisher(topic string) (Publisher, error)
	}
)
