package requests

import (
	"context"
	"net/http"
	"strconv"

	observability "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
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

	logger := b.logger.WithValue(keys.AccountSubscriptionPlanIDKey, planID)

	tracing.AttachAccountSubscriptionPlanIDToSpan(span, planID)

	uri := b.BuildURL(
		ctx,
		nil,
		plansBasePath,
		id(planID),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building user status request")
	}

	return req, nil
}

// BuildGetAccountSubscriptionPlansRequest builds an HTTP request for fetching account subscription plans.
func (b *Builder) BuildGetAccountSubscriptionPlansRequest(ctx context.Context, filter *types.QueryFilter) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	logger := filter.AttachToLogger(b.logger)

	uri := b.BuildURL(ctx, filter.ToValues(), plansBasePath)
	tracing.AttachRequestURIToSpan(span, uri)
	tracing.AttachQueryFilterToSpan(span, filter)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building user status request")
	}

	return req, nil
}

// BuildCreateAccountSubscriptionPlanRequest builds an HTTP request for creating an account subscription plan.
func (b *Builder) BuildCreateAccountSubscriptionPlanRequest(ctx context.Context, input *types.AccountSubscriptionPlanCreationInput) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, ErrNilInputProvided
	}

	logger := b.logger.WithValue(keys.NameKey, input.Name)

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, logger, span, "validating input")
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

	logger := b.logger.WithValue(keys.AccountSubscriptionPlanIDKey, planID)

	tracing.AttachAccountSubscriptionPlanIDToSpan(span, planID)

	uri := b.BuildURL(
		ctx,
		nil,
		plansBasePath,
		id(planID),
	)
	tracing.AttachRequestURIToSpan(span, uri)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building user status request")
	}

	return req, nil
}

// BuildGetAuditLogForAccountSubscriptionPlanRequest builds an HTTP request for fetching a list of audit log entries pertaining to an account subscription plan.
func (b *Builder) BuildGetAuditLogForAccountSubscriptionPlanRequest(ctx context.Context, planID uint64) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if planID == 0 {
		return nil, ErrInvalidIDProvided
	}

	logger := b.logger.WithValue(keys.AccountSubscriptionPlanIDKey, planID)

	tracing.AttachAccountSubscriptionPlanIDToSpan(span, planID)

	uri := b.BuildURL(ctx, nil, plansBasePath, id(planID), "audit")
	tracing.AttachRequestURIToSpan(span, uri)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "building user status request")
	}

	return req, nil
}
