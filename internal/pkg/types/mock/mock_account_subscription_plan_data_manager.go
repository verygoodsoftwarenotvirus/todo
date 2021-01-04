package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.AccountSubscriptionPlanDataManager = (*AccountSubscriptionPlanDataManager)(nil)

// AccountSubscriptionPlanDataManager is a mocked types.AccountSubscriptionPlanDataManager for testing.
type AccountSubscriptionPlanDataManager struct {
	mock.Mock
}

// GetAccountSubscriptionPlan is a mock function.
func (m *AccountSubscriptionPlanDataManager) GetAccountSubscriptionPlan(ctx context.Context, planID uint64) (*types.AccountSubscriptionPlan, error) {
	args := m.Called(ctx, planID)
	return args.Get(0).(*types.AccountSubscriptionPlan), args.Error(1)
}

// GetAllAccountSubscriptionPlansCount is a mock function.
func (m *AccountSubscriptionPlanDataManager) GetAllAccountSubscriptionPlansCount(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

// GetAccountSubscriptionPlans is a mock function.
func (m *AccountSubscriptionPlanDataManager) GetAccountSubscriptionPlans(ctx context.Context, filter *types.QueryFilter) (*types.AccountSubscriptionPlanList, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*types.AccountSubscriptionPlanList), args.Error(1)
}

// CreateAccountSubscriptionPlan is a mock function.
func (m *AccountSubscriptionPlanDataManager) CreateAccountSubscriptionPlan(ctx context.Context, input *types.AccountSubscriptionPlanCreationInput) (*types.AccountSubscriptionPlan, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*types.AccountSubscriptionPlan), args.Error(1)
}

// UpdateAccountSubscriptionPlan is a mock function.
func (m *AccountSubscriptionPlanDataManager) UpdateAccountSubscriptionPlan(ctx context.Context, updated *types.AccountSubscriptionPlan) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveAccountSubscriptionPlan is a mock function.
func (m *AccountSubscriptionPlanDataManager) ArchiveAccountSubscriptionPlan(ctx context.Context, itemID uint64) error {
	return m.Called(ctx, itemID).Error(0)
}

// LogAccountSubscriptionPlanCreationEvent implements our interface.
func (m *AuditLogDataManager) LogAccountSubscriptionPlanCreationEvent(ctx context.Context, plan *types.AccountSubscriptionPlan) {
	m.Called(ctx, plan)
}

// AccountSubscriptionLogPlanUpdateEvent implements our interface.
func (m *AuditLogDataManager) AccountSubscriptionLogPlanUpdateEvent(ctx context.Context, userID, planID uint64, changes []types.FieldChangeSummary) {
	m.Called(ctx, userID, planID, changes)
}

// AccountSubscriptionLogPlanArchiveEvent implements our interface.
func (m *AuditLogDataManager) AccountSubscriptionLogPlanArchiveEvent(ctx context.Context, userID, planID uint64) {
	m.Called(ctx, userID, planID)
}

// GetAuditLogEntriesForAccountSubscriptionPlan is a mock function.
func (m *AuditLogDataManager) GetAuditLogEntriesForAccountSubscriptionPlan(ctx context.Context, planID uint64) ([]types.AuditLogEntry, error) {
	args := m.Called(ctx, planID)
	return args.Get(0).([]types.AuditLogEntry), args.Error(1)
}
