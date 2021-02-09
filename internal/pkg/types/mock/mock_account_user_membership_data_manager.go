package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.AccountUserMembershipDataManager = (*AccountUserMembershipDataManager)(nil)

// AccountUserMembershipDataManager is a mocked types.AccountUserMembershipDataManager for testing.
type AccountUserMembershipDataManager struct {
	mock.Mock
}

// GetAccountUserMembership satisfies our interface contract.
func (m *AccountUserMembershipDataManager) GetAccountUserMembership(ctx context.Context, itemID, userID uint64) (*types.AccountUserMembership, error) {
	args := m.Called(ctx, itemID, userID)
	return args.Get(0).(*types.AccountUserMembership), args.Error(1)
}

// GetAllAccountUserMembershipsCount satisfies our interface contract.
func (m *AccountUserMembershipDataManager) GetAllAccountUserMembershipsCount(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

// GetAllAccountUserMemberships satisfies our interface contract.
func (m *AccountUserMembershipDataManager) GetAllAccountUserMemberships(ctx context.Context, results chan []*types.AccountUserMembership, bucketSize uint16) error {
	args := m.Called(ctx, results, bucketSize)
	return args.Error(0)
}

// GetAccountUserMemberships satisfies our interface contract.
func (m *AccountUserMembershipDataManager) GetAccountUserMemberships(ctx context.Context, userID uint64, filter *types.QueryFilter) (*types.AccountUserMembershipList, error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).(*types.AccountUserMembershipList), args.Error(1)
}

// GetAccountUserMembershipsForAdmin satisfies our interface contract.
func (m *AccountUserMembershipDataManager) GetAccountUserMembershipsForAdmin(ctx context.Context, filter *types.QueryFilter) (*types.AccountUserMembershipList, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*types.AccountUserMembershipList), args.Error(1)
}

// CreateAccountUserMembership satisfies our interface contract.
func (m *AccountUserMembershipDataManager) CreateAccountUserMembership(ctx context.Context, input *types.AccountUserMembershipCreationInput) (*types.AccountUserMembership, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*types.AccountUserMembership), args.Error(1)
}

// ArchiveAccountUserMembership satisfies our interface contract.
func (m *AccountUserMembershipDataManager) ArchiveAccountUserMembership(ctx context.Context, itemID, userID uint64) error {
	return m.Called(ctx, itemID, userID).Error(0)
}

// GetAuditLogEntriesForAccountUserMembership is a mock function.
func (m *AccountUserMembershipDataManager) GetAuditLogEntriesForAccountUserMembership(ctx context.Context, itemID uint64) ([]*types.AuditLogEntry, error) {
	args := m.Called(ctx, itemID)
	return args.Get(0).([]*types.AuditLogEntry), args.Error(1)
}
