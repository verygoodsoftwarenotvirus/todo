package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1"

	"github.com/stretchr/testify/mock"
)

var _ metrics.UnitCounter = (*UnitCounter)(nil)

type UnitCounter struct {
	mock.Mock
}

func (m *UnitCounter) Increment(ctx context.Context) {
	m.Called()
}

func (m *UnitCounter) IncrementBy(ctx context.Context, val uint64) {
	m.Called(val)
}

func (m *UnitCounter) Decrement(ctx context.Context) {
	m.Called()
}
