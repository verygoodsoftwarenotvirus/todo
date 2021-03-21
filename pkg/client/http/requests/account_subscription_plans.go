package requests

import (
	"context"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/errs"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	plansBasePath = "account_subscription_plans"
)

// BuildGetAccountSubscriptionPlanRequest builds an HTTP request for fetching an account subscription plan.
func (b *Builder) BuildGetAccountSubscriptionPlanRequest(ctx context.Context, planID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if planID == 0 {
		return nil, ErrInvalidIDProvided
	}

	tracing.AttachAccountSubscriptionPlanIDToSpan(span, planID)

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
	tracing.AttachQueryFilterToSpan(span, filter)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// BuildCreateAccountSubscriptionPlanRequest builds an HTTP request for creating an account subscription plan.
func (b *Builder) BuildCreateAccountSubscriptionPlanRequest(ctx context.Context, input *types.AccountSubscriptionPlanCreationInput) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	logger := b.logger.WithValue(keys.NameKey, input.Name)

	if err := input.Validate(ctx); err != nil {
		return nil, errs.PrepareError(err, logger, span, "validating input")
	}

	uri := b.BuildURL(ctx, nil, plansBasePath)
	tracing.AttachRequestURIToSpan(span, uri)

	return b.buildDataRequest(ctx, http.MethodPost, uri, input)
}

// BuildUpdateAccountSubscriptionPlanRequest builds an HTTP request for updating an account subscription plan.
func (b *Builder) BuildUpdateAccountSubscriptionPlanRequest(ctx context.Context, plan *types.AccountSubscriptionPlan) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if plan == nil {
		return nil, ErrNilInputProvided
	}

	tracing.AttachAccountSubscriptionPlanIDToSpan(span, plan.ID)

	uri := b.BuildURL(
		ctx,
		nil,
		plansBasePath,
		strconv.FormatUint(plan.ID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return b.buildDataRequest(ctx, http.MethodPut, uri, plan)
}

// BuildArchiveAccountSubscriptionPlanRequest builds an HTTP request for archiving an account subscription plan.
func (b *Builder) BuildArchiveAccountSubscriptionPlanRequest(ctx context.Context, planID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if planID == 0 {
		return nil, ErrInvalidIDProvided
	}

	tracing.AttachAccountSubscriptionPlanIDToSpan(span, planID)

	uri := b.BuildURL(
		ctx,
		nil,
		plansBasePath,
		strconv.FormatUint(planID, 10),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
}

// BuildGetAuditLogForAccountSubscriptionPlanRequest builds an HTTP request for fetching a list of audit log entries pertaining to an account subscription plan.
func (b *Builder) BuildGetAuditLogForAccountSubscriptionPlanRequest(ctx context.Context, planID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if planID == 0 {
		return nil, ErrInvalidIDProvided
	}

	tracing.AttachAccountSubscriptionPlanIDToSpan(span, planID)

	uri := b.BuildURL(ctx, nil, plansBasePath, strconv.FormatUint(planID, 10), "audit")
	tracing.AttachRequestURIToSpan(span, uri)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}
