package oauth2clients

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/go-oauth2/oauth2/v4"
	oauth2errors "github.com/go-oauth2/oauth2/v4/errors"
	oauth2server "github.com/go-oauth2/oauth2/v4/server"
)

// github.com/go-oauth2/oauth2/v4/server specific implementations

var _ oauth2server.InternalErrorHandler = (*service)(nil).OAuth2InternalErrorHandler

// OAuth2InternalErrorHandler fulfills a role for the OAuth2 server-side provider.
func (s *service) OAuth2InternalErrorHandler(err error) *oauth2errors.Response {
	s.logger.Error(err, "OAuth2 Internal Error")

	res := &oauth2errors.Response{
		Error:       err,
		Description: "Internal error",
		ErrorCode:   http.StatusInternalServerError,
		StatusCode:  http.StatusInternalServerError,
	}

	return res
}

var _ oauth2server.ResponseErrorHandler = (*service)(nil).OAuth2ResponseErrorHandler

// OAuth2ResponseErrorHandler fulfills a role for the OAuth2 server-side provider.
func (s *service) OAuth2ResponseErrorHandler(re *oauth2errors.Response) {
	s.logger.WithValues(map[string]interface{}{
		"error_code":  re.ErrorCode,
		"description": re.Description,
		"uri":         re.URI,
		"status_code": re.StatusCode,
		"header":      re.Header,
	}).Error(re.Error, "OAuth2ResponseErrorHandler")
}

var _ oauth2server.AuthorizeScopeHandler = (*service)(nil).AuthorizeScopeHandler

var (
	errUnauthorizedForScope = errors.New("not authorized for scope")
	errNoScopeInformation   = errors.New("no scope information found")
)

// AuthorizeScopeHandler satisfies the oauth2server AuthorizeScopeHandler interface.
func (s *service) AuthorizeScopeHandler(res http.ResponseWriter, req *http.Request) (scope string, err error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	scope = determineScope(req)
	logger = logger.WithValue("scope", scope)

	// check for client and return if valid.
	var client = s.fetchOAuth2ClientFromRequest(req)
	if client != nil && client.HasScope(scope) {
		res.WriteHeader(http.StatusOK)
		return scope, nil
	}

	// check to see if the client ID is present instead.
	if clientID := s.fetchOAuth2ClientIDFromRequest(req); clientID != "" {
		// fetch oauth2 client from clientDataManager.
		client, err = s.clientDataManager.GetOAuth2ClientByClientID(ctx, clientID)

		if errors.Is(err, sql.ErrNoRows) {
			logger.Error(err, "error fetching OAuth2 Client")
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "no such oauth2 client", http.StatusUnauthorized)
			return "", err
		} else if err != nil {
			logger.Error(err, "error fetching OAuth2 Client")
			s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)
			return "", err
		}

		// authorization check.
		if !client.HasScope(scope) {
			s.encoderDecoder.EncodeErrorResponse(ctx, res, "not authorized for scope", http.StatusUnauthorized)
			return "", errUnauthorizedForScope
		}

		return scope, nil
	}

	// invalid credentials.
	s.encoderDecoder.EncodeErrorResponse(ctx, res, "no scope information found", http.StatusBadRequest)
	return "", errNoScopeInformation
}

var _ oauth2server.UserAuthorizationHandler = (*service)(nil).UserAuthorizationHandler

// UserAuthorizationHandler satisfies the oauth2server UserAuthorizationHandler interface.
func (s *service) UserAuthorizationHandler(_ http.ResponseWriter, req *http.Request) (userID string, err error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	var uid uint64

	logger := s.logger.WithRequest(req)

	// check context for client.
	if client, clientOk := ctx.Value(types.OAuth2ClientKey).(*types.OAuth2Client); !clientOk {
		// check for user instead.
		si, userOk := ctx.Value(types.SessionInfoKey).(*types.SessionInfo)
		if !userOk || si == nil {
			logger.Debug("no user attached to this request")
			return "", errors.New("user not found")
		}

		if si.UserAccountStatus == types.BannedAccountStatus {
			logger.Debug("banned user making this request")
			return "", errors.New("user is banned")
		}

		uid = si.UserID
	} else {
		uid = client.BelongsToUser
	}

	return strconv.FormatUint(uid, 10), nil
}

var _ oauth2server.ClientAuthorizedHandler = (*service)(nil).ClientAuthorizedHandler

// ClientAuthorizedHandler satisfies the oauth2server ClientAuthorizedHandler interface.
func (s *service) ClientAuthorizedHandler(clientID string, grant oauth2.GrantType) (allowed bool, err error) {
	// NOTE: it's a shame the interface we're implementing doesn't have this as its first argument
	ctx, span := s.tracer.StartSpan(context.Background())
	defer span.End()

	logger := s.logger.WithValues(map[string]interface{}{
		"grant":     grant,
		"client_id": clientID,
	})

	// reject invalid grant type.
	if grant == oauth2.PasswordCredentials {
		return false, errors.New("invalid grant type: password")
	}

	// fetch client data.
	client, err := s.clientDataManager.GetOAuth2ClientByClientID(ctx, clientID)
	if err != nil {
		logger.Error(err, "fetching oauth2 client from clientDataManager")
		return false, fmt.Errorf("fetching oauth2 client from clientDataManager: %w", err)
	}

	// disallow implicit grants unless authorized.
	if grant == oauth2.Implicit && !client.ImplicitAllowed {
		return false, errors.New("client not authorized for implicit grants")
	}

	return true, nil
}

var _ oauth2server.ClientScopeHandler = (*service)(nil).ClientScopeHandler

var errUnauthorized = errors.New("unauthorized")

// ClientScopeHandler satisfies the oauth2server ClientScopeHandler interface.
func (s *service) ClientScopeHandler(clientID, scope string) (authed bool, err error) {
	// NOTE: it's a shame the interface we're implementing doesn't have this as its first argument
	ctx, span := s.tracer.StartSpan(context.Background())
	defer span.End()

	logger := s.logger.WithValues(map[string]interface{}{
		"client_id": clientID,
		"scope":     scope,
	})

	// fetch client info.
	c, err := s.clientDataManager.GetOAuth2ClientByClientID(ctx, clientID)
	if err != nil {
		logger.Error(err, "error fetching OAuth2 client for ClientScopeHandler")
		return false, err
	}

	// check for scope.
	if c.HasScope(scope) {
		return true, nil
	}

	return false, errUnauthorized
}
