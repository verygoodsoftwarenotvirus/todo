package accounts

import (
	"database/sql"
	"errors"
	"net/http"

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

	logger := s.logger.WithRequest(req)
	logger.Debug("ListHandler invoked")

	// ensure query filter.
	filter := types.ExtractQueryFilter(req)

	// determine user ID.
	reqCtx, sessionInfoRetrievalErr := s.requestContextFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.UserIDKey, reqCtx.User.ID)

	// determine if it's an admin request
	rawQueryAdminKey := req.URL.Query().Get("admin")
	adminQueryPresent := parseBool(rawQueryAdminKey)
	isAdminRequest := reqCtx.User.ServiceAdminPermissions.IsServiceAdmin() && adminQueryPresent

	var (
		accounts *types.AccountList
		err      error
	)

	if reqCtx.User.ServiceAdminPermissions.IsServiceAdmin() && isAdminRequest {
		accounts, err = s.accountDataManager.GetAccountsForAdmin(ctx, filter)
	} else {
		accounts, err = s.accountDataManager.GetAccounts(ctx, reqCtx.User.ID, filter)
	}

	if errors.Is(err, sql.ErrNoRows) {
		// in the event no rows exist return an empty list.
		accounts = &types.AccountList{Accounts: []*types.Account{}}
	} else if err != nil {
		logger.Error(err, "error encountered fetching accounts")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(ctx, res, accounts)
}

// CreateHandler is our account creation route.
func (s *service) CreateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// check request context for parsed input struct.
	input, ok := ctx.Value(createMiddlewareCtxKey).(*types.AccountCreationInput)
	if !ok {
		logger.Info("valid input not attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	// determine user ID.
	reqCtx, requestContextRetrievalError := s.requestContextFetcher(req)
	if requestContextRetrievalError != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.UserIDKey, reqCtx.User.ID)
	input.BelongsToUser = reqCtx.User.ID

	// create account in database.
	x, err := s.accountDataManager.CreateAccount(ctx, input, reqCtx.User.ID)
	if err != nil {
		logger.Error(err, "error creating account")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	tracing.AttachAccountIDToSpan(span, x.ID)

	// notify relevant parties.

	s.accountCounter.Increment(ctx)

	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, x, http.StatusCreated)
}

// ReadHandler returns a GET handler that returns an account.
func (s *service) ReadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	reqCtx, sessionInfoRetrievalErr := s.requestContextFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.UserIDKey, reqCtx.User.ID)

	// determine account ID.
	accountID := s.accountIDFetcher(req)
	tracing.AttachAccountIDToSpan(span, accountID)
	logger = logger.WithValue(keys.AccountIDKey, accountID)

	// fetch account from database.
	x, err := s.accountDataManager.GetAccount(ctx, accountID, reqCtx.User.ID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		logger.Error(err, "error fetching account from database")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(ctx, res, x)
}

