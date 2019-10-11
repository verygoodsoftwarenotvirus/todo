package database

import (
	"context"

	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"

	"github.com/stretchr/testify/mock"
)

var _ Database = (*MockDatabase)(nil)

// BuildMockDatabase builds a mock database
func BuildMockDatabase() *MockDatabase {
	return &MockDatabase{
		ItemDataManager:         &mockmodels.ItemDataManager{},
		UserDataManager:         &mockmodels.UserDataManager{},
		OAuth2ClientDataManager: &mockmodels.OAuth2ClientDataManager{},
		WebhookDataManager:      &mockmodels.WebhookDataManager{},
	}
}

// MockDatabase is our mock database structure
type MockDatabase struct {
	mock.Mock

	*mockmodels.ItemDataManager
	*mockmodels.UserDataManager
	*mockmodels.OAuth2ClientDataManager
	*mockmodels.WebhookDataManager
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
