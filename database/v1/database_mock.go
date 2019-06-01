package database

import (
	"context"
	mmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"

	"github.com/stretchr/testify/mock"
)

var _ Database = (*MockDatabase)(nil)

// BuildMockDatabase builds a mock database
func BuildMockDatabase() *MockDatabase {
	return &MockDatabase{
		ItemDataManager:         &mmodels.ItemDataManager{},
		UserDataManager:         &mmodels.UserDataManager{},
		OAuth2ClientDataManager: &mmodels.OAuth2ClientDataManager{},
		WebhookDataManager:      &mmodels.WebhookDataManager{},
	}
}

// MockDatabase is our mock database structure
type MockDatabase struct {
	mock.Mock

	*mmodels.ItemDataManager
	*mmodels.UserDataManager
	*mmodels.OAuth2ClientDataManager
	*mmodels.WebhookDataManager
}

// Migrate satisfies the database.Database interface
func (m *MockDatabase) Migrate(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// IsReady satisfies the database.Database interface
func (m *MockDatabase) IsReady(ctx context.Context) (ready bool) {
	args := m.Called(ctx)
	return args.Bool(0)
}
