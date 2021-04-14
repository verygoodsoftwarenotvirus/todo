package accounts

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
	// AccountIDURIParamKey is a standard string that we'll use to refer to account IDs with.
	AccountIDURIParamKey = "accountID"
	// UserIDURIParamKey is a standard string that we'll use to refer to user IDs with.
	UserIDURIParamKey = "userID"
)

// parseBool differs from strconv.ParseBool in that it returns false by default.
func parseBool(str string) bool {
	switch str {
	case "1", "t", "T", "true", "TRUE", "True":
		return true
	default:
		return false
	}
}

// ListHandler is our list route.
func (s *service) ListHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	filter := types.ExtractQueryFilter(req)
	logger := s.logger.WithRequest(req).
		WithValue(keys.FilterLimitKey, filter.Limit).
		WithValue(keys.FilterPageKey, filter.Page).
		WithValue(keys.FilterSortByKey, string(filter.SortBy))

	tracing.AttachRequestToSpan(span, req)
	tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))

	// fetch request context
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	requester := reqCtx.Requester.ID
	logger = logger.WithValue(keys.RequesterIDKey, requester)
	tracing.AttachRequestContextToSpan(span, reqCtx)

	// determine if this is an admin request
	rawQueryAdminKey := req.URL.Query().Get("admin")
	adminQueryPresent := parseBool(rawQueryAdminKey)
	isAdminRequest := reqCtx.Requester.ServiceAdminPermission.IsServiceAdmin() && adminQueryPresent

	var accounts *types.AccountList

	if reqCtx.Requester.ServiceAdminPermission.IsServiceAdmin() && isAdminRequest {
		accounts, err = s.accountDataManager.GetAccountsForAdmin(ctx, filter)
	} else {
		accounts, err = s.accountDataManager.GetAccounts(ctx, requester, filter)
	}

	if errors.Is(err, sql.ErrNoRows) {
		// in the event no rows exist, return an empty list.
		accounts = &types.AccountList{Accounts: []*types.Account{}}
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching accounts")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and say farewell.
	s.encoderDecoder.RespondWithData(ctx, res, accounts)
}

// CreateHandler is our account creation route.
func (s *service) CreateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// check request context for parsed input struct.
	input, ok := ctx.Value(createMiddlewareCtxKey).(*types.AccountCreationInput)
	if !ok {
		logger.Info("valid input not attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	logger = logger.WithValue(keys.NameKey, input.Name)

	// retrieve request context.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	requester := reqCtx.Requester.ID
	logger = logger.WithValue(keys.RequesterIDKey, requester)
	input.BelongsToUser = requester

	// create account in database.
	account, err := s.accountDataManager.CreateAccount(ctx, input, requester)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "creating account")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	logger = logger.WithValue(keys.AccountIDKey, account.ID)
	tracing.AttachAccountIDToSpan(span, account.ID)

	// notify relevant parties.
	logger.Debug("created account")
	s.accountCounter.Increment(ctx)

	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, account, http.StatusCreated)
}

// ReadHandler returns a GET handler that returns an account.
func (s *service) ReadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// determine user ID.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	requester := reqCtx.Requester.ID
	logger = logger.WithValue(keys.RequesterIDKey, requester)
	tracing.AttachRequestContextToSpan(span, reqCtx)

	// determine account ID.
	accountID := s.accountIDFetcher(req)
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	// fetch account from database.
	account, err := s.accountDataManager.GetAccount(ctx, accountID, requester)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching account from database")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, account)
}

// UpdateHandler returns a handler that updates an account.
func (s *service) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// check for parsed input attached to request context.
	input, ok := ctx.Value(updateMiddlewareCtxKey).(*types.AccountUpdateInput)
	if !ok {
		logger.Info("no input attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	// determine user ID.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	requester := reqCtx.Requester.ID
	logger = logger.WithValue(keys.RequesterIDKey, requester)
	input.BelongsToUser = requester

	// determine account ID.
	accountID := s.accountIDFetcher(req)
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	// fetch account from database.
	account, err := s.accountDataManager.GetAccount(ctx, accountID, requester)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching account from database")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// update the data structure.
	changeReport := account.Update(input)
	tracing.AttachChangeSummarySpan(span, "account", changeReport)

	// update account in database.
	if err = s.accountDataManager.UpdateAccount(ctx, account, requester, changeReport); err != nil {
		observability.AcknowledgeError(err, logger, span, "updating account")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, account)
}

// ArchiveHandler returns a handler that archives an account.
func (s *service) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// determine user ID.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	requester := reqCtx.Requester.ID
	logger = logger.WithValue(keys.RequesterIDKey, requester)
	tracing.AttachRequestContextToSpan(span, reqCtx)

	// determine account ID.
	accountID := s.accountIDFetcher(req)
	tracing.AttachAccountIDToSpan(span, accountID)
	logger = logger.WithValue(keys.AccountIDKey, accountID)

	// archive the account in the database.
	err = s.accountDataManager.ArchiveAccount(ctx, accountID, requester, requester)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "archiving account")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// notify relevant parties.
	s.accountCounter.Decrement(ctx)

	// encode our response and peace.
	res.WriteHeader(http.StatusNoContent)
}

