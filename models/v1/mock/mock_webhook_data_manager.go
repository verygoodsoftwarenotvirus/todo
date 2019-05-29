package mock

import (
	"context"
	"github.com/stretchr/testify/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

var _ models.WebhookDataManager = (*WebhookDataManager)(nil)

type WebhookDataManager struct {
	mock.Mock
}

func (m *WebhookDataManager) GetWebhook(ctx context.Context, itemID, userID uint64) (*models.Webhook, error) {
	args := m.Called(ctx, itemID, userID)
	return args.Get(0).(*models.Webhook), args.Error(1)
}

func (m *WebhookDataManager) GetWebhookCount(ctx context.Context, filter *models.QueryFilter, userID uint64) (uint64, error) {
	args := m.Called(ctx, filter, userID)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *WebhookDataManager) GetAllWebhooksCount(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *WebhookDataManager) GetWebhooks(ctx context.Context, filter *models.QueryFilter, userID uint64) (*models.WebhookList, error) {
	args := m.Called(ctx, filter, userID)
	return args.Get(0).(*models.WebhookList), args.Error(1)
}

func (m *WebhookDataManager) GetAllWebhooks(ctx context.Context) (*models.WebhookList, error) {
	args := m.Called(ctx)
	return args.Get(0).(*models.WebhookList), args.Error(1)
}

func (m *WebhookDataManager) CreateWebhook(ctx context.Context, input *models.WebhookInput) (*models.Webhook, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*models.Webhook), args.Error(1)
}

func (m *WebhookDataManager) UpdateWebhook(ctx context.Context, updated *models.Webhook) error {
	return m.Called(ctx, updated).Error(0)
}

func (m *WebhookDataManager) DeleteWebhook(ctx context.Context, id uint64, userID uint64) error {
	return m.Called(ctx, id, userID).Error(0)
}
