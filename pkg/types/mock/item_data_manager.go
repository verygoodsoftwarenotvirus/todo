package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/stretchr/testify/mock"
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

// GetAllItemsCount is a mock function.
func (m *ItemDataManager) GetAllItemsCount(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

// GetAllItems is a mock function.
func (m *ItemDataManager) GetAllItems(ctx context.Context, results chan []*types.Item, bucketSize uint16) error {
	args := m.Called(ctx, results, bucketSize)
	return args.Error(0)
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
func (m *ItemDataManager) CreateItem(ctx context.Context, input *types.ItemCreationInput, createdByUser string) (*types.Item, error) {
	args := m.Called(ctx, input, createdByUser)
	return args.Get(0).(*types.Item), args.Error(1)
}

// UpdateItem is a mock function.
func (m *ItemDataManager) UpdateItem(ctx context.Context, updated *types.Item, changedByUser string, changes []*types.FieldChangeSummary) error {
	return m.Called(ctx, updated, changedByUser, changes).Error(0)
}

// ArchiveItem is a mock function.
func (m *ItemDataManager) ArchiveItem(ctx context.Context, itemID, accountID, archivedBy string) error {
	return m.Called(ctx, itemID, accountID, archivedBy).Error(0)
}

// GetAuditLogEntriesForItem is a mock function.
func (m *ItemDataManager) GetAuditLogEntriesForItem(ctx context.Context, itemID string) ([]*types.AuditLogEntry, error) {
	args := m.Called(ctx, itemID)
	return args.Get(0).([]*types.AuditLogEntry), args.Error(1)
}
