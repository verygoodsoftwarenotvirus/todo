package accountsubscriptionplans

import (
	"database/sql"
	"errors"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
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

	filter := types.ExtractQueryFilter(req)
	logger := s.logger.WithRequest(req).
		WithValue(keys.FilterLimitKey, filter.Limit).
		WithValue(keys.FilterPageKey, filter.Page).
		WithValue(keys.FilterSortByKey, string(filter.SortBy))

	tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)
	logger = logger.WithValue(keys.RequesterIDKey, sessionCtxData.Requester.ID)

	accountSubscriptionPlans, err := s.accountSubscriptionPlanDataManager.GetAccountSubscriptionPlans(ctx, filter)

	if errors.Is(err, sql.ErrNoRows) {
		// in the event no rows exist return an empty list.
		accountSubscriptionPlans = &types.AccountSubscriptionPlanList{AccountSubscriptionPlans: []*types.AccountSubscriptionPlan{}}
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching account subscription accountSubscriptionPlans")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, accountSubscriptionPlans)
}

// CreateHandler is our plan creation route.
func (s *service) CreateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// check session context data for parsed input struct.
	input, ok := ctx.Value(createMiddlewareCtxKey).(*types.AccountSubscriptionPlanCreationInput)
	if !ok {
		logger.Info("valid input not attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)
	logger = logger.WithValue(keys.RequesterIDKey, sessionCtxData.Requester.ID)

	// create plan in database.
	accountSubscriptionPlan, err := s.accountSubscriptionPlanDataManager.CreateAccountSubscriptionPlan(ctx, input)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "creating plan")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	tracing.AttachAccountSubscriptionPlanIDToSpan(span, accountSubscriptionPlan.ID)
	logger = logger.WithValue(keys.AccountSubscriptionPlanIDKey, accountSubscriptionPlan.ID)
	logger.Debug("plan created")

	s.planCounter.Increment(ctx)
	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, accountSubscriptionPlan, http.StatusCreated)
}

// ReadHandler returns a GET handler that returns an account subscription plan.
func (s *service) ReadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)
	logger = logger.WithValue(keys.RequesterIDKey, sessionCtxData.Requester.ID)

	// determine plan ID.
	accountSubscriptionPlanID := s.accountSubscriptionPlanIDFetcher(req)
	tracing.AttachAccountSubscriptionPlanIDToSpan(span, accountSubscriptionPlanID)
	logger = logger.WithValue(keys.AccountSubscriptionPlanIDKey, accountSubscriptionPlanID)

	// fetch plan from database.
	accountSubscriptionPlan, err := s.accountSubscriptionPlanDataManager.GetAccountSubscriptionPlan(ctx, accountSubscriptionPlanID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching account subscription plan")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, accountSubscriptionPlan)
}

// UpdateHandler returns a handler that updates an account subscription plan.
func (s *service) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// check for parsed input attached to session context data.
	input, ok := ctx.Value(updateMiddlewareCtxKey).(*types.AccountSubscriptionPlanUpdateInput)
	if !ok {
		logger.Info("no input attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)
	logger = logger.WithValue(keys.RequesterIDKey, sessionCtxData.Requester.ID)

	// determine plan ID.
	accountSubscriptionPlanID := s.accountSubscriptionPlanIDFetcher(req)
	logger = logger.WithValue(keys.AccountSubscriptionPlanIDKey, accountSubscriptionPlanID)
	tracing.AttachAccountSubscriptionPlanIDToSpan(span, accountSubscriptionPlanID)

	// fetch plan from database.
	accountSubscriptionPlan, err := s.accountSubscriptionPlanDataManager.GetAccountSubscriptionPlan(ctx, accountSubscriptionPlanID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching account subscription plan")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// update the data structure.
	changeReport := accountSubscriptionPlan.Update(input)
	tracing.AttachChangeSummarySpan(span, "account_subscription_plan", changeReport)

	// update plan in database.
	if err = s.accountSubscriptionPlanDataManager.UpdateAccountSubscriptionPlan(ctx, accountSubscriptionPlan, sessionCtxData.Requester.ID, changeReport); err != nil {
		observability.AcknowledgeError(err, logger, span, "updating account subscription plan")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, accountSubscriptionPlan)
}

// ArchiveHandler returns a handler that archives an account subscription plan.
func (s *service) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)
	logger = logger.WithValue(keys.RequesterIDKey, sessionCtxData.Requester.ID)

	// determine plan ID.
	planID := s.accountSubscriptionPlanIDFetcher(req)
	logger = logger.WithValue(keys.AccountSubscriptionPlanIDKey, planID)
	tracing.AttachAccountSubscriptionPlanIDToSpan(span, planID)

	// archive the plan in the database.
	err = s.accountSubscriptionPlanDataManager.ArchiveAccountSubscriptionPlan(ctx, planID, sessionCtxData.Requester.ID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "archiving account subscription plan")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// notify relevant parties.
	s.planCounter.Decrement(ctx)

	// encode our response and peace.
	res.WriteHeader(http.StatusNoContent)
}

// AuditEntryHandler returns a GET handler that returns all audit log entries related to an account subscription plan.
func (s *service) AuditEntryHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)
	logger = logger.WithValue(keys.RequesterIDKey, sessionCtxData.Requester.ID)

	// determine plan ID.
	planID := s.accountSubscriptionPlanIDFetcher(req)
	tracing.AttachAccountSubscriptionPlanIDToSpan(span, planID)
	logger = logger.WithValue(keys.AccountSubscriptionPlanIDKey, planID)

	auditLogEntries, err := s.accountSubscriptionPlanDataManager.GetAuditLogEntriesForAccountSubscriptionPlan(ctx, planID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching audit log entries")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	logger.WithValue("entry_count", len(auditLogEntries)).Debug("returning from AuditEntryHandler")

	// encode our response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, auditLogEntries)
}
