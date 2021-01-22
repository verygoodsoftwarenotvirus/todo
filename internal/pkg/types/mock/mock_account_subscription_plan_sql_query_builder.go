package mock

import (
	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.AccountSubscriptionPlanSQLQueryBuilder = (*AccountSubscriptionPlanSQLQueryBuilder)(nil)

// AccountSubscriptionPlanSQLQueryBuilder is a mocked types.AccountSubscriptionPlanSQLQueryBuilder for testing.
type AccountSubscriptionPlanSQLQueryBuilder struct {
	mock.Mock
}

// BuildGetAccountSubscriptionPlanQuery implements our interface.
func (m *AccountSubscriptionPlanSQLQueryBuilder) BuildGetAccountSubscriptionPlanQuery(planID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(planID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAllAccountSubscriptionPlansCountQuery implements our interface.
func (m *AccountSubscriptionPlanSQLQueryBuilder) BuildGetAllAccountSubscriptionPlansCountQuery() string {
	returnArgs := m.Called()

	return returnArgs.String(0)
}

// BuildGetAccountSubscriptionPlansQuery implements our interface.
func (m *AccountSubscriptionPlanSQLQueryBuilder) BuildGetAccountSubscriptionPlansQuery(filter *types.QueryFilter) (query string, args []interface{}) {
	returnArgs := m.Called(filter)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildCreateAccountSubscriptionPlanQuery implements our interface.
func (m *AccountSubscriptionPlanSQLQueryBuilder) BuildCreateAccountSubscriptionPlanQuery(input *types.AccountSubscriptionPlan) (query string, args []interface{}) {
	returnArgs := m.Called(input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildUpdateAccountSubscriptionPlanQuery implements our interface.
func (m *AccountSubscriptionPlanSQLQueryBuilder) BuildUpdateAccountSubscriptionPlanQuery(input *types.AccountSubscriptionPlan) (query string, args []interface{}) {
	returnArgs := m.Called(input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildArchiveAccountSubscriptionPlanQuery implements our interface.
func (m *AccountSubscriptionPlanSQLQueryBuilder) BuildArchiveAccountSubscriptionPlanQuery(planID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(planID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery implements our interface.
func (m *AccountSubscriptionPlanSQLQueryBuilder) BuildGetAuditLogEntriesForAccountSubscriptionPlanQuery(planID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(planID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}
