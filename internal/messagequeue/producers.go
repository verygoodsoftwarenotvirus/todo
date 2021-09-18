package messagequeue

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
)

type (
	// Producer produces events onto a queue.
	Producer interface {
		Publish(ctx context.Context, data interface{}) error
	}

	// ProducerConfig configures a topic producer.
	ProducerConfig struct {
		Address EventQueueAddress
	}
)

// TopicWriter produces messages on a topic.
type TopicWriter struct {
	logger  logging.Logger
	tracer  tracing.Tracer
	encoder encoding.ClientEncoder
	rdb     *redis.Client
	topic   string
}

// Publish publishes a message onto a topic queue.
func (w *TopicWriter) Publish(ctx context.Context, data interface{}) error {
	_, span := w.tracer.StartSpan(ctx)
	defer span.End()

	w.logger.Debug("publishing message")

	var b bytes.Buffer
	if err := w.encoder.Encode(ctx, &b, data); err != nil {
		return observability.PrepareError(err, w.logger, span, "encoding topic message")
	}

	return w.rdb.Publish(ctx, w.topic, b.Bytes()).Err()
}

type producerProvider struct {
	logger logging.Logger
	addr   string
}

// ProviderProducer returns a Producer for a given topic.
func (p *producerProvider) ProviderProducer(topic string) (Producer, error) {
	logger := logging.EnsureLogger(p.logger).WithValue("topic", topic)

	rdb := redis.NewClient(&redis.Options{
		Addr:     p.addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	tw := &TopicWriter{
		rdb:     rdb,
		topic:   topic,
		encoder: encoding.ProvideClientEncoder(logger, encoding.ContentTypeJSON),
		logger:  logger,
		tracer:  tracing.NewTracer(fmt.Sprintf("%s_writer", topic)),
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
