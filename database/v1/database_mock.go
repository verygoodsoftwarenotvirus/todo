package database

import (
	"context"
	mmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"

	"github.com/stretchr/testify/mock"
)

var _ Database = (*MockDatabase)(nil)

func BuildMockDatabase() *MockDatabase {
	return &MockDatabase{
		ItemDataManager:         &mmodels.ItemDataManager{},
		UserDataManager:         &mmodels.UserDataManager{},
		OAuth2ClientDataManager: &mmodels.OAuth2ClientDataManager{},
		WebhookDataManager:      &mmodels.WebhookDataManager{},
	}
}

type MockDatabase struct {
	mock.Mock

	*mmodels.ItemDataManager
	*mmodels.UserDataManager
	*mmodels.OAuth2ClientDataManager
	*mmodels.WebhookDataManager
}

func (m *MockDatabase) Migrate(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDatabase) IsReady(ctx context.Context) (ready bool) {
	args := m.Called(ctx)
	return args.Bool(0)
}

func (m *MockDatabase) AdminUserExists(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}
