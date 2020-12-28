package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.PlanDataManager = (*PlanDataManager)(nil)

// PlanDataManager is a mocked types.PlanDataManager for testing.
type PlanDataManager struct {
	mock.Mock
}

// GetPlan is a mock function.
func (m *PlanDataManager) GetPlan(ctx context.Context, planID uint64) (*types.Plan, error) {
	args := m.Called(ctx, planID)
	return args.Get(0).(*types.Plan), args.Error(1)
}

// GetAllPlansCount is a mock function.
func (m *PlanDataManager) GetAllPlansCount(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

// GetPlans is a mock function.
func (m *PlanDataManager) GetPlans(ctx context.Context, filter *types.QueryFilter) (*types.PlanList, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*types.PlanList), args.Error(1)
}

// CreatePlan is a mock function.
func (m *PlanDataManager) CreatePlan(ctx context.Context, input *types.PlanCreationInput) (*types.Plan, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*types.Plan), args.Error(1)
}

// UpdatePlan is a mock function.
func (m *PlanDataManager) UpdatePlan(ctx context.Context, updated *types.Plan) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchivePlan is a mock function.
func (m *PlanDataManager) ArchivePlan(ctx context.Context, itemID uint64) error {
	return m.Called(ctx, itemID).Error(0)
}

// LogPlanCreationEvent implements our interface.
func (m *AuditLogDataManager) LogPlanCreationEvent(ctx context.Context, plan *types.Plan) {
	m.Called(ctx, plan)
}

// LogPlanUpdateEvent implements our interface.
func (m *AuditLogDataManager) LogPlanUpdateEvent(ctx context.Context, userID, planID uint64, changes []types.FieldChangeSummary) {
	m.Called(ctx, userID, planID, changes)
}

// LogPlanArchiveEvent implements our interface.
func (m *AuditLogDataManager) LogPlanArchiveEvent(ctx context.Context, userID, planID uint64) {
	m.Called(ctx, userID, planID)
}

// GetAuditLogEntriesForPlan is a mock function.
func (m *AuditLogDataManager) GetAuditLogEntriesForPlan(ctx context.Context, planID uint64) ([]types.AuditLogEntry, error) {
	args := m.Called(ctx, planID)
	return args.Get(0).([]types.AuditLogEntry), args.Error(1)
}
