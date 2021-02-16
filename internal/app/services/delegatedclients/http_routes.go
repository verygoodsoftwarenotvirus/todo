package delegatedclients

import (
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"errors"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// DelegatedClientIDURIParamKey is used for referring to Delegated client IDs in router params.
	DelegatedClientIDURIParamKey = "delegatedClientID"

	clientIDKey types.ContextKey = "client_id"
)

// randString produces a random string.
// https://blog.questionable.services/article/generating-secure-random-numbers-crypto-rand/
func buildClientID() string {
	b := randByteArrayOfLength(32)

	// this is so that we don't end up with `=` in IDs
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
}

// randString produces a random string.
// https://blog.questionable.services/article/generating-secure-random-numbers-crypto-rand/
func buildClientSecret() []byte {
	return randByteArrayOfLength(128)
}

func randByteArrayOfLength(length uint) []byte {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return b
}

// fetchUserID grabs a userID out of the request context.
func (s *service) fetchUserID(req *http.Request) uint64 {
	if si, ok := req.Context().Value(types.SessionInfoKey).(*types.SessionInfo); ok && si != nil {
		return si.UserID
	}
	return 0
}

// ListHandler is a handler that returns a list of Delegated clients.
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

	// fetch delegated clients.
	delegatedClients, err := s.clientDataManager.GetDelegatedClients(ctx, userID, filter)
	if errors.Is(err, sql.ErrNoRows) {
		// just return an empty list if there are no results.
		delegatedClients = &types.DelegatedClientList{
			Clients: []*types.DelegatedClient{},
		}
	} else if err != nil {
		logger.Error(err, "encountered error getting list of delegated clients from clientDataManager")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode response and peace.
	s.encoderDecoder.EncodeResponse(ctx, res, delegatedClients)
}

// CreateHandler is our Delegated client creation route.
func (s *service) CreateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// fetch creation input from request context.
	input, ok := ctx.Value(creationMiddlewareCtxKey).(*types.DelegatedClientCreationInput)
	if !ok {
		logger.Info("valid input not attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
		return
	}

	// keep relevant data in mind.
	logger = logger.WithValues(map[string]interface{}{
		"username": input.Username,
	})

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
	client, err := s.clientDataManager.CreateDelegatedClient(ctx, input)
	if err != nil {
		logger.Error(err, "creating delegatedClient in the clientDataManager")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// notify interested parties.
	tracing.AttachDelegatedClientDatabaseIDToSpan(span, client.ID)
	s.delegatedClientCounter.Increment(ctx)

	resObj := &types.DelegatedClientCreationResponse{
		ClientID:     client.ClientID,
		ClientSecret: client.ClientSecret,
	}

	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, resObj, http.StatusCreated)
}

// AuditEntryHandler returns a GET handler that returns all audit log entries related to an item.
func (s *service) AuditEntryHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Debug("AuditEntryHandler invoked")

	userID := s.fetchUserID(req)
	tracing.AttachUserIDToSpan(span, userID)
	logger = logger.WithValue(keys.UserIDKey, userID)

	// determine relevant delegated client ID.
	delegatedClientID := s.urlClientIDExtractor(req)
	tracing.AttachDelegatedClientDatabaseIDToSpan(span, delegatedClientID)
	logger = logger.WithValue(keys.DelegatedClientIDKey, delegatedClientID)

	x, err := s.clientDataManager.GetAuditLogEntriesForDelegatedClient(ctx, delegatedClientID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		logger.Error(err, "error encountered fetching items")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	logger.WithValue("entry_count", len(x)).Debug("returning from AuditEntryHandler")

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(ctx, res, x)
}
