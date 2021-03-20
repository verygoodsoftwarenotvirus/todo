package accountsubscriptionplans

import (
	"database/sql"
	"errors"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// AccountSubscriptionPlanIDURIParamKey is a standard string that we'll use to refer to plan IDs with.
	AccountSubscriptionPlanIDURIParamKey = "accountSubscriptionPlanID"
)

// ListHandler is our list route.
func (s *service) ListHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Debug("ListHandler invoked")

	// ensure query filter.
	filter := types.ExtractQueryFilter(req)

	// determine user ID.
	reqCtx, requestContextRetrievalErr := s.requestContextFetcher(req)
	if requestContextRetrievalErr != nil {
		s.logger.Error(requestContextRetrievalErr, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.UserIDKey, reqCtx.User.ID)

	plans, err := s.planDataManager.GetAccountSubscriptionPlans(ctx, filter)

	if errors.Is(err, sql.ErrNoRows) {
		// in the event no rows exist return an empty list.
		plans = &types.AccountSubscriptionPlanList{AccountSubscriptionPlans: []*types.AccountSubscriptionPlan{}}
	} else if err != nil {
		logger.Error(err, "error encountered fetching accountsubscriptionplans")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, plans)
}

// CreateHandler is our plan creation route.
func (s *service) CreateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Debug("CreateHandler called")

	// check request context for parsed input struct.
	input, ok := ctx.Value(createMiddlewareCtxKey).(*types.AccountSubscriptionPlanCreationInput)
	if !ok {
		logger.Info("valid input not attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	// determine user ID.
	reqCtx, requestContextRetrievalErr := s.requestContextFetcher(req)
	if requestContextRetrievalErr != nil {
		s.logger.Error(requestContextRetrievalErr, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.UserIDKey, reqCtx.User.ID)

	// create plan in database.
	x, err := s.planDataManager.CreateAccountSubscriptionPlan(ctx, input)
	if err != nil {
		logger.Error(err, "error creating plan")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	logger.WithValues(map[string]interface{}{
		"id":          x.ID,
		"name":        x.Name,
		"description": x.Description,
		"price":       x.Price,
		"period":      x.Period,
	}).Info("created plan")

	tracing.AttachAccountSubscriptionPlanIDToSpan(span, x.ID)
	logger = logger.WithValue(keys.AccountSubscriptionPlanIDKey, x.ID)
	logger.Debug("plan created")

	s.planCounter.Increment(ctx)
	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, x, http.StatusCreated)
}

// ReadHandler returns a GET handler that returns an plan.
func (s *service) ReadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	reqCtx, requestContextRetrievalErr := s.requestContextFetcher(req)
	if requestContextRetrievalErr != nil {
		s.logger.Error(requestContextRetrievalErr, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.UserIDKey, reqCtx.User.ID)

	// determine plan ID.
	planID := s.planIDFetcher(req)
	tracing.AttachAccountSubscriptionPlanIDToSpan(span, planID)
	logger = logger.WithValue(keys.AccountSubscriptionPlanIDKey, planID)

	// fetch plan from database.
	x, err := s.planDataManager.GetAccountSubscriptionPlan(ctx, planID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		logger.Error(err, "error fetching plan from database")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, x)
}

// UpdateHandler returns a handler that updates an plan.
func (s *service) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// check for parsed input attached to request context.
	input, ok := ctx.Value(updateMiddlewareCtxKey).(*types.AccountSubscriptionPlanUpdateInput)
	if !ok {
		logger.Info("no input attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	// determine user ID.
	reqCtx, requestContextRetrievalErr := s.requestContextFetcher(req)
	if requestContextRetrievalErr != nil {
		s.logger.Error(requestContextRetrievalErr, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.UserIDKey, reqCtx.User.ID)

	// determine plan ID.
	planID := s.planIDFetcher(req)
	logger = logger.WithValue(keys.AccountSubscriptionPlanIDKey, planID)
	tracing.AttachAccountSubscriptionPlanIDToSpan(span, planID)

	// fetch plan from database.
	x, err := s.planDataManager.GetAccountSubscriptionPlan(ctx, planID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		logger.Error(err, "error encountered getting plan")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// update the data structure.
	changeReport := x.Update(input)

	// update plan in database.
	if err = s.planDataManager.UpdateAccountSubscriptionPlan(ctx, x, reqCtx.User.ID, changeReport); err != nil {
		logger.Error(err, "error encountered updating plan")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, x)
}

// ArchiveHandler returns a handler that archives an plan.
func (s *service) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	reqCtx, requestContextRetrievalErr := s.requestContextFetcher(req)
	if requestContextRetrievalErr != nil {
		s.logger.Error(requestContextRetrievalErr, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.UserIDKey, reqCtx.User.ID)

	// determine plan ID.
	planID := s.planIDFetcher(req)
	logger = logger.WithValue(keys.AccountSubscriptionPlanIDKey, planID)
	tracing.AttachAccountSubscriptionPlanIDToSpan(span, planID)

	// archive the plan in the database.
	err := s.planDataManager.ArchiveAccountSubscriptionPlan(ctx, planID, reqCtx.User.ID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		logger.Error(err, "error encountered deleting plan")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// notify relevant parties.
	s.planCounter.Decrement(ctx)

	// encode our response and peace.
	res.WriteHeader(http.StatusNoContent)
}

// AuditEntryHandler returns a GET handler that returns all audit log entries related to an plan.
func (s *service) AuditEntryHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Debug("AuditEntryHandler invoked")

	// determine user ID.
	reqCtx, requestContextRetrievalErr := s.requestContextFetcher(req)
	if requestContextRetrievalErr != nil {
		s.logger.Error(requestContextRetrievalErr, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.UserIDKey, reqCtx.User.ID)

	// determine plan ID.
	planID := s.planIDFetcher(req)
	tracing.AttachAccountSubscriptionPlanIDToSpan(span, planID)
	logger = logger.WithValue(keys.AccountSubscriptionPlanIDKey, planID)

	x, err := s.planDataManager.GetAuditLogEntriesForAccountSubscriptionPlan(ctx, planID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		logger.Error(err, "error encountered fetching audit log entries")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	logger.WithValue("entry_count", len(x)).Debug("returning from AuditEntryHandler")

	// encode our response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, x)
}
