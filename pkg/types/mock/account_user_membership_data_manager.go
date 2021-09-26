package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

var _ types.AccountUserMembershipDataManager = (*AccountUserMembershipDataManager)(nil)

// AccountUserMembershipDataManager is a mocked types.AccountUserMembershipDataManager for testing.
type AccountUserMembershipDataManager struct {
	mock.Mock
}

// BuildSessionContextDataForUser satisfies our interface contract.
func (m *AccountUserMembershipDataManager) BuildSessionContextDataForUser(ctx context.Context, userID string) (*types.SessionContextData, error) {
	returnValues := m.Called(ctx, userID)

	return returnValues.Get(0).(*types.SessionContextData), returnValues.Error(1)
}

// GetDefaultAccountIDForUser satisfies our interface contract.
func (m *AccountUserMembershipDataManager) GetDefaultAccountIDForUser(ctx context.Context, userID string) (string, error) {
	returnValues := m.Called(ctx, userID)

	return returnValues.Get(0).(string), returnValues.Error(1)
}

// MarkAccountAsUserDefault implements the interface.
func (m *AccountUserMembershipDataManager) MarkAccountAsUserDefault(ctx context.Context, userID, accountID string) error {
	return m.Called(ctx, userID, accountID).Error(0)
}

// UserIsMemberOfAccount implements the interface.
func (m *AccountUserMembershipDataManager) UserIsMemberOfAccount(ctx context.Context, userID, accountID string) (bool, error) {
	returnValues := m.Called(ctx, userID, accountID)

	return returnValues.Bool(0), returnValues.Error(1)
}

// AddUserToAccount implements the interface.
func (m *AccountUserMembershipDataManager) AddUserToAccount(ctx context.Context, input *types.AddUserToAccountInput) error {
	return m.Called(ctx, input).Error(0)
}

// RemoveUserFromAccount implements the interface.
func (m *AccountUserMembershipDataManager) RemoveUserFromAccount(ctx context.Context, userID, accountID string) error {
	return m.Called(ctx, userID, accountID).Error(0)
}

// ModifyUserPermissions implements the interface.
func (m *AccountUserMembershipDataManager) ModifyUserPermissions(ctx context.Context, accountID, userID string, input *types.ModifyUserPermissionsInput) error {
	return m.Called(ctx, userID, accountID, input).Error(0)
}

// TransferAccountOwnership implements the interface.
func (m *AccountUserMembershipDataManager) TransferAccountOwnership(ctx context.Context, accountID string, input *types.AccountOwnershipTransferInput) error {
	return m.Called(ctx, accountID, input).Error(0)
}
