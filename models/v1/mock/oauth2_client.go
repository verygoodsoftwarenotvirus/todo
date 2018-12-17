package mock

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

var _ models.Oauth2ClientHandler = (*MockOauth2ClientHandler)(nil)

type MockOauth2ClientHandler struct {
	mock.Mock
}

func (m *MockOauth2ClientHandler) GetOauth2Client(identifier string) (*models.Oauth2Client, error) {
	args := m.Called(identifier)
	return args.Get(0).(*models.Oauth2Client), args.Error(1)
}

func (m *MockOauth2ClientHandler) GetOauth2ClientCount(filter *models.QueryFilter) (uint64, error) {
	args := m.Called(filter)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockOauth2ClientHandler) GetOauth2Clients(filter *models.QueryFilter) (*models.Oauth2ClientList, error) {
	args := m.Called(filter)
	return args.Get(0).(*models.Oauth2ClientList), args.Error(1)
}

func (m *MockOauth2ClientHandler) CreateOauth2Client(input *models.Oauth2ClientInput) (*models.Oauth2Client, error) {
	args := m.Called(input)
	return args.Get(0).(*models.Oauth2Client), args.Error(1)
}

func (m *MockOauth2ClientHandler) UpdateOauth2Client(updated *models.Oauth2Client) error {
	return m.Called(updated).Error(0)
}

func (m *MockOauth2ClientHandler) DeleteOauth2Client(id uint) error {
	return m.Called(id).Error(0)
}
