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
	plansBasePath = "plans"
)

// BuildGetPlanRequest builds an HTTP request for fetching an plan.
func (c *Client) BuildGetPlanRequest(ctx context.Context, planID uint64) (*http.Request, error) {
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

// GetPlan retrieves an plan.
func (c *Client) GetPlan(ctx context.Context, planID uint64) (plan *types.AccountSubscriptionPlan, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildGetPlanRequest(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &plan); retrieveErr != nil {
		return nil, retrieveErr
	}

	return plan, nil
}

// BuildGetPlansRequest builds an HTTP request for fetching plans.
func (c *Client) BuildGetPlansRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(
		filter.ToValues(),
		plansBasePath,
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// GetPlans retrieves a list of plans.
func (c *Client) GetPlans(ctx context.Context, filter *types.QueryFilter) (plans *types.AccountSubscriptionPlanList, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildGetPlansRequest(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &plans); retrieveErr != nil {
		return nil, retrieveErr
	}

	return plans, nil
}

// BuildCreatePlanRequest builds an HTTP request for creating an plan.
func (c *Client) BuildCreatePlanRequest(ctx context.Context, input *types.AccountSubscriptionPlanCreationInput) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(
		nil,
		plansBasePath,
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return c.buildDataRequest(ctx, http.MethodPost, uri, input)
}

// CreatePlan creates an plan.
func (c *Client) CreatePlan(ctx context.Context, input *types.AccountSubscriptionPlanCreationInput) (plan *types.AccountSubscriptionPlan, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildCreatePlanRequest(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	err = c.executeRequest(ctx, req, &plan)

	return plan, err
}

// BuildUpdatePlanRequest builds an HTTP request for updating an plan.
func (c *Client) BuildUpdatePlanRequest(ctx context.Context, plan *types.AccountSubscriptionPlan) (*http.Request, error) {
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

// UpdatePlan updates an plan.
func (c *Client) UpdatePlan(ctx context.Context, plan *types.AccountSubscriptionPlan) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildUpdatePlanRequest(ctx, plan)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, &plan)
}

// BuildArchivePlanRequest builds an HTTP request for updating an plan.
func (c *Client) BuildArchivePlanRequest(ctx context.Context, planID uint64) (*http.Request, error) {
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

// ArchivePlan archives an plan.
func (c *Client) ArchivePlan(ctx context.Context, planID uint64) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildArchivePlanRequest(ctx, planID)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	return c.executeRequest(ctx, req, nil)
}

// BuildGetAuditLogForPlanRequest builds an HTTP request for fetching a list of audit log entries pertaining to an plan.
func (c *Client) BuildGetAuditLogForPlanRequest(ctx context.Context, planID uint64) (*http.Request, error) {
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

// GetAuditLogForPlan retrieves a list of audit log entries pertaining to an plan.
func (c *Client) GetAuditLogForPlan(ctx context.Context, planID uint64) (entries []*types.AuditLogEntry, err error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	req, err := c.BuildGetAuditLogForPlanRequest(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	if retrieveErr := c.retrieve(ctx, req, &entries); retrieveErr != nil {
		return nil, retrieveErr
	}

	return entries, nil
}
