package database

import (
	"context"
	"database/sql"

	"github.com/alexedwards/scs/v2"
	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"
)

var _ DataManager = (*MockDatabase)(nil)

// BuildMockDatabase builds a mock database.
func BuildMockDatabase() *MockDatabase {
	return &MockDatabase{
		AccountDataManager:               &mocktypes.AccountDataManager{},
		AccountUserMembershipDataManager: &mocktypes.AccountUserMembershipDataManager{},
		ItemDataManager:                  &mocktypes.ItemDataManager{},
		UserDataManager:                  &mocktypes.UserDataManager{},
		AdminUserDataManager:             &mocktypes.AdminUserDataManager{},
		APIClientDataManager:             &mocktypes.APIClientDataManager{},
		WebhookDataManager:               &mocktypes.WebhookDataManager{},
	}
}

// MockDatabase is our mock database structure. Note, when using this in tests, you must directly access the type name of all the implicit fields.
// So `mockDB.On("GetUserByUsername"...)` is destined to fail, whereas `mockDB.UserDataManager.On("GetUserByUsername"...)` would do what you want it to do.
type MockDatabase struct {
	*mocktypes.AdminUserDataManager
	*mocktypes.AccountUserMembershipDataManager
	*mocktypes.ItemDataManager
	*mocktypes.UserDataManager
	*mocktypes.APIClientDataManager
	*mocktypes.WebhookDataManager
	*mocktypes.AccountDataManager
	mock.Mock
}

// ProvideSessionStore satisfies the DataManager interface.
func (m *MockDatabase) ProvideSessionStore() scs.Store {
	return m.Called().Get(0).(scs.Store)
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
