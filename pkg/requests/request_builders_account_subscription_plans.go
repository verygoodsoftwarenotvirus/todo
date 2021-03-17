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
func (b *Builder) BuildGetAccountSubscriptionPlanRequest(ctx context.Context, planID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if planID == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := b.BuildURL(
		ctx,
		nil,
		plansBasePath,
		strconv.FormatUint(planID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildGetAccountSubscriptionPlansRequest builds an HTTP request for fetching account subscription plans.
func (b *Builder) BuildGetAccountSubscriptionPlansRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	uri := b.BuildURL(ctx, filter.ToValues(), plansBasePath)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildCreateAccountSubscriptionPlanRequest builds an HTTP request for creating an plan.
func (b *Builder) BuildCreateAccountSubscriptionPlanRequest(ctx context.Context, input *types.AccountSubscriptionPlanCreationInput) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	if validationErr := input.Validate(ctx); validationErr != nil {
		b.logger.Error(validationErr, "validating input")
		return nil, fmt.Errorf("validating input: %w", validationErr)
	}

	uri := b.BuildURL(ctx, nil, plansBasePath)
	tracing.AttachRequestURIToSpan(span, uri)

	return b.buildDataRequest(ctx, http.MethodPost, uri, input)
}

// BuildUpdateAccountSubscriptionPlanRequest builds an HTTP request for updating an plan.
func (b *Builder) BuildUpdateAccountSubscriptionPlanRequest(ctx context.Context, plan *types.AccountSubscriptionPlan) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if plan == nil {
		return nil, ErrNilInputProvided
	}

	uri := b.BuildURL(
		ctx,
		nil,
		plansBasePath,
		strconv.FormatUint(plan.ID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return b.buildDataRequest(ctx, http.MethodPut, uri, plan)
}

// BuildArchiveAccountSubscriptionPlanRequest builds an HTTP request for updating an plan.
func (b *Builder) BuildArchiveAccountSubscriptionPlanRequest(ctx context.Context, planID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if planID == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := b.BuildURL(
		ctx,
		nil,
		plansBasePath,
		strconv.FormatUint(planID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
}

// BuildGetAuditLogForAccountSubscriptionPlanRequest builds an HTTP request for fetching a list of audit log entries pertaining to an plan.
func (b *Builder) BuildGetAuditLogForAccountSubscriptionPlanRequest(ctx context.Context, planID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if planID == 0 {
		return nil, ErrInvalidIDProvided
	}

	uri := b.BuildURL(ctx, nil, plansBasePath, strconv.FormatUint(planID, 10), "audit")
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}
