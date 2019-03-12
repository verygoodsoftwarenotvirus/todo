package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

var _ models.UserHandler = (*UserHandler)(nil)

// UserHandler is what it says on the tin
type UserHandler struct {
	mock.Mock
}

// GetUser is a mock function
func (m *UserHandler) GetUser(ctx context.Context, userID uint64) (*models.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*models.User), args.Error(1)
}

// GetUserByUsername is a mock function
func (m *UserHandler) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*models.User), args.Error(1)
}

// GetUserCount is a mock function
func (m *UserHandler) GetUserCount(ctx context.Context, filter *models.QueryFilter) (uint64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(uint64), args.Error(1)
}

// GetUsers is a mock function
func (m *UserHandler) GetUsers(ctx context.Context, filter *models.QueryFilter) (*models.UserList, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*models.UserList), args.Error(1)
}

// CreateUser is a mock function
func (m *UserHandler) CreateUser(ctx context.Context, input *models.UserInput) (*models.User, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*models.User), args.Error(1)
}

// UpdateUser is a mock function
func (m *UserHandler) UpdateUser(ctx context.Context, updated *models.User) error {
	return m.Called(ctx, updated).Error(0)
}

// DeleteUser is a mock function
func (m *UserHandler) DeleteUser(ctx context.Context, userID uint64) error {
	return m.Called(ctx, userID).Error(0)
}
