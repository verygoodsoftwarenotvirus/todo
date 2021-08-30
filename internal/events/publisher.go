package events

import (
	"context"
	"encoding/json"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"gocloud.dev/pubsub"
)

var (
	_ Publisher = (*publisher)(nil)
)

type (
	// Publisher is our interface for interacting with the myriad pub/sub systems.
	Publisher interface {
		PublishEvent(ctx context.Context, data interface{}, extras map[string]string) error
	}

	publisher struct {
		logger logging.Logger
		tracer tracing.Tracer
		topic  *pubsub.Topic
	}
)

// ProvidePublisher provides a Publisher.
func ProvidePublisher(ctx context.Context, logger logging.Logger, cfg *Config) (Publisher, error) {
	if cfg == nil {
		return nil, errNilConfig
	}

	if !cfg.Enabled {
		return &NoopEventPublisher{}, nil
	}

	topic, err := ProvidePublishTopic(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("error initializing subscription: %w", err)
	}

	ep := &publisher{
		logger: logging.EnsureLogger(logger),
		tracer: tracing.NewTracer(fmt.Sprintf("event_publisher_%s", cfg.SubscriptionIdentifier)),
		topic:  topic,
	}

	return ep, nil
}

// PublishEvent takes the data received, marshals it to JSON, and publishes it to the event queue.
func (p *publisher) PublishEvent(ctx context.Context, data interface{}, extraInfo map[string]string) error {
	ctx, span := p.tracer.StartSpan(ctx)
	defer span.End()

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling provided data to JSON: %w", err)
	}

	if err = p.topic.Send(ctx, &pubsub.Message{
		Body:     jsonBytes,
		Metadata: extraInfo,
	}); err != nil {
		return err
	}

	return nil
}

// NoopEventPublisher is a Publisher that deliberately does nothing.
type NoopEventPublisher struct{}

// PublishEvent satisfies our interface contract.
func (n *NoopEventPublisher) PublishEvent(_ context.Context, _ interface{}, _ map[string]string) error {
	return nil
}
