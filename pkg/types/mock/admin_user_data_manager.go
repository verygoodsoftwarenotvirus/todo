package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.AdminUserDataManager = (*AdminUserDataManager)(nil)

// AdminUserDataManager is a mocked types.AdminUserDataManager for testing.
type AdminUserDataManager struct {
	mock.Mock
}

// UpdateUserReputation is a mock function.
func (m *AdminUserDataManager) UpdateUserReputation(ctx context.Context, userID string, input *types.UserReputationUpdateInput) error {
	return m.Called(ctx, userID, input).Error(0)
}
