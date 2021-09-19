package publishers

import (
	"context"

	"github.com/stretchr/testify/mock"
)

var _ Publisher = (*MockProducer)(nil)

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

// ProviderPublisher implements our interface.
func (m *MockProducerProvider) ProviderPublisher(topic string) (Publisher, error) {
	args := m.Called(topic)

	return args.Get(0).(Publisher), args.Error(1)
}
