package oauth2clients

import (
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"errors"
	"net/http"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	// OAuth2ClientIDURIParamKey is used for referring to OAuth2 client IDs in router params.
	OAuth2ClientIDURIParamKey = "oauth2ClientID"

	oauth2ClientIDURIParamKey                   = "client_id"
	clientIDKey               models.ContextKey = "client_id"
)

// randString produces a random string.
// https://blog.questionable.services/article/generating-secure-random-numbers-crypto-rand/
func randString() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	// this is so that we don't end up with `=` in IDs
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
}

// fetchUserID grabs a userID out of the request context.
func (s *Service) fetchUserID(req *http.Request) uint64 {
	if si, ok := req.Context().Value(models.SessionInfoKey).(*models.SessionInfo); ok && si != nil {
		return si.UserID
	}
	return 0
}

// ListHandler is a handler that returns a list of OAuth2 clients.
func (s *Service) ListHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "ListHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// extract filter.
	filter := models.ExtractQueryFilter(req)

	// determine user.
	userID := s.fetchUserID(req)
	tracing.AttachUserIDToSpan(span, userID)
	logger = logger.WithValue("user_id", userID)

	// fetch oauth2 clients.
	oauth2Clients, err := s.clientDataManager.GetOAuth2ClientsForUser(ctx, userID, filter)
	if errors.Is(err, sql.ErrNoRows) {
		// just return an empty list if there are no results.
		oauth2Clients = &models.OAuth2ClientList{
			Clients: []models.OAuth2Client{},
		}
	} else if err != nil {
		logger.Error(err, "encountered error getting list of oauth2 clients from clientDataManager")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		return
	}

	// encode response and peace.
	s.encoderDecoder.EncodeResponse(res, oauth2Clients)
}

// CreateHandler is our OAuth2 client creation route.
func (s *Service) CreateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "CreateHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// fetch creation input from request context.
	input, ok := ctx.Value(creationMiddlewareCtxKey).(*models.OAuth2ClientCreationInput)
	if !ok {
		logger.Info("valid input not attached to request")
		s.encoderDecoder.EncodeNoInputResponse(res)
		return
	}

	// set some data.
	input.ClientID, input.ClientSecret = randString(), randString()
	input.BelongsToUser = s.fetchUserID(req)

	// keep relevant data in mind.
	logger = logger.WithValues(map[string]interface{}{
		"username":     input.Username,
		"scopes":       strings.Join(input.Scopes, scopesSeparator),
		"redirect_uri": input.RedirectURI,
	})

	// retrieve user.
	user, err := s.userDataManager.GetUserByUsername(ctx, input.Username)
	if err != nil {
		logger.Error(err, "fetching user by username")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
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
		s.encoderDecoder.EncodeUnauthorizedResponse(res)
		return
	} else if err != nil {
		logger.Error(err, "validating user credentials")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		return
	}

	// create the client.
	client, err := s.clientDataManager.CreateOAuth2Client(ctx, input)
	if err != nil {
		logger.Error(err, "creating oauth2Client in the clientDataManager")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		return
	}

	// notify interested parties.
	tracing.AttachOAuth2ClientDatabaseIDToSpan(span, client.ID)
	s.oauth2ClientCounter.Increment(ctx)
	s.auditLog.LogOAuth2ClientCreationEvent(ctx, client)

	s.encoderDecoder.EncodeResponseWithStatus(res, client, http.StatusCreated)
}

// ReadHandler is a route handler for retrieving an OAuth2 client.
func (s *Service) ReadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "ReadHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine subject of request.
	userID := s.fetchUserID(req)
	tracing.AttachUserIDToSpan(span, userID)
	logger = logger.WithValue("user_id", userID)

	// determine relevant oauth2 client ID.
	oauth2ClientID := s.urlClientIDExtractor(req)
	tracing.AttachOAuth2ClientDatabaseIDToSpan(span, oauth2ClientID)
	logger = logger.WithValue("oauth2_client_id", oauth2ClientID)

	// fetch oauth2 client.
	x, err := s.clientDataManager.GetOAuth2Client(ctx, oauth2ClientID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Debug("ReadHandler called on nonexistent client")
		s.encoderDecoder.EncodeNotFoundResponse(res)
		return
	} else if err != nil {
		logger.Error(err, "error fetching oauth2Client from clientDataManager")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		return
	}

	// encode response and peace.
	s.encoderDecoder.EncodeResponse(res, x)
}

// ArchiveHandler is a route handler for archiving an OAuth2 client.
func (s *Service) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "ArchiveHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine subject of request.
	userID := s.fetchUserID(req)
	tracing.AttachUserIDToSpan(span, userID)
	logger = logger.WithValue("user_id", userID)

	// determine relevant oauth2 client ID.
	oauth2ClientID := s.urlClientIDExtractor(req)
	tracing.AttachOAuth2ClientDatabaseIDToSpan(span, oauth2ClientID)
	logger = logger.WithValue("oauth2_client_id", oauth2ClientID)

	// mark client as archived.
	err := s.clientDataManager.ArchiveOAuth2Client(ctx, oauth2ClientID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(res)
		return
	} else if err != nil {
		logger.Error(err, "encountered error deleting oauth2 client")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		return
	}

	// notify relevant parties.
	s.oauth2ClientCounter.Decrement(ctx)
	s.auditLog.LogOAuth2ClientArchiveEvent(ctx, userID, oauth2ClientID)

	res.WriteHeader(http.StatusNoContent)
}
