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
		AuditLogEntryDataManager:           &mocktypes.AuditLogEntryDataManager{},
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

	*mocktypes.AuditLogEntryDataManager
	*mocktypes.AccountSubscriptionPlanDataManager
	*mocktypes.ItemDataManager
	*mocktypes.UserDataManager
	*mocktypes.AdminUserDataManager
	*mocktypes.OAuth2ClientDataManager
	*mocktypes.WebhookDataManager
	*mocktypes.AccountDataManager
}

// Migrate satisfies the DataManager interface.
func (m *MockDatabase) Migrate(ctx context.Context, maxAttempts uint8, ucc *types.TestUserCreationConfig) error {
	return m.Called(ctx, maxAttempts, ucc).Error(0)
}

// IsReady satisfies the DataManager interface.
func (m *MockDatabase) IsReady(ctx context.Context, maxAttempts uint8) (ready bool) {
	return m.Called(ctx, maxAttempts).Bool(0)
}

// BeginTx satisfies the DataManager interface.
func (m *MockDatabase) BeginTx(ctx context.Context, options *sql.TxOptions) (*sql.Tx, error) {
	args := m.Called(ctx, options)
	return args.Get(0).(*sql.Tx), args.Error(1)
}

var _ SQLQueryBuilder = (*MockSQLQueryBuilder)(nil)

// BuildMockSQLQueryBuilder builds a MockSQLQueryBuilder.
func BuildMockSQLQueryBuilder() *MockSQLQueryBuilder {
	return &MockSQLQueryBuilder{
		AccountSQLQueryBuilder:                 &mocktypes.AccountSQLQueryBuilder{},
		AccountSubscriptionPlanSQLQueryBuilder: &mocktypes.AccountSubscriptionPlanSQLQueryBuilder{},
		AuditLogEntrySQLQueryBuilder:           &mocktypes.AuditLogEntrySQLQueryBuilder{},
		ItemSQLQueryBuilder:                    &mocktypes.ItemSQLQueryBuilder{},
		OAuth2ClientSQLQueryBuilder:            &mocktypes.OAuth2ClientSQLQueryBuilder{},
		UserSQLQueryBuilder:                    &mocktypes.UserSQLQueryBuilder{},
		WebhookSQLQueryBuilder:                 &mocktypes.WebhookSQLQueryBuilder{},
	}
}

// MockSQLQueryBuilder is our mock database structure.
type MockSQLQueryBuilder struct {
	mock.Mock

	*mocktypes.AccountSQLQueryBuilder
	*mocktypes.AccountSubscriptionPlanSQLQueryBuilder
	*mocktypes.AuditLogEntrySQLQueryBuilder
	*mocktypes.ItemSQLQueryBuilder
	*mocktypes.UserSQLQueryBuilder
	*mocktypes.OAuth2ClientSQLQueryBuilder
	*mocktypes.WebhookSQLQueryBuilder
}

// BuildMigrationFunc implements our interface.
func (m *MockSQLQueryBuilder) BuildMigrationFunc(db *sql.DB) func() {
	args := m.Called(db)

	return args.Get(0).(func())
}

// BuildTestUserCreationQuery implements our interface.
func (m *MockSQLQueryBuilder) BuildTestUserCreationQuery(testUserConfig *types.TestUserCreationConfig) (query string, args []interface{}) {
	returnValues := m.Called(testUserConfig)

	return returnValues.Get(0).(string), returnValues.Get(0).([]interface{})
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

// MockSQLResult mocks a sql.Result.
type MockSQLResult struct {
	mock.Mock
}

// LastInsertId implements our interface.
func (m *MockSQLResult) LastInsertId() (int64, error) {
	args := m.Called()

	return args.Get(0).(int64), args.Error(1)
}

// RowsAffected implements our interface.
func (m *MockSQLResult) RowsAffected() (int64, error) {
	args := m.Called()

	return args.Get(0).(int64), args.Error(1)
}
