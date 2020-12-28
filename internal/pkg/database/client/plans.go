package dbclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.PlanDataManager = (*Client)(nil)

// GetPlan fetches an plan from the database.
func (c *Client) GetPlan(ctx context.Context, planID uint64) (*types.Plan, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachPlanIDToSpan(span, planID)

	c.logger.WithValues(map[string]interface{}{
		"plan_id": planID,
	}).Debug("GetPlan called")

	return c.querier.GetPlan(ctx, planID)
}

// GetAllPlansCount fetches the count of plans from the database that meet a particular filter.
func (c *Client) GetAllPlansCount(ctx context.Context) (count uint64, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllPlansCount called")

	return c.querier.GetAllPlansCount(ctx)
}

// GetPlans fetches a list of plans from the database that meet a particular filter.
func (c *Client) GetPlans(ctx context.Context, filter *types.QueryFilter) (*types.PlanList, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit)
	}

	c.logger.Debug("GetPlans called")

	return c.querier.GetPlans(ctx, filter)
}

// CreatePlan creates an plan in the database.
func (c *Client) CreatePlan(ctx context.Context, input *types.PlanCreationInput) (*types.Plan, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("CreatePlan called")

	return c.querier.CreatePlan(ctx, input)
}

// UpdatePlan updates a particular plan. Note that UpdatePlan expects the
// provided input to have a valid ID.
func (c *Client) UpdatePlan(ctx context.Context, updated *types.Plan) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachPlanIDToSpan(span, updated.ID)
	c.logger.WithValue(keys.PlanIDKey, updated.ID).Debug("UpdatePlan called")

	return c.querier.UpdatePlan(ctx, updated)
}

// ArchivePlan archives an plan from the database by its ID.
func (c *Client) ArchivePlan(ctx context.Context, planID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachPlanIDToSpan(span, planID)

	c.logger.WithValues(map[string]interface{}{
		"plan_id": planID,
	}).Debug("ArchivePlan called")

	return c.querier.ArchivePlan(ctx, planID)
}

// LogPlanCreationEvent implements our AuditLogDataManager interface.
func (c *Client) LogPlanCreationEvent(ctx context.Context, plan *types.Plan) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("LogPlanCreationEvent called")

	c.querier.LogPlanCreationEvent(ctx, plan)
}

// LogPlanUpdateEvent implements our AuditLogDataManager interface.
func (c *Client) LogPlanUpdateEvent(ctx context.Context, userID, planID uint64, changes []types.FieldChangeSummary) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogPlanUpdateEvent called")

	c.querier.LogPlanUpdateEvent(ctx, userID, planID, changes)
}

// LogPlanArchiveEvent implements our AuditLogDataManager interface.
func (c *Client) LogPlanArchiveEvent(ctx context.Context, userID, planID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogPlanArchiveEvent called")

	c.querier.LogItemArchiveEvent(ctx, userID, planID)
}

// GetAuditLogEntriesForPlan fetches a list of audit log entries from the database that relate to a given plan.
func (c *Client) GetAuditLogEntriesForPlan(ctx context.Context, planID uint64) ([]types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAuditLogEntriesForPlan called")

	return c.querier.GetAuditLogEntriesForPlan(ctx, planID)
}
