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
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, si.UserID, si.UserIsSiteAdmin)
	logger = logger.WithValue(keys.UserIDKey, si.UserID)

	// determine if it's an admin request
	rawQueryAdminKey := req.URL.Query().Get("admin")
	adminQueryPresent := parseBool(rawQueryAdminKey)
	isAdminRequest := si.UserIsSiteAdmin && adminQueryPresent

	var (
		accounts *types.AccountList
		err      error
	)

	if si.UserIsSiteAdmin && isAdminRequest {
		accounts, err = s.accountDataManager.GetAccountsForAdmin(ctx, filter)
	} else {
		accounts, err = s.accountDataManager.GetAccounts(ctx, si.UserID, filter)
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
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, si.UserID, si.UserIsSiteAdmin)
	logger = logger.WithValue(keys.UserIDKey, si.UserID)
	input.BelongsToUser = si.UserID

	// create account in database.
	x, err := s.accountDataManager.CreateAccount(ctx, input)
	if err != nil {
		logger.Error(err, "error creating account")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	tracing.AttachAccountIDToSpan(span, x.ID)

	// notify relevant parties.

	s.accountCounter.Increment(ctx)
	s.auditLog.LogAccountCreationEvent(ctx, x)

	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, x, http.StatusCreated)
}

// ReadHandler returns a GET handler that returns an account.
func (s *service) ReadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, si.UserID, si.UserIsSiteAdmin)
	logger = logger.WithValue(keys.UserIDKey, si.UserID)

	// determine account ID.
	accountID := s.accountIDFetcher(req)
	tracing.AttachAccountIDToSpan(span, accountID)
	logger = logger.WithValue(keys.AccountIDKey, accountID)

	// fetch account from database.
	x, err := s.accountDataManager.GetAccount(ctx, accountID, si.UserID)
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
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, si.UserID, si.UserIsSiteAdmin)
	logger = logger.WithValue(keys.UserIDKey, si.UserID)
	input.BelongsToUser = si.UserID

	// determine account ID.
	accountID := s.accountIDFetcher(req)
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	// fetch account from database.
	x, err := s.accountDataManager.GetAccount(ctx, accountID, si.UserID)
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
	if err = s.accountDataManager.UpdateAccount(ctx, x); err != nil {
		logger.Error(err, "error encountered updating account")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	s.auditLog.LogAccountUpdateEvent(ctx, si.UserID, x.ID, changeReport)

	// notify relevant parties.

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(ctx, res, x)
}

// ArchiveHandler returns a handler that archives an account.
func (s *service) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, si.UserID, si.UserIsSiteAdmin)
	logger = logger.WithValue(keys.UserIDKey, si.UserID)

	// determine account ID.
	accountID := s.accountIDFetcher(req)
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	// archive the account in the database.
	err := s.accountDataManager.ArchiveAccount(ctx, accountID, si.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		logger.Error(err, "error encountered deleting account")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	s.auditLog.LogAccountArchiveEvent(ctx, si.UserID, accountID)

	// notify relevant parties.
	s.accountCounter.Decrement(ctx)

	// encode our response and peace.
	res.WriteHeader(http.StatusNoContent)
}

// AuditEntryHandler returns a GET handler that returns all audit log entries related to an account.
func (s *service) AuditEntryHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Debug("AuditEntryHandler invoked")

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, si.UserID, si.UserIsSiteAdmin)
	logger = logger.WithValue(keys.UserIDKey, si.UserID)

	// determine account ID.
	accountID := s.accountIDFetcher(req)
	tracing.AttachAccountIDToSpan(span, accountID)
	logger = logger.WithValue(keys.AccountIDKey, accountID)

	x, err := s.auditLog.GetAuditLogEntriesForAccount(ctx, accountID)
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
