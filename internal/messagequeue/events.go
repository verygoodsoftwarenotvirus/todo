package messagequeue

type (
	// EventQueueAddress is a simple string alias for the location of our event queue server.
	EventQueueAddress string

	// ProducerProvider is a function that provides a Producer for a given topic.
	ProducerProvider interface {
		ProviderProducer(topic string) (Producer, error)
	}
)
