package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.AccountUserMembershipDataManager = (*AccountUserMembershipDataManager)(nil)

// AccountUserMembershipDataManager is a mocked types.AccountUserMembershipDataManager for testing.
type AccountUserMembershipDataManager struct {
	mock.Mock
}

// BuildSessionContextDataForUser satisfies our interface contract.
func (m *AccountUserMembershipDataManager) BuildSessionContextDataForUser(ctx context.Context, userID uint64) (*types.SessionContextData, error) {
	returnValues := m.Called(ctx, userID)

	return returnValues.Get(0).(*types.SessionContextData), returnValues.Error(1)
}

// GetDefaultAccountIDForUser satisfies our interface contract.
func (m *AccountUserMembershipDataManager) GetDefaultAccountIDForUser(ctx context.Context, userID uint64) (uint64, error) {
	returnValues := m.Called(ctx, userID)

	return returnValues.Get(0).(uint64), returnValues.Error(1)
}

// GetMembershipsForUser satisfies our interface contract.
func (m *AccountUserMembershipDataManager) GetMembershipsForUser(ctx context.Context, userID uint64) (defaultAccount uint64, permissionsMap map[uint64]permissions.ServiceUserPermission, err error) {
	args := m.Called(ctx, userID)

	return args.Get(0).(uint64), args.Get(1).(map[uint64]permissions.ServiceUserPermission), args.Error(2)
}

// MarkAccountAsUserDefault implements the interface.
func (m *AccountUserMembershipDataManager) MarkAccountAsUserDefault(ctx context.Context, userID, accountID, changedByUser uint64) error {
	return m.Called(ctx, userID, accountID, changedByUser).Error(0)
}

// UserIsMemberOfAccount implements the interface.
func (m *AccountUserMembershipDataManager) UserIsMemberOfAccount(ctx context.Context, userID, accountID uint64) (bool, error) {
	returnValues := m.Called(ctx, userID, accountID)

	return returnValues.Bool(0), returnValues.Error(1)
}

// AddUserToAccount implements the interface.
func (m *AccountUserMembershipDataManager) AddUserToAccount(ctx context.Context, input *types.AddUserToAccountInput, addedByUser uint64) error {
	return m.Called(ctx, input, addedByUser).Error(0)
}

// RemoveUserFromAccount implements the interface.
func (m *AccountUserMembershipDataManager) RemoveUserFromAccount(ctx context.Context, userID, accountID, removedByUser uint64, reason string) error {
	return m.Called(ctx, userID, accountID, removedByUser, reason).Error(0)
}

// ModifyUserPermissions implements the interface.
func (m *AccountUserMembershipDataManager) ModifyUserPermissions(ctx context.Context, userID, accountID, changedByUser uint64, input *types.ModifyUserPermissionsInput) error {
	return m.Called(ctx, userID, accountID, changedByUser, input).Error(0)
}

// TransferAccountOwnership implements the interface.
func (m *AccountUserMembershipDataManager) TransferAccountOwnership(ctx context.Context, accountID, transferredBy uint64, input *types.AccountOwnershipTransferInput) error {
	return m.Called(ctx, accountID, transferredBy, input).Error(0)
}
