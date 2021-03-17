package requests

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	plansBasePath = "account_subscription_plans"
)

// BuildGetAccountSubscriptionPlanRequest builds an HTTP request for fetching an plan.
func (c *Builder) BuildGetAccountSubscriptionPlanRequest(ctx context.Context, planID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if planID == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := c.BuildURL(
		ctx,
		nil,
		plansBasePath,
		strconv.FormatUint(planID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildGetAccountSubscriptionPlansRequest builds an HTTP request for fetching account subscription plans.
func (c *Builder) BuildGetAccountSubscriptionPlansRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	uri := c.BuildURL(ctx, filter.ToValues(), plansBasePath)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildCreateAccountSubscriptionPlanRequest builds an HTTP request for creating an plan.
func (c *Builder) BuildCreateAccountSubscriptionPlanRequest(ctx context.Context, input *types.AccountSubscriptionPlanCreationInput) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		c.logger.Error(validationErr, "validating input")
		return nil, fmt.Errorf("validating input: %w", validationErr)
	}

	uri := c.BuildURL(ctx, nil, plansBasePath)
	tracing.AttachRequestURIToSpan(span, uri)

	return c.buildDataRequest(ctx, http.MethodPost, uri, input)
}

// BuildUpdateAccountSubscriptionPlanRequest builds an HTTP request for updating an plan.
func (c *Builder) BuildUpdateAccountSubscriptionPlanRequest(ctx context.Context, plan *types.AccountSubscriptionPlan) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if plan == nil {
		return nil, ErrNilInputProvided
	}

	uri := c.BuildURL(
		ctx,
		nil,
		plansBasePath,
		strconv.FormatUint(plan.ID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return c.buildDataRequest(ctx, http.MethodPut, uri, plan)
}

// BuildArchiveAccountSubscriptionPlanRequest builds an HTTP request for updating an plan.
func (c *Builder) BuildArchiveAccountSubscriptionPlanRequest(ctx context.Context, planID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if planID == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := c.BuildURL(
		ctx,
		nil,
		plansBasePath,
		strconv.FormatUint(planID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
}

// BuildGetAuditLogForAccountSubscriptionPlanRequest builds an HTTP request for fetching a list of audit log entries pertaining to an plan.
func (c *Builder) BuildGetAuditLogForAccountSubscriptionPlanRequest(ctx context.Context, planID uint64) (*http.Request, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if planID == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := c.BuildURL(ctx, nil, plansBasePath, strconv.FormatUint(planID, 10), "audit")
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}
