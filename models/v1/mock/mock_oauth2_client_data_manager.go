package mock

import (
	"context"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

var _ models.OAuth2ClientDataManager = (*OAuth2ClientDataManager)(nil)

// OAuth2ClientDataManager is a mocked models.OAuth2ClientDataManager for testing
type OAuth2ClientDataManager struct {
	mock.Mock
}

// GetOAuth2Client is a mock function
func (m *OAuth2ClientDataManager) GetOAuth2Client(ctx context.Context, clientID, userID uint64) (*models.OAuth2Client, error) {
	args := m.Called(ctx, clientID, userID)
	return args.Get(0).(*models.OAuth2Client), args.Error(1)
}

// GetOAuth2ClientByClientID is a mock function
func (m *OAuth2ClientDataManager) GetOAuth2ClientByClientID(ctx context.Context, identifier string) (*models.OAuth2Client, error) {
	args := m.Called(ctx, identifier)
	return args.Get(0).(*models.OAuth2Client), args.Error(1)
}

// GetOAuth2ClientCount is a mock function
func (m *OAuth2ClientDataManager) GetOAuth2ClientCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (uint64, error) {
	args := m.Called(ctx, filter, userID)
	return args.Get(0).(uint64), args.Error(1)
}

// GetAllOAuth2ClientCount is a mock function
func (m *OAuth2ClientDataManager) GetAllOAuth2ClientCount(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

// GetAllOAuth2Clients is a mock function
func (m *OAuth2ClientDataManager) GetAllOAuth2Clients(ctx context.Context) ([]*models.OAuth2Client, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.OAuth2Client), args.Error(1)
}

// GetAllOAuth2ClientsForUser is a mock function
func (m *OAuth2ClientDataManager) GetAllOAuth2ClientsForUser(ctx context.Context, userID uint64) ([]*models.OAuth2Client, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.OAuth2Client), args.Error(1)
}

// GetOAuth2Clients is a mock function
func (m *OAuth2ClientDataManager) GetOAuth2Clients(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.OAuth2ClientList, error) {
	args := m.Called(ctx, filter, userID)
	return args.Get(0).(*models.OAuth2ClientList), args.Error(1)
}

// CreateOAuth2Client is a mock function
func (m *OAuth2ClientDataManager) CreateOAuth2Client(ctx context.Context, input *models.OAuth2ClientCreationInput) (*models.OAuth2Client, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*models.OAuth2Client), args.Error(1)
}

// UpdateOAuth2Client is a mock function
func (m *OAuth2ClientDataManager) UpdateOAuth2Client(ctx context.Context, updated *models.OAuth2Client) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveOAuth2Client is a mock function
func (m *OAuth2ClientDataManager) ArchiveOAuth2Client(ctx context.Context, clientID, userID uint64) error {
	return m.Called(ctx, clientID, userID).Error(0)
}
