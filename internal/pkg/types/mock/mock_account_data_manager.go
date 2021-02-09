package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.AccountDataManager = (*AccountDataManager)(nil)

// AccountDataManager is a mocked types.AccountDataManager for testing.
type AccountDataManager struct {
	mock.Mock
}

// AccountExists is a mock function.
func (m *AccountDataManager) AccountExists(ctx context.Context, accountID, userID uint64) (bool, error) {
	args := m.Called(ctx, accountID, userID)
	return args.Bool(0), args.Error(1)
}

// GetAccount is a mock function.
func (m *AccountDataManager) GetAccount(ctx context.Context, accountID, userID uint64) (*types.Account, error) {
	args := m.Called(ctx, accountID, userID)
	return args.Get(0).(*types.Account), args.Error(1)
}

// GetAllAccountsCount is a mock function.
func (m *AccountDataManager) GetAllAccountsCount(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

// GetAllAccounts is a mock function.
func (m *AccountDataManager) GetAllAccounts(ctx context.Context, results chan []*types.Account, bucketSize uint16) error {
	args := m.Called(ctx, results, bucketSize)
	return args.Error(0)
}

// GetAccounts is a mock function.
func (m *AccountDataManager) GetAccounts(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.AccountList, error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).(*types.AccountList), args.Error(1)
}

// GetAccountsForAdmin is a mock function.
func (m *AccountDataManager) GetAccountsForAdmin(ctx context.Context, filter *types.QueryFilter) (*types.AccountList, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*types.AccountList), args.Error(1)
}

// CreateAccount is a mock function.
func (m *AccountDataManager) CreateAccount(ctx context.Context, input *types.AccountCreationInput) (*types.Account, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*types.Account), args.Error(1)
}

// UpdateAccount is a mock function.
func (m *AccountDataManager) UpdateAccount(ctx context.Context, updated *types.Account, changes []types.FieldChangeSummary) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveAccount is a mock function.
func (m *AccountDataManager) ArchiveAccount(ctx context.Context, accountID, userID uint64) error {
	return m.Called(ctx, accountID, userID).Error(0)
}

// GetAuditLogEntriesForAccount is a mock function.
func (m *AccountDataManager) GetAuditLogEntriesForAccount(ctx context.Context, accountID uint64) ([]*types.AuditLogEntry, error) {
	args := m.Called(ctx, accountID)
	return args.Get(0).([]*types.AuditLogEntry), args.Error(1)
}
