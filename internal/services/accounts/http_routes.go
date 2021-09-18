package accounts

import (
	"database/sql"
	"errors"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/segmentio/ksuid"
)

const (
	// AccountIDURIParamKey is a standard string that we'll use to refer to account IDs with.
	AccountIDURIParamKey = "accountID"
	// UserIDURIParamKey is a standard string that we'll use to refer to user IDs with.
	UserIDURIParamKey = "userID"
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

	tracing.AttachRequestToSpan(span, req)
	tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))

	// fetch session context data
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	requester := sessionCtxData.Requester.UserID
	logger = logger.WithValue(keys.RequesterIDKey, requester)
	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)

	accounts, err := s.accountDataManager.GetAccounts(ctx, requester, filter)
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

	// retrieve session context data.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	// check session context data for parsed input struct.
	input := new(types.AccountCreationInput)
	if err = s.encoderDecoder.DecodeRequest(ctx, req, input); err != nil {
		observability.AcknowledgeError(err, logger, span, "decoding request body")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
		return
	}

	if err = input.ValidateWithContext(ctx); err != nil {
		logger.WithValue(keys.ValidationErrorKey, err).Debug("invalid input attached to request")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
		return
	}

	logger = logger.WithValue(keys.NameKey, input.Name)

	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)
	requester := sessionCtxData.Requester.UserID
	logger = logger.WithValue(keys.RequesterIDKey, requester)

	input.BelongsToUser = requester
	input.ID = ksuid.New().String()

	// create account in database.
	account, err := s.accountDataManager.CreateAccount(ctx, input)
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
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	requester := sessionCtxData.Requester.UserID
	logger = logger.WithValue(keys.RequesterIDKey, requester)
	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)

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

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	input := new(types.AccountUpdateInput)
	if err = s.encoderDecoder.DecodeRequest(ctx, req, input); err != nil {
		observability.AcknowledgeError(err, logger, span, "decoding request body")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
		return
	}

	if err = input.ValidateWithContext(ctx); err != nil {
		logger.WithValue(keys.ValidationErrorKey, err).Debug("invalid input attached to request")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
		return
	}

	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)
	requester := sessionCtxData.Requester.UserID
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
	account.Update(input)

	// update account in database.
	if err = s.accountDataManager.UpdateAccount(ctx, account); err != nil {
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
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	requester := sessionCtxData.Requester.UserID
	logger = logger.WithValue(keys.RequesterIDKey, requester)
	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)

	// determine account ID.
	accountID := s.accountIDFetcher(req)
	tracing.AttachAccountIDToSpan(span, accountID)
	logger = logger.WithValue(keys.AccountIDKey, accountID)

	// archive the account in the database.
	err = s.accountDataManager.ArchiveAccount(ctx, accountID, requester)
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

// AddMemberHandler is our account creation route.
func (s *service) AddMemberHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	// check session context data for parsed input struct.
	input := new(types.AddUserToAccountInput)
	if err = s.encoderDecoder.DecodeRequest(ctx, req, input); err != nil {
		observability.AcknowledgeError(err, logger, span, "decoding request body")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
		return
	}
	input.ID = ksuid.New().String()

	if err = input.ValidateWithContext(ctx); err != nil {
		logger.WithValue(keys.ValidationErrorKey, err).Debug("invalid input attached to request")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
		return
	}

	requester := sessionCtxData.Requester.UserID
	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)
	logger = logger.WithValue(keys.RequesterIDKey, requester)

	accountID := s.accountIDFetcher(req)
	tracing.AttachAccountIDToSpan(span, accountID)
	logger = logger.WithValue(keys.AccountIDKey, accountID)

	// create account in database.
	if err = s.accountMembershipDataManager.AddUserToAccount(ctx, input); err != nil {
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

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	// check session context data for parsed input struct.
	input := new(types.ModifyUserPermissionsInput)
	if err = s.encoderDecoder.DecodeRequest(ctx, req, input); err != nil {
		observability.AcknowledgeError(err, logger, span, "decoding request body")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
		return
	}

	if err = input.ValidateWithContext(ctx); err != nil {
		logger.WithValue(keys.ValidationErrorKey, err).Debug("invalid input attached to request")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
		return
	}

	requester := sessionCtxData.Requester.UserID
	logger = logger.WithValue(keys.RequesterIDKey, requester)
	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)

	accountID := s.accountIDFetcher(req)
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	userID := s.userIDFetcher(req)
	logger = logger.WithValue(keys.UserIDKey, userID)
	tracing.AttachUserIDToSpan(span, userID)

	// create account in database.
	if err = s.accountMembershipDataManager.ModifyUserPermissions(ctx, accountID, userID, input); err != nil {
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

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "transferring account ownership")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	// check session context data for parsed input struct.
	input := new(types.AccountOwnershipTransferInput)
	if err = s.encoderDecoder.DecodeRequest(ctx, req, input); err != nil {
		observability.AcknowledgeError(err, logger, span, "decoding request body")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "invalid request content", http.StatusBadRequest)
		return
	}

	if err = input.ValidateWithContext(ctx); err != nil {
		logger.WithValue(keys.ValidationErrorKey, err).Debug("invalid input attached to request")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, err.Error(), http.StatusBadRequest)
		return
	}

	accountID := s.accountIDFetcher(req)
	tracing.AttachAccountIDToSpan(span, accountID)
	logger = logger.WithValue(keys.AccountIDKey, accountID)

	requester := sessionCtxData.Requester.UserID
	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)
	logger = logger.WithValue(keys.RequesterIDKey, requester)

	// transfer ownership of account in database.
	if err = s.accountMembershipDataManager.TransferAccountOwnership(ctx, accountID, input); err != nil {
		observability.AcknowledgeError(err, logger, span, "transferring account ownership")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	res.WriteHeader(http.StatusAccepted)
}

// RemoveMemberHandler is our account creation route.
func (s *service) RemoveMemberHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// check session context data for parsed input struct.
	reason := req.URL.Query().Get("reason")
	logger = logger.WithValue(keys.ReasonKey, reason)

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	requester := sessionCtxData.Requester.UserID
	logger = logger.WithValue(keys.RequesterIDKey, requester)
	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)

	accountID := s.accountIDFetcher(req)
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	userID := s.userIDFetcher(req)
	logger = logger.WithValue(keys.UserIDKey, userID)
	tracing.AttachUserIDToSpan(span, userID)

	// remove user from account in database.
	if err = s.accountMembershipDataManager.RemoveUserFromAccount(ctx, userID, accountID); err != nil {
		observability.AcknowledgeError(err, logger, span, "removing user from account")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	res.WriteHeader(http.StatusAccepted)
}

// MarkAsDefaultAccountHandler is our account creation route.
func (s *service) MarkAsDefaultAccountHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	accountID := s.accountIDFetcher(req)
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachAccountIDToSpan(span, accountID)

	// determine user ID.
	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving session context data")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	requester := sessionCtxData.Requester.UserID
	logger = logger.WithValue(keys.RequesterIDKey, requester)
	tracing.AttachSessionContextDataToSpan(span, sessionCtxData)

	// mark account as default in database.
	if err = s.accountMembershipDataManager.MarkAccountAsUserDefault(ctx, requester, accountID); err != nil {
		observability.AcknowledgeError(err, logger, span, "marking account as default")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	res.WriteHeader(http.StatusAccepted)
}