// AddUserHandler is our account creation route.
func (s *service) AddUserHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// check request context for parsed input struct.
	input, ok := ctx.Value(addUserToAccountMiddlewareCtxKey).(*types.AddUserToAccountInput)
	if !ok {
		logger.Info("valid input not attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	// determine user ID.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	requester := reqCtx.Requester.ID
	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.RequesterIDKey, requester)

	accountID := s.accountIDFetcher(req)
	tracing.AttachAccountIDToSpan(span, accountID)
	logger = logger.WithValue(keys.AccountIDKey, accountID)

	// create account in database.
	if err = s.accountMembershipDataManager.AddUserToAccount(ctx, input, requester); err != nil {
		observability.AcknowledgeError(err, logger, span, "adding user to account")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	res.WriteHeader(http.StatusAccepted)
}

// ModifyMemberPermissionsHandler is our account creation route.
func (s *service) ModifyMemberPermissionsHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// check request context for parsed input struct.
	input, ok := ctx.Value(addUserToAccountMiddlewareCtxKey).(*types.ModifyUserPermissionsInput)
	if !ok {
		logger.Info("valid input not attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	// determine user ID.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	requester := reqCtx.Requester.ID
	logger = logger.WithValue(keys.RequesterIDKey, requester)
	tracing.AttachRequestContextToSpan(span, reqCtx)

	accountID := s.accountIDFetcher(req)
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	userID := s.userIDFetcher(req)
	logger = logger.WithValue(keys.UserIDKey, userID)
	tracing.AttachAccountIDToSpan(span, userID)

	// create account in database.
	if err = s.accountMembershipDataManager.ModifyUserPermissions(ctx, accountID, userID, requester, input); err != nil {
		observability.AcknowledgeError(err, logger, span, "modifying user permissions")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	res.WriteHeader(http.StatusAccepted)
}

// TransferAccountOwnershipHandler is our account creation route.
func (s *service) TransferAccountOwnershipHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// check request context for parsed input struct.
	input, ok := ctx.Value(addUserToAccountMiddlewareCtxKey).(*types.TransferAccountOwnershipInput)
	if !ok {
		logger.Info("valid input not attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	accountID := s.accountIDFetcher(req)
	tracing.AttachAccountIDToSpan(span, accountID)
	logger = logger.WithValue(keys.AccountIDKey, accountID)

	// determine user ID.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "transferring account ownership")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	requester := reqCtx.Requester.ID
	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.RequesterIDKey, requester)

	// transfer ownership of account in database.
	if err = s.accountMembershipDataManager.TransferAccountOwnership(ctx, accountID, requester, input); err != nil {
		observability.AcknowledgeError(err, logger, span, "transferring account ownership")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	res.WriteHeader(http.StatusAccepted)
}

// RemoveUserHandler is our account creation route.
func (s *service) RemoveUserHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// check request context for parsed input struct.
	reason := req.URL.Query().Get("reason")
	logger = logger.WithValue(keys.ReasonKey, reason)

	// determine user ID.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	requester := reqCtx.Requester.ID
	logger = logger.WithValue(keys.RequesterIDKey, requester)
	tracing.AttachRequestContextToSpan(span, reqCtx)

	accountID := s.accountIDFetcher(req)
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	userID := s.userIDFetcher(req)
	logger = logger.WithValue(keys.UserIDKey, userID)
	tracing.AttachUserIDToSpan(span, userID)

	// remove user from account in database.
	if err = s.accountMembershipDataManager.RemoveUserFromAccount(ctx, userID, accountID, requester, reason); err != nil {
		observability.AcknowledgeError(err, logger, span, "removing user from account")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	res.WriteHeader(http.StatusAccepted)
}

// MarkAsDefaultHandler is our account creation route.
func (s *service) MarkAsDefaultHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	accountID := s.accountIDFetcher(req)
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	// determine user ID.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	requester := reqCtx.Requester.ID
	logger = logger.WithValue(keys.RequesterIDKey, requester)
	tracing.AttachRequestContextToSpan(span, reqCtx)

	// mark account as default in database.
	if err = s.accountMembershipDataManager.MarkAccountAsUserDefault(ctx, requester, accountID, requester); err != nil {
		observability.AcknowledgeError(err, logger, span, "marking account as default")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	res.WriteHeader(http.StatusAccepted)
}

// AuditEntryHandler returns a GET handler that returns all audit log entries related to an account.
func (s *service) AuditEntryHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// determine user ID.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	requester := reqCtx.Requester.ID
	logger = logger.WithValue(keys.RequesterIDKey, requester)
	tracing.AttachRequestContextToSpan(span, reqCtx)

	// determine account ID.
	accountID := s.accountIDFetcher(req)
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	x, err := s.accountDataManager.GetAuditLogEntriesForAccount(ctx, accountID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	}

	if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching audit log entries")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, x)
}
