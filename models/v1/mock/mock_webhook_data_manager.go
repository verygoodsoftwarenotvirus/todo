package mock

import (
	"context"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

var _ models.WebhookDataManager = (*WebhookDataManager)(nil)

// WebhookDataManager is a mocked models.WebhookDataManager for testing
type WebhookDataManager struct {
	mock.Mock
}

// GetWebhook satisfies our WebhookDataManager interface
func (m *WebhookDataManager) GetWebhook(ctx context.Context, webhookID, userID uint64) (*models.Webhook, error) {
	args := m.Called(ctx, webhookID, userID)
	return args.Get(0).(*models.Webhook), args.Error(1)
}

// GetWebhookCount satisfies our WebhookDataManager interface
func (m *WebhookDataManager) GetWebhookCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (uint64, error) {
	args := m.Called(ctx, filter, userID)
	return args.Get(0).(uint64), args.Error(1)
}

// GetAllWebhooksCount satisfies our WebhookDataManager interface
func (m *WebhookDataManager) GetAllWebhooksCount(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

// GetWebhooks satisfies our WebhookDataManager interface
func (m *WebhookDataManager) GetWebhooks(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.WebhookList, error) {
	args := m.Called(ctx, filter, userID)
	return args.Get(0).(*models.WebhookList), args.Error(1)
}

// GetAllWebhooks satisfies our WebhookDataManager interface
func (m *WebhookDataManager) GetAllWebhooks(ctx context.Context) (*models.WebhookList, error) {
	args := m.Called(ctx)
	return args.Get(0).(*models.WebhookList), args.Error(1)
}

// GetAllWebhooksForUser satisfies our WebhookDataManager interface
func (m *WebhookDataManager) GetAllWebhooksForUser(ctx context.Context, userID uint64) ([]models.Webhook, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Webhook), args.Error(1)
}

// CreateWebhook satisfies our WebhookDataManager interface
func (m *WebhookDataManager) CreateWebhook(ctx context.Context, input *models.WebhookCreationInput) (*models.Webhook, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*models.Webhook), args.Error(1)
}

// UpdateWebhook satisfies our WebhookDataManager interface
func (m *WebhookDataManager) UpdateWebhook(ctx context.Context, updated *models.Webhook) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveWebhook satisfies our WebhookDataManager interface
func (m *WebhookDataManager) ArchiveWebhook(ctx context.Context, webhookID, userID uint64) error {
	return m.Called(ctx, webhookID, userID).Error(0)
}
