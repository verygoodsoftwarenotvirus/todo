package oauth2clients

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// OAuth2ClientIDURIParamKey is used for referring to OAuth2 client IDs in router params.
	OAuth2ClientIDURIParamKey = "oauth2ClientID"

	oauth2ClientIDURIParamKey                  = "client_id"
	clientIDKey               types.ContextKey = "client_id"
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
func (s *service) fetchUserID(req *http.Request) uint64 {
	if si, ok := req.Context().Value(types.SessionInfoKey).(*types.SessionInfo); ok && si != nil {
		return si.UserID
	}
	return 0
}

// determineScope determines the scope of a request by its url.
func determineScope(req *http.Request) string {
	_, scope := filepath.Split(req.URL.Path)
	return scope
}

// ExtractOAuth2ClientFromRequest extracts OAuth2 client data from a request.
func (s *service) ExtractOAuth2ClientFromRequest(ctx context.Context, req *http.Request) (*types.OAuth2Client, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithRequest(req)

	// validate bearer token.
	token, err := s.oauth2Handler.ValidationBearerToken(req)
	if err != nil {
		return nil, fmt.Errorf("validating bearer token: %w", err)
	}

	// fetch client ID.
	clientID := token.GetClientID()
	logger = logger.WithValue(keys.OAuth2ClientIDKey, clientID)

	// fetch client by client ID.
	c, err := s.clientDataManager.GetOAuth2ClientByClientID(ctx, clientID)
	if err != nil {
		logger.Error(err, "error fetching OAuth2 Client")
		return nil, err
	}

	// determine the scope.
	scope := determineScope(req)
	hasScope := c.HasScope(scope)
	logger = logger.WithValue("scope", scope).WithValue("scopes", strings.Join(c.Scopes, scopesSeparator))

	if !hasScope {
		logger.Info("rejecting client for invalid scope")
		return nil, errClientUnauthorizedForScope
	}

	return c, nil
}

// ListHandler is a handler that returns a list of OAuth2 clients.
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

	// fetch oauth2 clients.
	oauth2Clients, err := s.clientDataManager.GetOAuth2Clients(ctx, userID, filter)
	if errors.Is(err, sql.ErrNoRows) {
		// just return an empty list if there are no results.
		oauth2Clients = &types.OAuth2ClientList{
			Clients: []*types.OAuth2Client{},
		}
	} else if err != nil {
		logger.Error(err, "encountered error getting list of oauth2 clients from clientDataManager")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode response and peace.
	s.encoderDecoder.EncodeResponse(ctx, res, oauth2Clients)
}

// CreateHandler is our OAuth2 client creation route.
func (s *service) CreateHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// fetch creation input from request context.
	input, ok := ctx.Value(creationMiddlewareCtxKey).(*types.OAuth2ClientCreationInput)
	if !ok {
		logger.Info("valid input not attached to request")
		s.encoderDecoder.EncodeInvalidInputResponse(ctx, res)
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

	// create the client.
	client, err := s.clientDataManager.CreateOAuth2Client(ctx, input)
	if err != nil {
		logger.Error(err, "creating oauth2Client in the clientDataManager")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// notify interested parties.
	tracing.AttachOAuth2ClientDatabaseIDToSpan(span, client.ID)
	s.oauth2ClientCounter.Increment(ctx)

	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, client, http.StatusCreated)
}

// ReadHandler is a route handler for retrieving an OAuth2 client.
func (s *service) ReadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine subject of request.
	userID := s.fetchUserID(req)
	tracing.AttachUserIDToSpan(span, userID)
	logger = logger.WithValue(keys.UserIDKey, userID)

	// determine relevant oauth2 client ID.
	oauth2ClientID := s.urlClientIDExtractor(req)
	tracing.AttachOAuth2ClientDatabaseIDToSpan(span, oauth2ClientID)
	logger = logger.WithValue(keys.OAuth2ClientDatabaseIDKey, oauth2ClientID)

	// fetch oauth2 client.
	x, err := s.clientDataManager.GetOAuth2Client(ctx, oauth2ClientID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Debug("ReadHandler called on nonexistent client")
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		logger.Error(err, "error fetching oauth2Client from clientDataManager")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// encode response and peace.
	s.encoderDecoder.EncodeResponse(ctx, res, x)
}

// ArchiveHandler is a route handler for archiving an OAuth2 client.
func (s *service) ArchiveHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine subject of request.
	userID := s.fetchUserID(req)
	tracing.AttachUserIDToSpan(span, userID)
	logger = logger.WithValue(keys.UserIDKey, userID)

	// determine relevant oauth2 client ID.
	oauth2ClientID := s.urlClientIDExtractor(req)
	tracing.AttachOAuth2ClientDatabaseIDToSpan(span, oauth2ClientID)
	logger = logger.WithValue(keys.OAuth2ClientDatabaseIDKey, oauth2ClientID)

	// mark client as archived.
	err := s.clientDataManager.ArchiveOAuth2Client(ctx, oauth2ClientID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(ctx, res)
		return
	} else if err != nil {
		logger.Error(err, "encountered error deleting oauth2 client")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
		return
	}

	// notify relevant parties.
	s.oauth2ClientCounter.Decrement(ctx)

	res.WriteHeader(http.StatusNoContent)
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

	// determine relevant oauth2 client ID.
	oauth2ClientID := s.urlClientIDExtractor(req)
	tracing.AttachOAuth2ClientDatabaseIDToSpan(span, oauth2ClientID)
	logger = logger.WithValue(keys.OAuth2ClientDatabaseIDKey, oauth2ClientID)

	x, err := s.clientDataManager.GetAuditLogEntriesForOAuth2Client(ctx, oauth2ClientID)
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
