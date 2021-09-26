package consumers

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-redis/redis/v8"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
)

type (
	redisConsumer struct {
		tracer       tracing.Tracer
		encoder      encoding.ClientEncoder
		logger       logging.Logger
		redisClient  *redis.Client
		handlerFunc  func(context.Context, []byte) error
		subscription *redis.PubSub
		topic        string
	}

	// RedisConfig configures a Redis-backed consumer.
	RedisConfig struct {
		QueueAddress MessageQueueAddress `json:"message_queue_address" mapstructure:"message_queue_address" toml:"message_queue_address,omitempty"`
	}
)

func provideRedisConsumer(ctx context.Context, logger logging.Logger, redisClient *redis.Client, topic string, handlerFunc func(context.Context, []byte) error) *redisConsumer {
	subscription := redisClient.Subscribe(ctx, topic)

	return &redisConsumer{
		redisClient:  redisClient,
		topic:        topic,
		subscription: subscription,
		handlerFunc:  handlerFunc,
		encoder:      encoding.ProvideClientEncoder(logger, encoding.ContentTypeJSON),
		logger:       logging.EnsureLogger(logger),
		tracer:       tracing.NewTracer(fmt.Sprintf("%s_consumer", topic)),
	}
}

// Consume reads messages and applies the handler to their payloads.
// Writes errors to the error chan if it isn't nil.
func (r *redisConsumer) Consume(stopChan chan bool, errors chan error) {
	if stopChan == nil {
		stopChan = make(chan bool, 1)
	}
	subChan := r.subscription.Channel()

	for {
		select {
		case msg := <-subChan:
			ctx := context.Background()
			if err := r.handlerFunc(ctx, []byte(msg.Payload)); err != nil {
				r.logger.Error(err, "handling message")
				if errors != nil {
					errors <- err
				}
			}
		case <-stopChan:
			return
		}
	}
}

type consumerProvider struct {
	logger           logging.Logger
	consumerCache    map[string]Consumer
	redisClient      *redis.Client
	consumerCacheHat sync.RWMutex
}

// ProvideRedisConsumerProvider returns a ConsumerProvider for a given address.
func ProvideRedisConsumerProvider(logger logging.Logger, queueAddress string) ConsumerProvider {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     queueAddress,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return &consumerProvider{
		logger:        logging.EnsureLogger(logger),
		redisClient:   redisClient,
		consumerCache: map[string]Consumer{},
	}
}

// ProviderConsumer returns a Consumer for a given topic.
func (p *consumerProvider) ProviderConsumer(ctx context.Context, topic string, handlerFunc func(context.Context, []byte) error) (Consumer, error) {
	logger := logging.EnsureLogger(p.logger).WithValue("topic", topic)

	p.consumerCacheHat.Lock()
	defer p.consumerCacheHat.Unlock()
	if cachedPub, ok := p.consumerCache[topic]; ok {
		return cachedPub, nil
	}

	c := provideRedisConsumer(ctx, logger, p.redisClient, topic, handlerFunc)
	p.consumerCache[topic] = c

	return c, nil
}
