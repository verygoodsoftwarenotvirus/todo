package apiclients

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.APIClientDataService = (*service)(nil)

const (
	// APIClientIDURIParamKey is used for referring to API client IDs in router params.
	APIClientIDURIParamKey = "apiClientID"

	clientIDKey types.ContextKey = "client_id"

	clientIDSize     = 32
	clientSecretSize = 128
)

// fetchUserID grabs a userID out of the request context.
func (s *service) fetchUserID(req *http.Request) uint64 {
	if si, ok := req.Context().Value(types.RequestContextKey).(*types.RequestContext); ok && si != nil {
		return si.User.ID
	}
	return 0
}

// ListHandler is a handler that returns a list of API clients.
func (s *service) ListHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// extract filter.
	filter := types.ExtractQueryFilter(req)

	// determine user.
	userID := s.fetchUserID(req)
	tracing.AttachUserIDToSpan(span, userID)
	logger = logger.WithValue(keys.UserIDKey, userID)

	// fetch API clients.
	apiClients, err := s.apiClientDataManager.GetAPIClients(ctx, userID, filter)
	if errors.Is(err, sql.ErrNoRows) {
		// just return an empty list if there are no results.
		apiClients = &types.APIClientList{
			Clients: []*types.APIClient{},
		}
	} else if err != nil {
		logger.Error(err, "encountered error getting list of API clients from apiClientDataManager")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode response and peace.
	s.encoderDecoder.EncodeResponse(ctx, res, apiClients)
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

	// keep relevant data in mind.
	logger = logger.WithValue("username", input.Username)

	// retrieve user.
	user, err := s.userDataManager.GetUserByUsername(ctx, input.Username)
	if err != nil {
		logger.Error(err, "fetching user by username")
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
		logger.Debug("invalid credentials provided")
		s.encoderDecoder.EncodeUnauthorizedResponse(ctx, res)
		return
	} else if err != nil {
		logger.Error(err, "validating user credentials")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// set some data.
	if input.ClientID, err = s.secretGenerator.GenerateClientID(); err != nil {
		logger.Error(err, "generating client id")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	if input.ClientSecret, err = s.secretGenerator.GenerateClientSecret(); err != nil {
		logger.Error(err, "generating client secret")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	input.BelongsToUser = s.fetchUserID(req)
	input.ServiceAdminPermissions = user.ServiceAdminPermissions

	// create the client.
	client, err := s.apiClientDataManager.CreateAPIClient(ctx, input)
	if err != nil {
		logger.Error(err, "creating API client")
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
	si, sessionInfoRetrievalErr := s.requestContextFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, si)
	logger = logger.WithValue(keys.UserIDKey, si.User.ID)

	// determine API client ID.
	apiClientID := s.urlClientIDExtractor(req)
	tracing.AttachItemIDToSpan(span, apiClientID)
	logger = logger.WithValue(keys.APIClientDatabaseIDKey, apiClientID)

	// fetch item from database.
	x, err := s.apiClientDataManager.GetAPIClientByDatabaseID(ctx, apiClientID, si.User.ID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		logger.Error(err, "error fetching item from database")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(ctx, res, x)
}

// ArchiveHandler returns a handler that archives an API client.
func (s *service) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.requestContextFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(ctx, res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachRequestContextToSpan(span, si)
	logger = logger.WithValue(keys.UserIDKey, si.User.ID)

	// determine API client ID.
	apiClientID := s.urlClientIDExtractor(req)
	logger = logger.WithValue(keys.APIClientDatabaseIDKey, apiClientID)
	tracing.AttachItemIDToSpan(span, apiClientID)

	// archive the API client in the database.
	err := s.apiClientDataManager.ArchiveAPIClient(ctx, apiClientID, si.User.ID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		logger.Error(err, "error encountered deleting API client")
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
	logger.Debug("AuditEntryHandler invoked")

	userID := s.fetchUserID(req)
	tracing.AttachUserIDToSpan(span, userID)
	logger = logger.WithValue(keys.UserIDKey, userID)

	// determine relevant API client ID.
	apiClientID := s.urlClientIDExtractor(req)
	tracing.AttachAPIClientDatabaseIDToSpan(span, apiClientID)
	logger = logger.WithValue(keys.APIClientClientIDKey, apiClientID)

	x, err := s.apiClientDataManager.GetAuditLogEntriesForAPIClient(ctx, apiClientID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		logger.Error(err, "error encountered fetching API clients")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	logger.WithValue("entry_count", len(x)).Debug("returning from AuditEntryHandler")

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(ctx, res, x)
}
