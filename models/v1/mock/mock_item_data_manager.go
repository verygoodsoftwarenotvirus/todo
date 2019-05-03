package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

var _ models.ItemDataManager = (*ItemDataManager)(nil)

// ItemDataManager is a mock item handler
type ItemDataManager struct {
	mock.Mock
}

// GetItem is a mock function
func (m *ItemDataManager) GetItem(ctx context.Context, itemID, userID uint64) (*models.Item, error) {
	args := m.Called(ctx, itemID, userID)
	return args.Get(0).(*models.Item), args.Error(1)
}

// GetItemCount is a mock function
func (m *ItemDataManager) GetItemCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (uint64, error) {
	args := m.Called(ctx, filter, userID)
	return args.Get(0).(uint64), args.Error(1)
}

// GetAllItemsCount is a mock function
func (m *ItemDataManager) GetAllItemsCount(ctx context.Context, filter *models.QueryFilter) (uint64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(uint64), args.Error(1)
}

// GetItems is a mock function
func (m *ItemDataManager) GetItems(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.ItemList, error) {
	args := m.Called(ctx, filter, userID)
	return args.Get(0).(*models.ItemList), args.Error(1)
}

// CreateItem is a mock function
func (m *ItemDataManager) CreateItem(ctx context.Context, input *models.ItemInput) (*models.Item, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*models.Item), args.Error(1)
}

// UpdateItem is a mock function
func (m *ItemDataManager) UpdateItem(ctx context.Context, updated *models.Item) error {
	return m.Called(ctx, updated).Error(0)
}

// DeleteItem is a mock function
func (m *ItemDataManager) DeleteItem(ctx context.Context, id uint64, userID uint64) error {
	return m.Called(ctx, id, userID).Error(0)
}
