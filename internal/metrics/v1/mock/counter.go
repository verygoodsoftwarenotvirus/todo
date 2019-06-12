package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1"

	"github.com/stretchr/testify/mock"
)

var _ metrics.UnitCounter = (*UnitCounter)(nil)

// UnitCounter is a mock metrics.UnitCounter
type UnitCounter struct {
	mock.Mock
}

// Increment implements our UnitCounter interface
func (m *UnitCounter) Increment(ctx context.Context) {
	m.Called()
}

// IncrementBy implements our UnitCounter interface
func (m *UnitCounter) IncrementBy(ctx context.Context, val uint64) {
	m.Called(val)
}

// Decrement implements our UnitCounter interface
func (m *UnitCounter) Decrement(ctx context.Context) {
	m.Called()
}