// UpdateHandler returns a handler that updates an account.
func (s *service) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// check for parsed input attached to request context.
	input, ok := ctx.Value(updateMiddlewareCtxKey).(*types.AccountUpdateInput)
	if !ok {
		logger.Info("no input attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	// determine user ID.
	reqCtx, requestContextRetrievalError := s.requestContextFetcher(req)
	if requestContextRetrievalError != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.UserIDKey, reqCtx.User.ID)
	input.BelongsToUser = reqCtx.User.ID

	// determine account ID.
	accountID := s.accountIDFetcher(req)
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	// fetch account from database.
	x, err := s.accountDataManager.GetAccount(ctx, accountID, reqCtx.User.ID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		logger.Error(err, "error encountered getting account")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// update the data structure.
	changeReport := x.Update(input)

	// update account in database.
	if err = s.accountDataManager.UpdateAccount(ctx, x, 0, changeReport); err != nil {
		logger.Error(err, "error encountered updating account")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(ctx, res, x)
}

// ArchiveHandler returns a handler that archives an account.
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

	// determine account ID.
	accountID := s.accountIDFetcher(req)
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	// archive the account in the database.
	err := s.accountDataManager.ArchiveAccount(ctx, accountID, reqCtx.User.ID, reqCtx.User.ID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		logger.Error(err, "error encountered deleting account")
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

	// check request context for parsed input struct.
	input, ok := ctx.Value(addUserToAccountMiddlewareCtxKey).(*types.AddUserToAccountInput)
	if !ok {
		logger.Info("valid input not attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	accountID := s.accountIDFetcher(req)
	tracing.AttachAccountIDToSpan(span, accountID)
	logger = logger.WithValue(keys.AccountIDKey, accountID)

	// determine user ID.
	reqCtx, requestContextRetrievalError := s.requestContextFetcher(req)
	if requestContextRetrievalError != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.UserIDKey, reqCtx.User.ID)

	// create account in database.
	if err := s.accountMembershipDataManager.AddUserToAccount(ctx, input, accountID, reqCtx.User.ID); err != nil {
		logger.Error(err, "error creating account")
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

	// check request context for parsed input struct.
	input, ok := ctx.Value(addUserToAccountMiddlewareCtxKey).(*types.ModifyUserPermissionsInput)
	if !ok {
		logger.Info("valid input not attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	accountID := s.accountIDFetcher(req)
	tracing.AttachAccountIDToSpan(span, accountID)
	logger = logger.WithValue(keys.AccountIDKey, accountID)

	userID := s.userIDFetcher(req)
	tracing.AttachAccountIDToSpan(span, userID)
	logger = logger.WithValue(keys.UserIDKey, userID)

	// determine user ID.
	reqCtx, requestContextRetrievalError := s.requestContextFetcher(req)
	if requestContextRetrievalError != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.UserIDKey, reqCtx.User.ID)

	// check if requesting user is authorized to make the change

	// create account in database.
	if err := s.accountMembershipDataManager.ModifyUserPermissions(ctx, accountID, userID, reqCtx.User.ID, input); err != nil {
		logger.Error(err, "error creating account")
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
	reqCtx, requestContextRetrievalError := s.requestContextFetcher(req)
	if requestContextRetrievalError != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.UserIDKey, reqCtx.User.ID)

	// transfer ownership of account in database.
	if err := s.accountMembershipDataManager.TransferAccountOwnership(ctx, accountID, reqCtx.User.ID, input); err != nil {
		logger.Error(err, "error creating account")
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

	// check request context for parsed input struct.
	reason := req.URL.Query().Get("reason")

	accountID := s.accountIDFetcher(req)
	tracing.AttachAccountIDToSpan(span, accountID)
	logger = logger.WithValue(keys.AccountIDKey, accountID)

	userID := s.userIDFetcher(req)
	tracing.AttachUserIDToSpan(span, userID)
	logger = logger.WithValue(keys.AccountIDKey, userID)

	// determine user ID.
	reqCtx, requestContextRetrievalError := s.requestContextFetcher(req)
	if requestContextRetrievalError != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.UserIDKey, reqCtx.User.ID)

	// remove user from account in database.
	if err := s.accountMembershipDataManager.RemoveUserFromAccount(ctx, userID, accountID, reqCtx.User.ID, reason); err != nil {
		logger.Error(err, "error creating account")
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

	accountID := s.accountIDFetcher(req)
	tracing.AttachAccountIDToSpan(span, accountID)
	logger = logger.WithValue(keys.AccountIDKey, accountID)

	// determine user ID.
	reqCtx, requestContextRetrievalError := s.requestContextFetcher(req)
	if requestContextRetrievalError != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.UserIDKey, reqCtx.User.ID)

	// mark account as default in database.
	err := s.accountMembershipDataManager.MarkAccountAsUserDefault(ctx, reqCtx.User.ID, accountID, reqCtx.User.ID)
	if err != nil {
		logger.Error(err, "error marking account as default")
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

	// determine account ID.
	accountID := s.accountIDFetcher(req)
	tracing.AttachAccountIDToSpan(span, accountID)
	logger = logger.WithValue(keys.AccountIDKey, accountID)

	x, err := s.accountDataManager.GetAuditLogEntriesForAccount(ctx, accountID)
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
	s.encoderDecoder.EncodeResponse(ctx, res, x)
}
