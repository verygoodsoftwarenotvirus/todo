package events

import (
	"context"

	"github.com/stretchr/testify/mock"
)

var _ Producer = (*MockProducer)(nil)

// MockProducer implements our interface.
type MockProducer struct {
	mock.Mock
}

// Publish implements our interface.
func (m *MockProducer) Publish(ctx context.Context, data interface{}) error {
	return m.Called(ctx, data).Error(0)
}

// Stop implements our interface.
func (m *MockProducer) Stop() {
	m.Called()
}

// MockProducerProvider implements our interface.
type MockProducerProvider struct {
	mock.Mock
}

// ProviderProducer implements our interface.
func (m *MockProducerProvider) ProviderProducer(topic string) (Producer, error) {
	args := m.Called(topic)

	return args.Get(0).(Producer), args.Error(1)
}
