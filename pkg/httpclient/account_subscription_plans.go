package httpclient

import (
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// GetAccountSubscriptionPlan retrieves an plan.
func (c *Client) GetAccountSubscriptionPlan(ctx context.Context, planID uint64) (plan *types.AccountSubscriptionPlan, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if planID == 0 {
		return nil, ErrInvalidIDProvided
	}

	req, err := c.BuildGetAccountSubscriptionPlanRequest(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &plan); retrieveErr != nil {
		return nil, retrieveErr
	}

	return plan, nil
}

// GetAccountSubscriptionPlans retrieves a list of account subscription plans.
func (c *Client) GetAccountSubscriptionPlans(ctx context.Context, filter *types.QueryFilter) (plans *types.AccountSubscriptionPlanList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildGetAccountSubscriptionPlansRequest(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &plans); retrieveErr != nil {
		return nil, retrieveErr
	}

	return plans, nil
}

// CreateAccountSubscriptionPlan creates an plan.
func (c *Client) CreateAccountSubscriptionPlan(ctx context.Context, input *types.AccountSubscriptionPlanCreationInput) (plan *types.AccountSubscriptionPlan, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		c.logger.Error(validationErr, "validating input")
		return nil, fmt.Errorf("validating input: %w", validationErr)
	}

	req, err := c.BuildCreateAccountSubscriptionPlanRequest(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.executeRequest(ctx, req, &plan)

	return plan, err
}

// UpdateAccountSubscriptionPlan updates an plan.
func (c *Client) UpdateAccountSubscriptionPlan(ctx context.Context, plan *types.AccountSubscriptionPlan) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if plan == nil {
		return ErrNilInputProvided
	}

	req, err := c.BuildUpdateAccountSubscriptionPlanRequest(ctx, plan)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, &plan)
}

// ArchiveAccountSubscriptionPlan archives an plan.
func (c *Client) ArchiveAccountSubscriptionPlan(ctx context.Context, planID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if planID == 0 {
		return ErrInvalidIDProvided
	}

	req, err := c.BuildArchiveAccountSubscriptionPlanRequest(ctx, planID)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// GetAuditLogForAccountSubscriptionPlan retrieves a list of audit log entries pertaining to an plan.
func (c *Client) GetAuditLogForAccountSubscriptionPlan(ctx context.Context, planID uint64) (entries []*types.AuditLogEntry, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if planID == 0 {
		return nil, ErrInvalidIDProvided
	}

	req, err := c.BuildGetAuditLogForAccountSubscriptionPlanRequest(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &entries); retrieveErr != nil {
		return nil, retrieveErr
	}

	return entries, nil
}
