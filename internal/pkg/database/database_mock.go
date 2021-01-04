package database

import (
	"context"
	"database/sql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/mock"
)

var _ DataManager = (*MockDatabase)(nil)

// BuildMockDatabase builds a mock database.
func BuildMockDatabase() *MockDatabase {
	return &MockDatabase{
		AuditLogDataManager:                &mocktypes.AuditLogDataManager{},
		AccountDataManager:                 &mocktypes.AccountDataManager{},
		AccountSubscriptionPlanDataManager: &mocktypes.AccountSubscriptionPlanDataManager{},
		ItemDataManager:                    &mocktypes.ItemDataManager{},
		UserDataManager:                    &mocktypes.UserDataManager{},
		AdminUserDataManager:               &mocktypes.AdminUserDataManager{},
		OAuth2ClientDataManager:            &mocktypes.OAuth2ClientDataManager{},
		WebhookDataManager:                 &mocktypes.WebhookDataManager{},
	}
}

// MockDatabase is our mock database structure.
type MockDatabase struct {
	mock.Mock

	*mocktypes.AuditLogDataManager
	*mocktypes.AccountSubscriptionPlanDataManager
	*mocktypes.ItemDataManager
	*mocktypes.UserDataManager
	*mocktypes.AdminUserDataManager
	*mocktypes.OAuth2ClientDataManager
	*mocktypes.WebhookDataManager
	*mocktypes.AccountDataManager
}

// Migrate satisfies the DataManager interface.
func (m *MockDatabase) Migrate(ctx context.Context, ucc *types.TestUserCreationConfig) error {
	return m.Called(ctx, ucc).Error(0)
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
