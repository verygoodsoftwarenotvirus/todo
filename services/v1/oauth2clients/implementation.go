package oauth2clients

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
	"gopkg.in/oauth2.v3"
	oauth2errors "gopkg.in/oauth2.v3/errors"
	oauth2server "gopkg.in/oauth2.v3/server"
)

// gopkg.in/oauth2.v3/server specific implementations

var _ oauth2server.InternalErrorHandler = (*Service)(nil).OAuth2InternalErrorHandler

// OAuth2InternalErrorHandler fulfills a role for the OAuth2 server-side provider
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

// OAuth2ResponseErrorHandler fulfills a role for the OAuth2 server-side provider
func (s *Service) OAuth2ResponseErrorHandler(re *oauth2errors.Response) {
	s.logger.WithValues(map[string]interface{}{
		"error":       re.Error,
		"error_code":  re.ErrorCode,
		"description": re.Description,
		"uri":         re.URI,
		"status_code": re.StatusCode,
		"header":      re.Header,
	})
}

var _ oauth2server.AuthorizeScopeHandler = (*Service)(nil).AuthorizeScopeHandler

// AuthorizeScopeHandler satisfies the oauth2server AuthorizeScopeHandler interface
func (s *Service) AuthorizeScopeHandler(res http.ResponseWriter, req *http.Request) (scope string, err error) {
	ctx := req.Context()
	scope = determineScope(req)

	logger := s.logger.WithValue("scope", scope).WithValue("BLAHBLAHBLAH", "BLAHBLAHBLAH").WithRequest(req)
	logger.Debug("AuthorizeScopeHandler called")

	var client = s.fetchOAuth2ClientFromRequest(req)
	if client != nil && client.HasScope(scope) {
		res.WriteHeader(http.StatusOK)
		return scope, nil
	}

	if clientID := s.fetchOAuth2ClientIDFromRequest(req); clientID != "" {
		client, err = s.database.GetOAuth2ClientByClientID(ctx, clientID)

		if err == sql.ErrNoRows {
			s.logger.Error(err, "error fetching OAuth2 Client")
			res.WriteHeader(http.StatusNotFound)
			return "", err
		} else if err != nil {
			s.logger.Error(err, "error fetching OAuth2 Client")
			res.WriteHeader(http.StatusInternalServerError)
			return "", err
		}

		if client.HasScope(scope) {
			res.WriteHeader(http.StatusOK)
			return scope, nil
		}

		res.WriteHeader(http.StatusUnauthorized)
		return "", errors.New("not authorized for scope")
	}

	res.WriteHeader(http.StatusBadRequest)
	return "", errors.New("no scope information found")
}

var _ oauth2server.UserAuthorizationHandler = (*Service)(nil).UserAuthorizationHandler

// UserAuthorizationHandler satisfies the oauth2server UserAuthorizationHandler interface
func (s *Service) UserAuthorizationHandler(res http.ResponseWriter, req *http.Request) (userID string, err error) {
	ctx := req.Context()
	s.logger.Debug("UserAuthorizationHandler called")

	var uid uint64
	if client, clientOk := ctx.Value(models.OAuth2ClientKey).(*models.OAuth2Client); !clientOk {
		user, ok := ctx.Value(models.UserKey).(*models.User)
		if !ok {
			s.logger.Debug("no user attached to this request")
			return "", errors.New("user not found")
		}
		uid = user.ID
	} else {
		uid = client.BelongsTo
	}
	return strconv.FormatUint(uid, 10), nil
}

var _ oauth2server.ClientAuthorizedHandler = (*Service)(nil).ClientAuthorizedHandler

// ClientAuthorizedHandler satisfies the oauth2server ClientAuthorizedHandler interface
func (s *Service) ClientAuthorizedHandler(clientID string, grant oauth2.GrantType) (allowed bool, err error) {
	s.logger.Debug("ClientAuthorizedHandler called")

	if grant == oauth2.PasswordCredentials {
		return false, errors.New("invalid grant type: password")
	}

	// NOTE: it's a shame the interface we're implementing doesn't have this as its first argument
	ctx := context.Background()

	client, err := s.database.GetOAuth2ClientByClientID(ctx, clientID)
	if err != nil {
		return false, err
	}

	if grant == oauth2.Implicit && !client.ImplicitAllowed {
		return false, errors.New("client not authorized for implicit grants")
	}

	return true, nil
}

var _ oauth2server.ClientScopeHandler = (*Service)(nil).ClientScopeHandler

// ClientScopeHandler satisfies the oauth2server ClientScopeHandler interface
func (s *Service) ClientScopeHandler(clientID, scope string) (authed bool, err error) {
	logger := s.logger.WithValues(map[string]interface{}{
		"client_id": clientID,
		"scope":     scope,
	})
	logger.Debug("ClientScopeHandler called")

	// NOTE: it's a shame the interface we're implementing doesn't have this as its first argument
	ctx := context.Background()

	c, err := s.database.GetOAuth2ClientByClientID(ctx, clientID)
	if err != nil {
		logger.Error(err, "error fetching OAuth2 client for ClientScopeHandler")
		return false, err
	}

	if c.HasScope(scope) {
		return true, nil
	}

	return false, errors.New("unauthorized")
}
