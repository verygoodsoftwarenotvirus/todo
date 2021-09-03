package events

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"github.com/nsqio/go-nsq"
)

type (
	// Producer produces events onto a queue.
	Producer interface {
		Publish(message []byte) error
		Stop()
	}

	// ProducerConfig configures a topic producer.
	ProducerConfig struct {
		Address EventQueueAddress
	}
)

type TopicProducer struct {
	topic    string
	logger   logging.Logger
	tracer   tracing.Tracer
	producer *nsq.Producer
}

//Publish publishes a message onto a topic queue.
func (w *TopicProducer) Publish(message []byte) error {
	w.logger.Debug("publishing message")

	return w.producer.Publish(w.topic, message)
}

// Stop stops the producer.
func (w *TopicProducer) Stop() {
	w.producer.Stop()
}

type producerProvider struct {
	logger logging.Logger
	addr   string
}

// ProviderProducer returns a Producer for a given topic.
func (p *producerProvider) ProviderProducer(topic string) (Producer, error) {
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer(p.addr, config)
	if err != nil {
		return nil, err
	}

	tw := &TopicProducer{
		topic:    topic,
		logger:   logging.EnsureLogger(p.logger).WithValue("topic", topic),
		tracer:   tracing.NewTracer(fmt.Sprintf("%s_writer", topic)),
		producer: producer,
	}

	return tw, nil
}

// NewProducerProvider returns a ProducerProvider for a given address.
func NewProducerProvider(logger logging.Logger, addr EventQueueAddress) ProducerProvider {
	return &producerProvider{
		logger: logging.EnsureLogger(logger),
		addr:   string(addr),
	}
}
