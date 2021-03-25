package apiclients

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.APIClientDataService = (*service)(nil)

const (
	// APIClientIDURIParamKey is used for referring to API client IDs in router params.
	APIClientIDURIParamKey = "apiClientID"

	clientIDSize                      = 32
	clientSecretSize                  = 128
	clientIDKey      types.ContextKey = "client_id"
)

// ListHandler is a handler that returns a list of API clients.
func (s *service) ListHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	filter := types.ExtractQueryFilter(req)
	logger := s.logger.WithRequest(req).
		WithValue(keys.FilterLimitKey, filter.Limit).
		WithValue(keys.FilterPageKey, filter.Page).
		WithValue(keys.FilterSortByKey, string(filter.SortBy))

	tracing.AttachFilterToSpan(span, filter.Page, filter.Limit, string(filter.SortBy))

	// determine user.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving request context")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	requester := reqCtx.User.ID
	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.UserIDKey, requester)

	// fetch API clients.
	apiClients, err := s.apiClientDataManager.GetAPIClients(ctx, requester, filter)
	if errors.Is(err, sql.ErrNoRows) {
		// just return an empty list if there are no results.
		apiClients = &types.APIClientList{
			Clients: []*types.APIClient{},
		}
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching API clients from database")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, apiClients)
}

// CreateHandler is our API client creation route.
func (s *service) CreateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// fetch creation input from request context.
	input, ok := ctx.Value(creationMiddlewareCtxKey).(*types.APICientCreationInput)
	if !ok {
		logger.Info("valid input not attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	// check request context for user ID.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving request context")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// keep relevant data in mind.
	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue("username", input.Username)

	// retrieve user.
	user, err := s.userDataManager.GetUser(ctx, reqCtx.User.ID)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching user")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// tag span since we have the info.
	tracing.AttachUserIDToSpan(span, user.ID)

	// check credentials.
	valid, err := s.authenticator.ValidateLogin(
		ctx,
		user.HashedPassword,
		input.Password,
		user.TwoFactorSecret,
		input.TOTPToken,
		user.Salt,
	)

	if !valid {
		logger.Debug("invalid credentials provided to API client creation route")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "validating user credentials")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// set some data.
	if input.ClientID, err = s.secretGenerator.GenerateClientID(); err != nil {
		observability.AcknowledgeError(err, logger, span, "generating client id")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	if input.ClientSecret, err = s.secretGenerator.GenerateClientSecret(); err != nil {
		observability.AcknowledgeError(err, logger, span, "generating client secret")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	input.BelongsToUser = user.ID

	// create the client.
	client, err := s.apiClientDataManager.CreateAPIClient(ctx, input, user.ID)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "creating API client")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// notify interested parties.
	tracing.AttachAPIClientDatabaseIDToSpan(span, client.ID)
	s.apiClientCounter.Increment(ctx)

	resObj := &types.APIClientCreationResponse{
		ID:           client.ID,
		ClientID:     client.ClientID,
		ClientSecret: base64.RawURLEncoding.EncodeToString(input.ClientSecret),
	}

	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, resObj, http.StatusCreated)
}

// ReadHandler returns a GET handler that returns an item.
func (s *service) ReadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	requester := reqCtx.User.ID
	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.RequesterKey, requester)

	// determine API client ID.
	apiClientID := s.urlClientIDExtractor(req)
	tracing.AttachItemIDToSpan(span, apiClientID)
	logger = logger.WithValue(keys.APIClientDatabaseIDKey, apiClientID)

	// fetch item from database.
	x, err := s.apiClientDataManager.GetAPIClientByDatabaseID(ctx, apiClientID, requester)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching API client from database")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, x)
}

// ArchiveHandler returns a handler that archives an API client.
func (s *service) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	requester := reqCtx.User.ID
	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.RequesterKey, requester)

	// determine API client ID.
	apiClientID := s.urlClientIDExtractor(req)
	logger = logger.WithValue(keys.APIClientDatabaseIDKey, apiClientID)
	tracing.AttachItemIDToSpan(span, apiClientID)

	// archive the API client in the database.
	err = s.apiClientDataManager.ArchiveAPIClient(ctx, apiClientID, reqCtx.ActiveAccountID, requester)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "archiving API client")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// notify relevant parties.
	s.apiClientCounter.Decrement(ctx)

	// encode our response and peace.
	res.WriteHeader(http.StatusNoContent)
}

// AuditEntryHandler returns a GET handler that returns all audit log entries related to an API client.
func (s *service) AuditEntryHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user.
	reqCtx, err := s.requestContextFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "retrieving request context")
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	requester := reqCtx.User.ID
	tracing.AttachRequestContextToSpan(span, reqCtx)
	logger = logger.WithValue(keys.UserIDKey, requester)

	// determine relevant API client ID.
	apiClientID := s.urlClientIDExtractor(req)
	tracing.AttachAPIClientDatabaseIDToSpan(span, apiClientID)
	logger = logger.WithValue(keys.APIClientClientIDKey, apiClientID)

	x, err := s.apiClientDataManager.GetAuditLogEntriesForAPIClient(ctx, apiClientID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching audit log entries for API client")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.RespondWithData(ctx, res, x)
}
