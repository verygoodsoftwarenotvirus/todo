package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	plansBasePath = "accountsubscriptionplans"
)

// BuildGetAccountSubscriptionPlanRequest builds an HTTP request for fetching an plan.
func (c *Client) BuildGetAccountSubscriptionPlanRequest(ctx context.Context, planID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(
		nil,
		plansBasePath,
		strconv.FormatUint(planID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetAccountSubscriptionPlan retrieves an plan.
func (c *Client) GetAccountSubscriptionPlan(ctx context.Context, planID uint64) (plan *types.AccountSubscriptionPlan, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildGetAccountSubscriptionPlanRequest(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &plan); retrieveErr != nil {
		return nil, retrieveErr
	}

	return plan, nil
}

// BuildGetAccountSubscriptionPlansRequest builds an HTTP request for fetching account subscription plans.
func (c *Client) BuildGetAccountSubscriptionPlansRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(
		filter.ToValues(),
		plansBasePath,
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
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

// BuildCreateAccountSubscriptionPlanRequest builds an HTTP request for creating an plan.
func (c *Client) BuildCreateAccountSubscriptionPlanRequest(ctx context.Context, input *types.AccountSubscriptionPlanCreationInput) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(
		nil,
		plansBasePath,
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return c.buildDataRequest(ctx, http.MethodPost, uri, input)
}

// CreateAccountSubscriptionPlan creates an plan.
func (c *Client) CreateAccountSubscriptionPlan(ctx context.Context, input *types.AccountSubscriptionPlanCreationInput) (plan *types.AccountSubscriptionPlan, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildCreateAccountSubscriptionPlanRequest(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.executeRequest(ctx, req, &plan)

	return plan, err
}

// BuildUpdateAccountSubscriptionPlanRequest builds an HTTP request for updating an plan.
func (c *Client) BuildUpdateAccountSubscriptionPlanRequest(ctx context.Context, plan *types.AccountSubscriptionPlan) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(
		nil,
		plansBasePath,
		strconv.FormatUint(plan.ID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return c.buildDataRequest(ctx, http.MethodPut, uri, plan)
}

// UpdateAccountSubscriptionPlan updates an plan.
func (c *Client) UpdateAccountSubscriptionPlan(ctx context.Context, plan *types.AccountSubscriptionPlan) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildUpdateAccountSubscriptionPlanRequest(ctx, plan)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, &plan)
}

// BuildArchiveAccountSubscriptionPlanRequest builds an HTTP request for updating an plan.
func (c *Client) BuildArchiveAccountSubscriptionPlanRequest(ctx context.Context, planID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(
		nil,
		plansBasePath,
		strconv.FormatUint(planID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
}

// ArchiveAccountSubscriptionPlan archives an plan.
func (c *Client) ArchiveAccountSubscriptionPlan(ctx context.Context, planID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildArchiveAccountSubscriptionPlanRequest(ctx, planID)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// BuildGetAuditLogForAccountSubscriptionPlanRequest builds an HTTP request for fetching a list of audit log entries pertaining to an plan.
func (c *Client) BuildGetAuditLogForAccountSubscriptionPlanRequest(ctx context.Context, planID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(
		nil,
		plansBasePath,
		strconv.FormatUint(planID, 10),
		"audit",
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetAuditLogForAccountSubscriptionPlan retrieves a list of audit log entries pertaining to an plan.
func (c *Client) GetAuditLogForAccountSubscriptionPlan(ctx context.Context, planID uint64) (entries []*types.AuditLogEntry, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildGetAuditLogForAccountSubscriptionPlanRequest(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &entries); retrieveErr != nil {
		return nil, retrieveErr
	}

	return entries, nil
}
