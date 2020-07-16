package mocksearch

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/search"

	"github.com/stretchr/testify/mock"
)

var _ search.IndexManager = (*MockIndexManager)(nil)

// MockIndexManager is a mock IndexManager
type MockIndexManager struct {
	mock.Mock
}

// Index implements our interface
func (m *MockIndexManager) Index(ctx context.Context, id uint64, value interface{}) error {
	args := m.Called(ctx, id, value)
	return args.Error(0)
}

// Search implements our interface
func (m *MockIndexManager) Search(ctx context.Context, query string, userID uint64) (ids []uint64, err error) {
	args := m.Called(ctx, query, userID)
	return args.Get(0).([]uint64), args.Error(1)
}

// Delete implements our interface
func (m *MockIndexManager) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
