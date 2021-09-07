package events

import (
	"github.com/stretchr/testify/mock"
)

// MockProducer implements our interface.
type MockProducer struct {
	mock.Mock
}

// Publish implements our interface.
func (m *MockProducer) Publish(message []byte) error {
	return m.Called(message).Error(0)
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
