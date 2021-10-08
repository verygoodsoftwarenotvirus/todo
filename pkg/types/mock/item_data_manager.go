package mocktypes

import (
	"context"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

var _ types.ItemDataManager = (*ItemDataManager)(nil)

// ItemDataManager is a mocked types.ItemDataManager for testing.
type ItemDataManager struct {
	mock.Mock
}

// ItemExists is a mock function.
func (m *ItemDataManager) ItemExists(ctx context.Context, itemID, accountID string) (bool, error) {
	args := m.Called(ctx, itemID, accountID)
	return args.Bool(0), args.Error(1)
}

// GetItem is a mock function.
func (m *ItemDataManager) GetItem(ctx context.Context, itemID, accountID string) (*types.Item, error) {
	args := m.Called(ctx, itemID, accountID)
	return args.Get(0).(*types.Item), args.Error(1)
}

// GetTotalItemCount is a mock function.
func (m *ItemDataManager) GetTotalItemCount(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

// GetItems is a mock function.
func (m *ItemDataManager) GetItems(ctx context.Context, accountID string, filter *types.QueryFilter) (*types.ItemList, error) {
	args := m.Called(ctx, accountID, filter)
	return args.Get(0).(*types.ItemList), args.Error(1)
}

// GetItemsWithIDs is a mock function.
func (m *ItemDataManager) GetItemsWithIDs(ctx context.Context, accountID string, limit uint8, ids []string) ([]*types.Item, error) {
	args := m.Called(ctx, accountID, limit, ids)
	return args.Get(0).([]*types.Item), args.Error(1)
}

// CreateItem is a mock function.
func (m *ItemDataManager) CreateItem(ctx context.Context, input *types.ItemDatabaseCreationInput) (*types.Item, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*types.Item), args.Error(1)
}

// UpdateItem is a mock function.
func (m *ItemDataManager) UpdateItem(ctx context.Context, updated *types.Item) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveItem is a mock function.
func (m *ItemDataManager) ArchiveItem(ctx context.Context, itemID, accountID string) error {
	return m.Called(ctx, itemID, accountID).Error(0)
}
