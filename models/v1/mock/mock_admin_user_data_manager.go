package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

var _ models.AdminUserDataManager = (*AdminUserDataManager)(nil)

// AdminUserDataManager is a mocked models.AdminUserDataManager for testing.
type AdminUserDataManager struct {
	mock.Mock
}

// MakeUserAdmin is a mock function.
func (m *AdminUserDataManager) MakeUserAdmin(ctx context.Context, userID uint64) error {
	return m.Called(ctx, userID).Error(0)
}
