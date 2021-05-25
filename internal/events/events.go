package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"gocloud.dev/pubsub"
)

var (
	errNilConfig = errors.New("nil Config provided")
)

type (
	EventPublisher interface {
		PublishEvent(ctx context.Context, data interface{}, extras map[string]string) error
	}

	eventPublisher struct {
		logger logging.Logger
		tracer tracing.Tracer
		topic  *pubsub.Topic
	}
)

// ProvideEventPublisher provides an EventPublisher.
func ProvideEventPublisher(ctx context.Context, logger logging.Logger, cfg *Config) (EventPublisher, error) {
	if cfg == nil || !cfg.Enabled {
		return &NoopEventPublisher{}, errNilConfig
	}

	topic, err := ProvideTopic(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("error initializing topic: %w", err)
	}

	ep := &eventPublisher{
		logger: logging.EnsureLogger(logger),
		tracer: tracing.NewTracer(fmt.Sprintf("event_publisher_%s", cfg.Name)),
		topic:  topic,
	}

	return ep, nil
}

func (p *eventPublisher) PublishEvent(ctx context.Context, data interface{}, extraInfo map[string]string) error {
	ctx, span := p.tracer.StartSpan(ctx)
	defer span.End()

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return p.topic.Send(ctx, &pubsub.Message{
		Body:     jsonBytes,
		Metadata: extraInfo,
	})
}
