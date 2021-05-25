package events

import (
	"context"
)

// NoopEventPublisher is an EventPublisher that deliberately does nothing.
type NoopEventPublisher struct{}

func (n *NoopEventPublisher) PublishEvent(_ context.Context, _ interface{}, _ map[string]string) error {
	return nil
}
