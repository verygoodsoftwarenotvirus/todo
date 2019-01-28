package mock

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

var _ models.OAuth2ClientHandler = (*MockOauth2ClientHandler)(nil)

type MockOauth2ClientHandler struct {
	mock.Mock
}

func (m *MockOauth2ClientHandler) GetOAuth2Client(identifier string) (*models.OAuth2Client, error) {
	args := m.Called(identifier)
	return args.Get(0).(*models.OAuth2Client), args.Error(1)
}

func (m *MockOauth2ClientHandler) GetOAuth2ClientCount(filter *models.QueryFilter) (uint64, error) {
	args := m.Called(filter)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockOauth2ClientHandler) GetOAuth2Clients(filter *models.QueryFilter) (*models.Oauth2ClientList, error) {
	args := m.Called(filter)
	return args.Get(0).(*models.Oauth2ClientList), args.Error(1)
}

func (m *MockOauth2ClientHandler) CreateOAuth2Client(input *models.Oauth2ClientCreationInput) (*models.OAuth2Client, error) {
	args := m.Called(input)
	return args.Get(0).(*models.OAuth2Client), args.Error(1)
}

func (m *MockOauth2ClientHandler) UpdateOAuth2Client(updated *models.OAuth2Client) error {
	return m.Called(updated).Error(0)
}

func (m *MockOauth2ClientHandler) DeleteOAuth2Client(identifier string) error {
	return m.Called(identifier).Error(0)
}
