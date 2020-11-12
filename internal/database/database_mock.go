package database

import (
	"context"
	"database/sql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/auth"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/mock"

	"github.com/stretchr/testify/mock"
)

var _ DataManager = (*MockDatabase)(nil)

// BuildMockDatabase builds a mock database.
func BuildMockDatabase() *MockDatabase {
	return &MockDatabase{
		AuditLogDataManager:     &mockmodels.AuditLogDataManager{},
		ItemDataManager:         &mockmodels.ItemDataManager{},
		UserDataManager:         &mockmodels.UserDataManager{},
		AdminUserDataManager:    &mockmodels.AdminUserDataManager{},
		OAuth2ClientDataManager: &mockmodels.OAuth2ClientDataManager{},
		WebhookDataManager:      &mockmodels.WebhookDataManager{},
	}
}

// MockDatabase is our mock database structure.
type MockDatabase struct {
	mock.Mock

	*mockmodels.AuditLogDataManager
	*mockmodels.ItemDataManager
	*mockmodels.UserDataManager
	*mockmodels.AdminUserDataManager
	*mockmodels.OAuth2ClientDataManager
	*mockmodels.WebhookDataManager
}

// Migrate satisfies the DataManager interface.
func (m *MockDatabase) Migrate(ctx context.Context, authenticator auth.Authenticator, ucc *UserCreationConfig) error {
	return m.Called(ctx, authenticator, ucc).Error(0)
}

// IsReady satisfies the DataManager interface.
func (m *MockDatabase) IsReady(ctx context.Context) (ready bool) {
	return m.Called(ctx).Bool(0)
}

// BeginTx satisfies the DataManager interface.
func (m *MockDatabase) BeginTx(ctx context.Context, options *sql.TxOptions) (*sql.Tx, error) {
	args := m.Called(ctx, options)
	return args.Get(0).(*sql.Tx), args.Error(1)
}

var _ ResultIterator = (*MockResultIterator)(nil)

// MockResultIterator is our mock sql.Rows structure.
type MockResultIterator struct {
	mock.Mock
}

// Scan satisfies the ResultIterator interface.
func (m *MockResultIterator) Scan(dest ...interface{}) error {
	return m.Called(dest...).Error(0)
}

// Next satisfies the ResultIterator interface.
func (m *MockResultIterator) Next() bool {
	return m.Called().Bool(0)
}

// Err satisfies the ResultIterator interface.
func (m *MockResultIterator) Err() error {
	return m.Called().Error(0)
}

// Close satisfies the ResultIterator interface.
func (m *MockResultIterator) Close() error {
	return m.Called().Error(0)
}
