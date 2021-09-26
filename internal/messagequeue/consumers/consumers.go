package consumers

import (
	"context"
)

type (
	// Consumer produces events onto a queue.
	Consumer interface {
		Consume(stopChan chan bool, errors chan error)
	}

	// MessageQueueAddress is a simple string alias for the location of our event queue server.
	MessageQueueAddress string

	// Topic is a simple string alias for the location of our event queue server.
	Topic string

	// ConsumerProvider is a function that provides a Consumer for a given topic.
	ConsumerProvider interface {
		ProviderConsumer(ctx context.Context, topic string, handlerFunc func(context.Context, []byte) error) (Consumer, error)
	}
)
