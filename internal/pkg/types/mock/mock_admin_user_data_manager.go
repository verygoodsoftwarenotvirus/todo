package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.AdminUserDataManager = (*AdminUserDataManager)(nil)

// AdminUserDataManager is a mocked types.AdminUserDataManager for testing.
type AdminUserDataManager struct {
	mock.Mock
}

// BanUser is a mock function.
func (m *AdminUserDataManager) BanUser(ctx context.Context, userID uint64) error {
	return m.Called(ctx, userID).Error(0)
}
