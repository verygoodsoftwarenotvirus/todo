package http

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// GetAccountSubscriptionPlan retrieves an account subscription plan.
func (c *Client) GetAccountSubscriptionPlan(ctx context.Context, planID uint64) (*types.AccountSubscriptionPlan, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if planID == 0 {
		return nil, ErrInvalidIDProvided
	}

	logger := c.logger.WithValue(keys.AccountSubscriptionPlanIDKey, planID)
	tracing.AttachAccountSubscriptionPlanIDToSpan(span, planID)

	req, err := c.requestBuilder.BuildGetAccountSubscriptionPlanRequest(ctx, planID)
	if err != nil {
		return nil, prepareError(err, logger, span, "building account subscription plan retrieval request")
	}

	var plan *types.AccountSubscriptionPlan
	if err = c.fetchAndUnmarshal(ctx, req, &plan); err != nil {
		return nil, prepareError(err, logger, span, "retrieving plan")
	}

	return plan, nil
}

// GetAccountSubscriptionPlans retrieves a list of account subscription plans.
func (c *Client) GetAccountSubscriptionPlans(ctx context.Context, filter *types.QueryFilter) (*types.AccountSubscriptionPlanList, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.loggerWithFilter(filter)

	tracing.AttachQueryFilterToSpan(span, filter)

	req, err := c.requestBuilder.BuildGetAccountSubscriptionPlansRequest(ctx, filter)
	if err != nil {
		return nil, prepareError(err, logger, span, "building account subscription plan list request")
	}

	var plans *types.AccountSubscriptionPlanList
	if err = c.fetchAndUnmarshal(ctx, req, &plans); err != nil {
		return nil, prepareError(err, logger, span, "retrieving plans")
	}

	return plans, nil
}

// CreateAccountSubscriptionPlan creates an account subscription plan.
func (c *Client) CreateAccountSubscriptionPlan(ctx context.Context, input *types.AccountSubscriptionPlanCreationInput) (*types.AccountSubscriptionPlan, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	logger := c.logger.WithValue("account_subscription_plan_name", input.Name)

	if err := input.Validate(ctx); err != nil {
		return nil, prepareError(err, logger, span, "validating input")
	}

	req, err := c.requestBuilder.BuildCreateAccountSubscriptionPlanRequest(ctx, input)
	if err != nil {
		return nil, prepareError(err, logger, span, "building account subscription plan creation request")
	}

	var plan *types.AccountSubscriptionPlan
	if err = c.fetchAndUnmarshal(ctx, req, &plan); err != nil {
		return nil, prepareError(err, logger, span, "creating plan")
	}

	return plan, nil
}

// UpdateAccountSubscriptionPlan updates an account subscription plan.
func (c *Client) UpdateAccountSubscriptionPlan(ctx context.Context, plan *types.AccountSubscriptionPlan) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if plan == nil {
		return ErrNilInputProvided
	}

	logger := c.logger.WithValue(keys.AccountSubscriptionPlanIDKey, plan.ID)
	tracing.AttachAccountSubscriptionPlanIDToSpan(span, plan.ID)

	req, err := c.requestBuilder.BuildUpdateAccountSubscriptionPlanRequest(ctx, plan)
	if err != nil {
		return prepareError(err, logger, span, "building account subscription plan update request")
	}

	if err = c.fetchAndUnmarshal(ctx, req, &plan); err != nil {
		return prepareError(err, logger, span, "updating account subscription plan")
	}

	return nil
}

// ArchiveAccountSubscriptionPlan archives an account subscription plan.
func (c *Client) ArchiveAccountSubscriptionPlan(ctx context.Context, planID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if planID == 0 {
		return ErrInvalidIDProvided
	}

	logger := c.logger.WithValue(keys.AccountSubscriptionPlanIDKey, planID)
	tracing.AttachAccountSubscriptionPlanIDToSpan(span, planID)

	req, err := c.requestBuilder.BuildArchiveAccountSubscriptionPlanRequest(ctx, planID)
	if err != nil {
		return prepareError(err, logger, span, "building account subscription plan archive request")
	}

	if err = c.fetchAndUnmarshal(ctx, req, nil); err != nil {
		return prepareError(err, logger, span, "archiving account subscription plan")
	}

	return nil
}

// GetAuditLogForAccountSubscriptionPlan retrieves a list of audit log entries pertaining to an account subscription plan.
func (c *Client) GetAuditLogForAccountSubscriptionPlan(ctx context.Context, planID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if planID == 0 {
		return nil, ErrInvalidIDProvided
	}

	logger := c.logger.WithValue(keys.AccountSubscriptionPlanIDKey, planID)
	tracing.AttachAccountSubscriptionPlanIDToSpan(span, planID)

	req, err := c.requestBuilder.BuildGetAuditLogForAccountSubscriptionPlanRequest(ctx, planID)
	if err != nil {
		return nil, prepareError(err, logger, span, "building fetch audit log entries for account subscription plan request")
	}

	var entries []*types.AuditLogEntry
	if err = c.fetchAndUnmarshal(ctx, req, &entries); err != nil {
		return nil, prepareError(err, logger, span, "retrieving audit log entries for account subscription plan")
	}

	return entries, nil
}
