package mock

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

var _ models.UserHandler = (*MockUserHandler)(nil)

type MockUserHandler struct {
	mock.Mock
}

func (m *MockUserHandler) GetUser(identifier string) (*models.User, error) {

	args := m.Called(identifier)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserHandler) GetUserCount(filter *models.QueryFilter) (uint64, error) {
	args := m.Called(filter)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockUserHandler) GetUsers(filter *models.QueryFilter) (*models.UserList, error) {
	args := m.Called(filter)
	return args.Get(0).(*models.UserList), args.Error(1)
}

func (m *MockUserHandler) CreateUser(input *models.UserInput, totpSecret string) (*models.User, error) {
	args := m.Called(input)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserHandler) UpdateUser(updated *models.User) error {
	return m.Called(updated).Error(0)
}

func (m *MockUserHandler) DeleteUser(id uint) error {
	return m.Called(id).Error(0)
}
