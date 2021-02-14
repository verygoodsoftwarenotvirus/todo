package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions/bitmask"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.AccountUserMembershipDataManager = (*AccountUserMembershipDataManager)(nil)

// AccountUserMembershipDataManager is a mocked types.AccountUserMembershipDataManager for testing.
type AccountUserMembershipDataManager struct {
	mock.Mock
}

// GetMembershipsForUser satisfies our interface contract.
func (m *AccountUserMembershipDataManager) GetMembershipsForUser(ctx context.Context, userID uint64) (defaultAccount uint64, permissionsMap map[uint64]bitmask.ServiceUserPermissions, err error) {
	args := m.Called(ctx, userID)

	return args.Get(0).(uint64), args.Get(1).(map[uint64]bitmask.ServiceUserPermissions), args.Error(2)
}
