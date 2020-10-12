package mock

import (
	"context"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

var _ models.ItemDataManager = (*ItemDataManager)(nil)

// ItemDataManager is a mocked models.ItemDataManager for testing.
type ItemDataManager struct {
	mock.Mock
}

// ItemExists is a mock function.
func (m *ItemDataManager) ItemExists(ctx context.Context, itemID, userID uint64) (bool, error) {
	args := m.Called(ctx, itemID, userID)
	return args.Bool(0), args.Error(1)
}

// GetItem is a mock function.
func (m *ItemDataManager) GetItem(ctx context.Context, itemID, userID uint64) (*models.Item, error) {
	args := m.Called(ctx, itemID, userID)
	return args.Get(0).(*models.Item), args.Error(1)
}

// GetAllItemsCount is a mock function.
func (m *ItemDataManager) GetAllItemsCount(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

// GetAllItems is a mock function.
func (m *ItemDataManager) GetAllItems(ctx context.Context, results chan []models.Item) error {
	args := m.Called(ctx, results)
	return args.Error(0)
}

// GetItems is a mock function.
func (m *ItemDataManager) GetItems(ctx context.Context, userID uint64, filter *models.QueryFilter) (*models.ItemList, error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).(*models.ItemList), args.Error(1)
}

// GetItemsForAdmin is a mock function.
func (m *ItemDataManager) GetItemsForAdmin(ctx context.Context, filter *models.QueryFilter) (*models.ItemList, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*models.ItemList), args.Error(1)
}

// GetItemsWithIDs is a mock function.
func (m *ItemDataManager) GetItemsWithIDs(ctx context.Context, userID uint64, limit uint8, ids []uint64) ([]models.Item, error) {
	args := m.Called(ctx, userID, limit, ids)
	return args.Get(0).([]models.Item), args.Error(1)
}

// GetItemsWithIDsForAdmin is a mock function.
func (m *ItemDataManager) GetItemsWithIDsForAdmin(ctx context.Context, limit uint8, ids []uint64) ([]models.Item, error) {
	args := m.Called(ctx, limit, ids)
	return args.Get(0).([]models.Item), args.Error(1)
}

// CreateItem is a mock function.
func (m *ItemDataManager) CreateItem(ctx context.Context, input *models.ItemCreationInput) (*models.Item, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*models.Item), args.Error(1)
}

// UpdateItem is a mock function.
func (m *ItemDataManager) UpdateItem(ctx context.Context, updated *models.Item) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveItem is a mock function.
func (m *ItemDataManager) ArchiveItem(ctx context.Context, itemID, userID uint64) error {
	return m.Called(ctx, itemID, userID).Error(0)
}
