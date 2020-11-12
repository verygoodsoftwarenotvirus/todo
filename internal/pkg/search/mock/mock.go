package mocksearch

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"

	"github.com/stretchr/testify/mock"
)

var _ search.IndexManager = (*IndexManager)(nil)

// IndexManager is a mock IndexManager.
type IndexManager struct {
	mock.Mock
}

// Index implements our interface.
func (m *IndexManager) Index(ctx context.Context, id uint64, value interface{}) error {
	args := m.Called(ctx, id, value)
	return args.Error(0)
}

// Search implements our interface.
func (m *IndexManager) Search(ctx context.Context, query string, userID uint64) (ids []uint64, err error) {
	args := m.Called(ctx, query, userID)
	return args.Get(0).([]uint64), args.Error(1)
}

// SearchForAdmin implements our interface.
func (m *IndexManager) SearchForAdmin(ctx context.Context, query string) (ids []uint64, err error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]uint64), args.Error(1)
}

// Delete implements our interface.
func (m *IndexManager) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
