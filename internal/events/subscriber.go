package events

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"gocloud.dev/pubsub"
)

var (
	_ Subscriber = (*subscriber)(nil)
)

type (
	// Subscriber allows you to listen to/handle events on a given subscription.
	Subscriber interface {
		HandleEvents(waitPeriod time.Duration, stopCh chan bool, handler func(body []byte))
	}

	subscriber struct {
		logger       logging.Logger
		tracer       tracing.Tracer
		subscription *pubsub.Subscription
	}
)

// ProvideSubscriber provides an Subscriber.
func ProvideSubscriber(logger logging.Logger, sub *pubsub.Subscription, cfg *Config) (Subscriber, error) {
	if sub == nil {
		return nil, errNilSubscription
	}

	if cfg == nil {
		return nil, errNilConfig
	}

	if !cfg.Enabled {
		return &NoopEventSubscriber{}, nil
	}

	ep := &subscriber{
		logger:       logging.EnsureLogger(logger),
		tracer:       tracing.NewTracer(fmt.Sprintf("event_publisher_%s", cfg.Topic)),
		subscription: sub,
	}

	return ep, nil
}

func (p *subscriber) HandleEvents(waitPeriod time.Duration, stopCh chan bool, handler func(body []byte)) {
	ticker := time.NewTicker(waitPeriod)

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			msg, err := p.subscription.Receive(context.Background())
			if err != nil {
				p.logger.Error(err, "receiving subscription message")
			}

			handler(msg.Body)
			msg.Ack()
		}
	}
}

var _ Subscriber = (*NoopEventSubscriber)(nil)

// NoopEventSubscriber is a Subscriber that deliberately does nothing.
type NoopEventSubscriber struct{}

// HandleEvents satisfies our interface contract.
func (n *NoopEventSubscriber) HandleEvents(_ time.Duration, _ chan bool, _ func(body []byte)) {}
