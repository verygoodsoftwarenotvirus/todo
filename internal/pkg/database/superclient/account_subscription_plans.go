package superclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.AccountSubscriptionPlanDataManager = (*Client)(nil)

// GetAccountSubscriptionPlan fetches an plan from the database.
func (c *Client) GetAccountSubscriptionPlan(ctx context.Context, planID uint64) (*types.AccountSubscriptionPlan, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachPlanIDToSpan(span, planID)

	c.logger.WithValues(map[string]interface{}{
		"plan_id": planID,
	}).Debug("GetAccountSubscriptionPlan called")

	return c.querier.GetAccountSubscriptionPlan(ctx, planID)
}

// GetAllAccountSubscriptionPlansCount fetches the count of plans from the database that meet a particular filter.
func (c *Client) GetAllAccountSubscriptionPlansCount(ctx context.Context) (count uint64, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("GetAllAccountSubscriptionPlansCount called")

	return c.querier.GetAllAccountSubscriptionPlansCount(ctx)
}

// GetAccountSubscriptionPlans fetches a list of plans from the database that meet a particular filter.
func (c *Client) GetAccountSubscriptionPlans(ctx context.Context, filter *types.QueryFilter) (*types.AccountSubscriptionPlanList, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if filter != nil {
		tracing.AttachFilterToSpan(span, filter.Page, filter.Limit)
	}

	c.logger.Debug("GetAccountSubscriptionPlans called")

	return c.querier.GetAccountSubscriptionPlans(ctx, filter)
}

// CreateAccountSubscriptionPlan creates an plan in the database.
func (c *Client) CreateAccountSubscriptionPlan(ctx context.Context, input *types.AccountSubscriptionPlanCreationInput) (*types.AccountSubscriptionPlan, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("CreateAccountSubscriptionPlan called")

	return c.querier.CreateAccountSubscriptionPlan(ctx, input)
}

// UpdateAccountSubscriptionPlan updates a particular plan. Note that UpdatePlan expects the
// provided input to have a valid ID.
func (c *Client) UpdateAccountSubscriptionPlan(ctx context.Context, updated *types.AccountSubscriptionPlan) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachPlanIDToSpan(span, updated.ID)
	c.logger.WithValue(keys.PlanIDKey, updated.ID).Debug("UpdateAccountSubscriptionPlan called")

	return c.querier.UpdateAccountSubscriptionPlan(ctx, updated)
}

// ArchiveAccountSubscriptionPlan archives an plan from the database by its ID.
func (c *Client) ArchiveAccountSubscriptionPlan(ctx context.Context, planID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachPlanIDToSpan(span, planID)

	c.logger.WithValues(map[string]interface{}{
		"plan_id": planID,
	}).Debug("ArchiveAccountSubscriptionPlan called")

	return c.querier.ArchiveAccountSubscriptionPlan(ctx, planID)
}

// LogAccountSubscriptionPlanCreationEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogAccountSubscriptionPlanCreationEvent(ctx context.Context, plan *types.AccountSubscriptionPlan) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("LogAccountSubscriptionPlanCreationEvent called")

	c.querier.LogAccountSubscriptionPlanCreationEvent(ctx, plan)
}

// AccountSubscriptionLogPlanUpdateEvent implements our AuditLogEntryDataManager interface.
func (c *Client) AccountSubscriptionLogPlanUpdateEvent(ctx context.Context, userID, planID uint64, changes []types.FieldChangeSummary) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("AccountSubscriptionLogPlanUpdateEvent called")

	c.querier.AccountSubscriptionLogPlanUpdateEvent(ctx, userID, planID, changes)
}

// AccountSubscriptionLogPlanArchiveEvent implements our AuditLogEntryDataManager interface.
func (c *Client) AccountSubscriptionLogPlanArchiveEvent(ctx context.Context, userID, planID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("AccountSubscriptionLogPlanArchiveEvent called")

	c.querier.AccountSubscriptionLogPlanArchiveEvent(ctx, userID, planID)
}

// GetAuditLogEntriesForAccountSubscriptionPlan fetches a list of audit log entries from the database that relate to a given plan.
func (c *Client) GetAuditLogEntriesForAccountSubscriptionPlan(ctx context.Context, planID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.PlanIDKey, planID).Debug("GetAuditLogEntriesForAccountSubscriptionPlan called")

	return c.querier.GetAuditLogEntriesForAccountSubscriptionPlan(ctx, planID)
}
