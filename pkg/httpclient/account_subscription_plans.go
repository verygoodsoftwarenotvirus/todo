package httpclient

import (
	"context"
	"fmt"

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

	var plan *types.AccountSubscriptionPlan
	logger := c.logger.WithValue(keys.AccountSubscriptionPlanIDKey, planID)

	req, err := c.requestBuilder.BuildGetAccountSubscriptionPlanRequest(ctx, planID)
	if err != nil {
		logger.Error(err, "building request")
		tracing.AttachErrorToSpan(span, err)

		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &plan); retrieveErr != nil {
		logger.Error(retrieveErr, "retrieving plan")
		tracing.AttachErrorToSpan(span, retrieveErr)

		return nil, retrieveErr
	}

	return plan, nil
}

// GetAccountSubscriptionPlans retrieves a list of account subscription plans.
func (c *Client) GetAccountSubscriptionPlans(ctx context.Context, filter *types.QueryFilter) (*types.AccountSubscriptionPlanList, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	var plans *types.AccountSubscriptionPlanList
	logger := c.logger.
		WithValue(keys.FilterLimitKey, filter.Limit).
		WithValue(keys.FilterIsNilKey, filter == nil).
		WithValue(keys.FilterPageKey, filter.Page)

	tracing.AttachQueryFilterToSpan(span, filter)

	req, err := c.requestBuilder.BuildGetAccountSubscriptionPlansRequest(ctx, filter)
	if err != nil {
		logger.Error(err, "building request")
		tracing.AttachErrorToSpan(span, err)

		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &plans); retrieveErr != nil {
		logger.Error(err, "retrieving plans")
		tracing.AttachErrorToSpan(span, err)

		return nil, retrieveErr
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

	var plan *types.AccountSubscriptionPlan
	logger := c.logger.WithValue("account_subscription_plan_name", input.Name)

	if err := input.Validate(ctx); err != nil {
		logger.Error(err, "validating input")
		tracing.AttachErrorToSpan(span, err)

		return nil, fmt.Errorf("validating input: %w", err)
	}

	req, err := c.requestBuilder.BuildCreateAccountSubscriptionPlanRequest(ctx, input)
	if err != nil {
		logger.Error(err, "building request")
		tracing.AttachErrorToSpan(span, err)

		return nil, fmt.Errorf("building request: %w", err)
	}

	if err = c.executeRequest(ctx, req, &plan); err != nil {
		logger.Error(err, "creating plan")
		tracing.AttachErrorToSpan(span, err)

		return nil, fmt.Errorf("creating plan: %w", err)
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

	req, err := c.requestBuilder.BuildUpdateAccountSubscriptionPlanRequest(ctx, plan)
	if err != nil {
		logger.Error(err, "building request")
		tracing.AttachErrorToSpan(span, err)

		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, &plan)
}

// ArchiveAccountSubscriptionPlan archives an account subscription plan.
func (c *Client) ArchiveAccountSubscriptionPlan(ctx context.Context, planID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if planID == 0 {
		return ErrInvalidIDProvided
	}

	logger := c.logger.WithValue(keys.AccountSubscriptionPlanIDKey, planID)

	req, err := c.requestBuilder.BuildArchiveAccountSubscriptionPlanRequest(ctx, planID)
	if err != nil {
		logger.Error(err, "building request")
		tracing.AttachErrorToSpan(span, err)

		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// GetAuditLogForAccountSubscriptionPlan retrieves a list of audit log entries pertaining to an account subscription plan.
func (c *Client) GetAuditLogForAccountSubscriptionPlan(ctx context.Context, planID uint64) ([]*types.AuditLogEntry, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if planID == 0 {
		return nil, ErrInvalidIDProvided
	}

	logger := c.logger.WithValue(keys.AccountSubscriptionPlanIDKey, planID)

	req, err := c.requestBuilder.BuildGetAuditLogForAccountSubscriptionPlanRequest(ctx, planID)
	if err != nil {
		logger.Error(err, "building request")
		tracing.AttachErrorToSpan(span, err)

		return nil, fmt.Errorf("building request: %w", err)
	}

	var entries []*types.AuditLogEntry

	if retrieveErr := c.retrieve(ctx, req, &entries); retrieveErr != nil {
		logger.Error(retrieveErr, "building request")
		tracing.AttachErrorToSpan(span, retrieveErr)

		return nil, retrieveErr
	}

	return entries, nil
}
