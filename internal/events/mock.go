package events

import (
	"context"
	"github.com/stretchr/testify/mock"
)

var _ EventPublisher = (*mockEventPublisher)(nil)

type mockEventPublisher struct {
	mock.Mock
}

// Publish satisfies our interface contract.
func (m *mockEventPublisher) PublishEvent(ctx context.Context, data interface{}, extras map[string]string) error {
	return m.Called(ctx, data, extras).Error(0)
}
