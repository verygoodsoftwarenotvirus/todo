package mockconsumers

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/messagequeue/consumers"

	"github.com/stretchr/testify/mock"
)

var _ consumers.ConsumerProvider = (*ConsumerProvider)(nil)

// ConsumerProvider is a mock consumers.ConsumerProvider.
type ConsumerProvider struct {
	mock.Mock
}

// ProviderConsumer implements the interface.
func (m *ConsumerProvider) ProviderConsumer(ctx context.Context, topic string, handlerFunc func(context.Context, []byte) error) (consumers.Consumer, error) {
	args := m.Called(ctx, topic, handlerFunc)

	return args.Get(0).(consumers.Consumer), args.Error(1)
}

// Consumer is a mock consumers.Consumer.
type Consumer struct {
	mock.Mock
}

// Consume implements the interface.
func (m *Consumer) Consume(stopChan chan bool, errors chan error) {
	m.Called(stopChan, errors)
}
