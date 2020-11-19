package oauth2clients

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	oauth2 "gopkg.in/oauth2.v3"
	oauth2errors "gopkg.in/oauth2.v3/errors"
	oauth2server "gopkg.in/oauth2.v3/server"
)

// gopkg.in/oauth2.v3/server specific implementations

var _ oauth2server.InternalErrorHandler = (*Service)(nil).OAuth2InternalErrorHandler

// OAuth2InternalErrorHandler fulfills a role for the OAuth2 server-side provider.
func (s *Service) OAuth2InternalErrorHandler(err error) *oauth2errors.Response {
	s.logger.Error(err, "OAuth2 Internal Error")

	res := &oauth2errors.Response{
		Error:       err,
		Description: "Internal error",
		ErrorCode:   http.StatusInternalServerError,
		StatusCode:  http.StatusInternalServerError,
	}

	return res
}

var _ oauth2server.ResponseErrorHandler = (*Service)(nil).OAuth2ResponseErrorHandler

// OAuth2ResponseErrorHandler fulfills a role for the OAuth2 server-side provider.
func (s *Service) OAuth2ResponseErrorHandler(re *oauth2errors.Response) {
	s.logger.WithValues(map[string]interface{}{
		"error_code":  re.ErrorCode,
		"description": re.Description,
		"uri":         re.URI,
		"status_code": re.StatusCode,
		"header":      re.Header,
	}).Error(re.Error, "OAuth2ResponseErrorHandler")
}

var _ oauth2server.AuthorizeScopeHandler = (*Service)(nil).AuthorizeScopeHandler

var (
	errUnauthorizedForScope = errors.New("not authorized for scope")
	errNoScopeInformation   = errors.New("no scope information found")
)

// AuthorizeScopeHandler satisfies the oauth2server AuthorizeScopeHandler interface.
func (s *Service) AuthorizeScopeHandler(res http.ResponseWriter, req *http.Request) (scope string, err error) {
	ctx, span := tracing.StartSpan(req.Context(), "oauth2clients.service.AuthorizeScopeHandler")
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
			s.encoderDecoder.EncodeErrorResponse(res, "no such oauth2 client", http.StatusUnauthorized)
			return "", err
		} else if err != nil {
			logger.Error(err, "error fetching OAuth2 Client")
			s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
			return "", err
		}

		// authorization check.
		if !client.HasScope(scope) {
			s.encoderDecoder.EncodeErrorResponse(res, "not authorized for scope", http.StatusUnauthorized)
			return "", errUnauthorizedForScope
		}

		return scope, nil
	}

	// invalid credentials.
	s.encoderDecoder.EncodeErrorResponse(res, "no scope information found", http.StatusBadRequest)
	return "", errNoScopeInformation
}

var _ oauth2server.UserAuthorizationHandler = (*Service)(nil).UserAuthorizationHandler

// UserAuthorizationHandler satisfies the oauth2server UserAuthorizationHandler interface.
func (s *Service) UserAuthorizationHandler(_ http.ResponseWriter, req *http.Request) (userID string, err error) {
	ctx, span := tracing.StartSpan(req.Context(), "oauth2clients.service.UserAuthorizationHandler")
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

		if si.UserAccountStatus == types.BannedStandingAccountStatus {
			logger.Debug("banned user making this request")
			return "", errors.New("user is banned")
		}

		uid = si.UserID
	} else {
		uid = client.BelongsToUser
	}

	return strconv.FormatUint(uid, 10), nil
}

var _ oauth2server.ClientAuthorizedHandler = (*Service)(nil).ClientAuthorizedHandler

var (
	errInvalidGrantTypePassword = errors.New("invalid grant type: password")
	errImplicitGrantsUnallowed  = errors.New("client not authorized for implicit grants")
)

// ClientAuthorizedHandler satisfies the oauth2server ClientAuthorizedHandler interface.
func (s *Service) ClientAuthorizedHandler(clientID string, grant oauth2.GrantType) (allowed bool, err error) {
	// NOTE: it's a shame the interface we're implementing doesn't have this as its first argument
	ctx, span := tracing.StartSpan(context.Background(), "ClientAuthorizedHandler")
	defer span.End()

	logger := s.logger.WithValues(map[string]interface{}{
		"grant":     grant,
		"client_id": clientID,
	})

	// reject invalid grant type.
	if grant == oauth2.PasswordCredentials {
		return false, errInvalidGrantTypePassword
	}

	// fetch client data.
	client, err := s.clientDataManager.GetOAuth2ClientByClientID(ctx, clientID)
	if err != nil {
		logger.Error(err, "fetching oauth2 client from clientDataManager")
		return false, fmt.Errorf("fetching oauth2 client from clientDataManager: %w", err)
	}

	// disallow implicit grants unless authorized.
	if grant == oauth2.Implicit && !client.ImplicitAllowed {
		return false, errImplicitGrantsUnallowed
	}

	return true, nil
}

var _ oauth2server.ClientScopeHandler = (*Service)(nil).ClientScopeHandler

var errUnauthorized = errors.New("unauthorized")

// ClientScopeHandler satisfies the oauth2server ClientScopeHandler interface.
func (s *Service) ClientScopeHandler(clientID, scope string) (authed bool, err error) {
	// NOTE: it's a shame the interface we're implementing doesn't have this as its first argument
	ctx, span := tracing.StartSpan(context.Background(), "UserAuthorizationHandler")
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
