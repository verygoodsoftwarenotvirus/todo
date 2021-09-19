package publishers

import (
	"bytes"
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"github.com/go-redis/redis/v8"
)

type (
	redisPublisher struct {
		tracer      tracing.Tracer
		encoder     encoding.ClientEncoder
		logger      logging.Logger
		redisClient *redis.Client
		topic       string
	}
)

// provideRedisPublisher provides a redis-backed Publisher.
func provideRedisPublisher(logger logging.Logger, addr MessageQueueAddress, topic string) *redisPublisher {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     string(addr),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return &redisPublisher{
		redisClient: redisClient,
		topic:       topic,
		encoder:     encoding.ProvideClientEncoder(logger, encoding.ContentTypeJSON),
		logger:      logging.EnsureLogger(logger),
		tracer:      tracing.NewTracer(fmt.Sprintf("%s_writer", topic)),
	}
}

func (r *redisPublisher) Publish(ctx context.Context, data interface{}) error {
	_, span := r.tracer.StartSpan(ctx)
	defer span.End()

	r.logger.Debug("publishing message")

	var b bytes.Buffer
	if err := r.encoder.Encode(ctx, &b, data); err != nil {
		return observability.PrepareError(err, r.logger, span, "encoding topic message")
	}

	return r.redisClient.Publish(ctx, r.topic, b.Bytes()).Err()
}

type producerProvider struct {
	logger logging.Logger
	addr   string
}

// ProvideRedisProducerProvider returns a PublisherProvider for a given address.
func ProvideRedisProducerProvider(logger logging.Logger, queueAddress string) PublisherProvider {
	return &producerProvider{
		logger: logging.EnsureLogger(logger),
		addr:   queueAddress,
	}
}

// ProviderPublisher returns a Publisher for a given topic.
func (p *producerProvider) ProviderPublisher(topic string) (Publisher, error) {
	logger := logging.EnsureLogger(p.logger).WithValue("topic", topic)

	return provideRedisPublisher(logger, MessageQueueAddress(p.addr), topic), nil
}
