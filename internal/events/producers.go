package events

import (
	"bytes"
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"github.com/nsqio/go-nsq"
)

type (
	// Producer produces events onto a queue.
	Producer interface {
		Publish(ctx context.Context, data interface{}) error
		Stop()
	}

	// ProducerConfig configures a topic producer.
	ProducerConfig struct {
		Address EventQueueAddress
	}
)

// TopicProducer produces messages on a topic.
type TopicProducer struct {
	logger   logging.Logger
	tracer   tracing.Tracer
	encoder  encoding.ClientEncoder
	producer *nsq.Producer
	topic    string
}

// Publish publishes a message onto a topic queue.
func (w *TopicProducer) Publish(ctx context.Context, data interface{}) error {
	ctx, span := w.tracer.StartSpan(ctx)
	defer span.End()

	w.logger.Debug("publishing message")

	var b bytes.Buffer
	if err := w.encoder.Encode(ctx, &b, data); err != nil {
		return observability.PrepareError(err, w.logger, span, "encoding topic message")
	}

	return w.producer.Publish(w.topic, b.Bytes())
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

	logger := logging.EnsureLogger(p.logger).WithValue("topic", topic)

	tw := &TopicProducer{
		topic:    topic,
		encoder:  encoding.ProvideClientEncoder(logger, encoding.ContentTypeJSON),
		logger:   logger,
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
