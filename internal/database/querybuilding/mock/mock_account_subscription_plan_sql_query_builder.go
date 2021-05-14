package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

var _ querybuilding.AccountSubscriptionPlanSQLQueryBuilder = (*AccountSubscriptionPlanSQLQueryBuilder)(nil)

// AccountSubscriptionPlanSQLQueryBuilder is a mocked types.AccountSubscriptionPlanSQLQueryBuilder for testing.
type AccountSubscriptionPlanSQLQueryBuilder struct {
	mock.Mock
}

// BuildGetAccountSubscriptionPlanQuery implements our interface.
func (m *AccountSubscriptionPlanSQLQueryBuilder) BuildGetAccountSubscriptionPlanQuery(ctx context.Context, accountSubscriptionPlanID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, accountSubscriptionPlanID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAllAccountSubscriptionPlansCountQuery implements our interface.
func (m *AccountSubscriptionPlanSQLQueryBuilder) BuildGetAllAccountSubscriptionPlansCountQuery(ctx context.Context) string {
	returnArgs := m.Called(ctx)

	return returnArgs.String(0)
}

// BuildGetAccountSubscriptionPlansQuery implements our interface.
func (m *AccountSubscriptionPlanSQLQueryBuilder) BuildGetAccountSubscriptionPlansQuery(ctx context.Context, filter *types.QueryFilter) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, filter)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildCreateAccountSubscriptionPlanQuery implements our interface.
func (m *AccountSubscriptionPlanSQLQueryBuilder) BuildCreateAccountSubscriptionPlanQuery(ctx context.Context, input *types.AccountSubscriptionPlanCreationInput) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildUpdateAccountSubscriptionPlanQuery implements our interface.
func (m *AccountSubscriptionPlanSQLQueryBuilder) BuildUpdateAccountSubscriptionPlanQuery(ctx context.Context, input *types.AccountSubscriptionPlan) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildArchiveAccountSubscriptionPlanQuery implements our interface.
func (m *AccountSubscriptionPlanSQLQueryBuilder) BuildArchiveAccountSubscriptionPlanQuery(ctx context.Context, accountSubscriptionPlanID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, accountSubscriptionPlanID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery implements our interface.
func (m *AccountSubscriptionPlanSQLQueryBuilder) BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery(ctx context.Context, accountSubscriptionPlanID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(ctx, accountSubscriptionPlanID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}
