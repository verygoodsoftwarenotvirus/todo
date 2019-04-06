package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

var _ models.OAuth2ClientDataManager = (*OAuth2ClientHandler)(nil)

// OAuth2ClientDataManager is what it says on the tin
type OAuth2ClientHandler struct {
	mock.Mock
}

// GetOAuth2Client is a mock function
func (m *OAuth2ClientHandler) GetOAuth2Client(ctx context.Context, clientID, userID uint64) (*models.OAuth2Client, error) {
	args := m.Called(ctx, clientID)
	return args.Get(0).(*models.OAuth2Client), args.Error(1)
}

// GetOAuth2ClientByClientID is a mock function
func (m *OAuth2ClientHandler) GetOAuth2ClientByClientID(ctx context.Context, identifier string) (*models.OAuth2Client, error) {
	args := m.Called(ctx, identifier)
	return args.Get(0).(*models.OAuth2Client), args.Error(1)
}

// GetOAuth2ClientCount is a mock function
func (m *OAuth2ClientHandler) GetOAuth2ClientCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (uint64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(uint64), args.Error(1)
}

// GetAllOAuth2Clients is a mock function
func (m *OAuth2ClientHandler) GetAllOAuth2Clients(ctx context.Context) ([]models.OAuth2Client, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.OAuth2Client), args.Error(1)
}

// GetOAuth2Clients is a mock function
func (m *OAuth2ClientHandler) GetOAuth2Clients(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.OAuth2ClientList, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*models.OAuth2ClientList), args.Error(1)
}

// CreateOAuth2Client is a mock function
func (m *OAuth2ClientHandler) CreateOAuth2Client(ctx context.Context, input *models.OAuth2ClientCreationInput) (*models.OAuth2Client, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*models.OAuth2Client), args.Error(1)
}

// UpdateOAuth2Client is a mock function
func (m *OAuth2ClientHandler) UpdateOAuth2Client(ctx context.Context, updated *models.OAuth2Client) error {
	return m.Called(ctx, updated).Error(0)
}

// DeleteOAuth2Client is a mock function
func (m *OAuth2ClientHandler) DeleteOAuth2Client(ctx context.Context, clientID, userID uint64) error {
	return m.Called(ctx, clientID).Error(0)
}
