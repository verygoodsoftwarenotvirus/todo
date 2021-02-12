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

// MarkAccountAsUserDefault satisfies our interface contract.
func (m *AccountUserMembershipDataManager) MarkAccountAsUserDefault(ctx context.Context, userID, accountID, performedBy uint64) error {
	return m.Called(ctx, userID, accountID, performedBy).Error(0)
}

// UserIsMemberOfAccount satisfies our interface contract.
func (m *AccountUserMembershipDataManager) UserIsMemberOfAccount(ctx context.Context, userID, accountID, performedBy uint64) error {
	return m.Called(ctx, userID, accountID, performedBy).Error(0)
}

// AddUserToAccount satisfies our interface contract.
func (m *AccountUserMembershipDataManager) AddUserToAccount(ctx context.Context, userID, accountID, performedBy uint64) error {
	return m.Called(ctx, userID, accountID, performedBy).Error(0)
}

// RemoveUserFromAccount satisfies our interface contract.
func (m *AccountUserMembershipDataManager) RemoveUserFromAccount(ctx context.Context, userID, accountID, performedBy uint64) error {
	return m.Called(ctx, userID, accountID, performedBy).Error(0)
}

// GetAuditLogEntriesForAccountUserMembership is a mock function.
func (m *AccountUserMembershipDataManager) GetAuditLogEntriesForAccountUserMembership(ctx context.Context, itemID uint64) ([]*types.AuditLogEntry, error) {
	args := m.Called(ctx, itemID)
	return args.Get(0).([]*types.AuditLogEntry), args.Error(1)
}
