package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gopkg.in/oauth2.v3"
	oauth2errors "gopkg.in/oauth2.v3/errors"
	oauth2manage "gopkg.in/oauth2.v3/manage"
	oauth2models "gopkg.in/oauth2.v3/models"
	oauth2server "gopkg.in/oauth2.v3/server"
	oauth2store "gopkg.in/oauth2.v3/store"
)

const (
	scopesKey   models.ContextKey = "scopes"
	clientIDKey models.ContextKey = "client_id"

	scopesSeparator           = ","
	oauth2ClientIDURIParamKey = "client_id"
)

var (
	errInvalidToken = errors.New("invalid token provided")
)

// ProvideTokenStore provides a token store for use with the server
func ProvideTokenStore(manager *oauth2manage.Manager) (oauth2.TokenStore, error) {
	tokenStore, err := oauth2store.NewMemoryTokenStore()
	manager.MustTokenStorage(tokenStore, err)
	return tokenStore, nil
}

// ProvideClientStore provides a client store for use with the server
func ProvideClientStore() *oauth2store.ClientStore {
	return oauth2store.NewClientStore()
}

// ProvideOAuth2Server provides an oauth2server.Server that meets the server's specifications
func ProvideOAuth2Server(manager *oauth2manage.Manager, tokenStore oauth2.TokenStore, clientStore *oauth2store.ClientStore) *oauth2server.Server {
	manager.MapClientStorage(clientStore)

	oauth2Handler := oauth2server.NewDefaultServer(manager)

	return oauth2Handler
}

// ProvideOAuth2Service provides an OAuth2 Clients service
func (s *Server) initializeOAuth2Clients() {
	paginating := true
	for page := 1; paginating; page++ {

		clientList, err := s.db.GetOAuth2Clients(
			context.Background(),
			&models.QueryFilter{
				Page:  uint64(page),
				Limit: 50,
			},
		)

		if (clientList != nil && len(clientList.Clients) == 0) || err == sql.ErrNoRows {
			paginating = false
		} else if err != nil {
			s.logger.Fatalln("error encountered querying oauth clients to add to the clientStore: ", err)
		}

		for _, client := range clientList.Clients {
			s.logger.WithField("client_id", client.ClientID).Debugln("loading client")
			if err := s.oauth2ClientStore.Set(client.ClientID, &oauth2models.Client{
				ID:     client.ClientID,
				Secret: client.ClientSecret,
				Domain: client.RedirectURI,
				UserID: strconv.FormatUint(client.BelongsTo, 10),
			}); err != nil {
				s.logger.WithError(err).Fatalln("error encountered loading oauth clients to the clientStore")
			}
		}
	}

	s.oauth2Handler.SetAllowGetAccessRequest(true)
	s.oauth2Handler.SetClientAuthorizedHandler(s.ClientAuthorizedHandler)
	s.oauth2Handler.SetClientScopeHandler(s.ClientScopeHandler)
	s.oauth2Handler.SetClientInfoHandler(oauth2server.ClientFormHandler)
	s.oauth2Handler.SetUserAuthorizationHandler(s.UserAuthorizationHandler)
	s.oauth2Handler.SetAuthorizeScopeHandler(s.AuthorizeScopeHandler)
	s.oauth2Handler.SetResponseErrorHandler(s.OAuth2ResponseErrorHandler)
	s.oauth2Handler.SetInternalErrorHandler(s.OAuth2InternalErrorHandler)
	s.oauth2Handler.Config.AllowedGrantTypes = []oauth2.GrantType{
		oauth2.AuthorizationCode,
		oauth2.ClientCredentials,
		oauth2.Refreshing,
		oauth2.Implicit,
	}
}

// gopkg.in/oauth2.v3/server specific implementations

var _ oauth2server.InternalErrorHandler = (*Server)(nil).OAuth2InternalErrorHandler

// OAuth2InternalErrorHandler fulfills a role for the OAuth2 server-side provider
func (s *Server) OAuth2InternalErrorHandler(err error) *oauth2errors.Response {
	res := &oauth2errors.Response{
		Error:       err,
		Description: "Internal error",
		ErrorCode:   http.StatusInternalServerError,
		StatusCode:  http.StatusInternalServerError,
	}

	s.logger.WithError(err).Errorln("Internal Error")
	return res
}

var _ oauth2server.ResponseErrorHandler = (*Server)(nil).OAuth2ResponseErrorHandler

// OAuth2ResponseErrorHandler fulfills a role for the OAuth2 server-side provider
func (s *Server) OAuth2ResponseErrorHandler(re *oauth2errors.Response) {
	s.logger.WithFields(map[string]interface{}{
		"error":       re.Error,
		"error_code":  re.ErrorCode,
		"description": re.Description,
		"URI":         re.URI,
		"status_code": re.StatusCode,
		"header":      re.Header,
	})
}

// OauthTokenAuthenticationMiddleware authenticates Oauth tokens
func (s *Server) OauthTokenAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		s.logger.Debugln("OauthTokenAuthenticationMiddleware called")

		token, err := s.oauth2Handler.ValidationBearerToken(req)
		if err != nil || token == nil {
			s.logger.WithError(err).Errorln("error validating bearer token")
			http.Error(res, errInvalidToken.Error(), http.StatusUnauthorized)
			return
		}

		cid := token.GetClientID()
		logger := s.logger.WithField("client_id", cid)

		c, err := s.db.GetOAuth2Client(ctx, cid)
		if err != nil {
			logger.WithError(err).Errorln("error fetching OAuth2 Client")
			http.Error(res, fmt.Sprintf("error fetching client ID: %s", err.Error()), http.StatusUnauthorized)
			return
		}

		req = req.WithContext(context.WithValue(ctx, models.UserIDKey, c.BelongsTo))
		next.ServeHTTP(res, req)
	})
}

// OAuth2ClientInfoMiddleware fetches clientOAuth2Client info from requests and attaches it eplicitly to a request
func (s *Server) OAuth2ClientInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		s.logger.Debugln("OauthInfoMiddleware called")

		if v := req.URL.Query().Get(oauth2ClientIDURIParamKey); v != "" {
			logger := s.logger.WithField("oauth2_client_id", v)
			logger.Debugln("fetching oauth2 client from database")

			client, err := s.db.GetOAuth2Client(ctx, v)
			if err != nil {
				logger.WithError(err).Errorln("error fetching OAuth2 client")
				http.Error(res, err.Error(), http.StatusInternalServerError)
			}
			req = req.WithContext(context.WithValue(ctx, models.OAuth2ClientKey, client))
		}

		next.ServeHTTP(res, req)
	})
}

func (s *Server) fetchOAuth2ClientFromRequest(req *http.Request) *models.OAuth2Client {
	s.logger.Debugln("fetchOAuth2ClientFromRequest called")
	ctx := req.Context()
	client, ok := ctx.Value(models.OAuth2ClientKey).(*models.OAuth2Client)
	if !ok {
		return nil
	}
	return client
}

func (s *Server) fetchOAuth2ClientScopesFromRequest(req *http.Request) []string {
	s.logger.Debugln("fetchOAuth2ClientScopesFromRequest called")
	ctx := req.Context()
	scopes, ok := ctx.Value(scopesKey).([]string)
	if !ok {
		return nil
	}
	return scopes
}

func (s *Server) fetchOAuth2ClientIDFromRequest(req *http.Request) string {
	s.logger.Debugln("fetchOAuth2ClientIDFromRequest called")
	ctx := req.Context()
	clientID, ok := ctx.Value(clientIDKey).(string)
	if !ok {
		return ""
	}
	return clientID
}

// gopkg.in/oauth2.v3/server specific implementations

var _ oauth2server.AuthorizeScopeHandler = (*Server)(nil).AuthorizeScopeHandler

// AuthorizeScopeHandler satisfies the oauth2server AuthorizeScopeHandler interface
func (s *Server) AuthorizeScopeHandler(res http.ResponseWriter, req *http.Request) (scope string, err error) {
	ctx := req.Context()

	s.logger.Debugln("AuthorizeScopeHandler called")
	client := s.fetchOAuth2ClientFromRequest(req)
	logger := s.logger.WithFields(map[string]interface{}{
		"client.ID": client.ID,
	})

	if client == nil {
		clientID := s.fetchOAuth2ClientIDFromRequest(req)
		if clientID != "" {
			client, err := s.db.GetOAuth2Client(ctx, clientID)
			if err != nil {
				logger.WithError(err).Errorln("error fetching OAuth2 Client")
				return "", err
			}

			req = req.WithContext(context.WithValue(ctx, models.OAuth2ClientKey, client))
			return strings.Join(client.Scopes, scopesSeparator), nil
		}
	} else {
		return strings.Join(client.Scopes, scopesSeparator), nil
	}

	logger.Errorln("no scope information found")
	return "", errors.New("no scope information found")
}

var _ oauth2server.UserAuthorizationHandler = (*Server)(nil).UserAuthorizationHandler

// UserAuthorizationHandler satisfies the oauth2server UserAuthorizationHandler interface
func (s *Server) UserAuthorizationHandler(res http.ResponseWriter, req *http.Request) (userID string, err error) {
	ctx := req.Context()
	s.logger.Debugln("UserAuthorizationHandler called")

	var uid uint64
	if client, clientOk := ctx.Value(models.OAuth2ClientKey).(*models.OAuth2Client); !clientOk {
		user, ok := ctx.Value(models.UserKey).(*models.User)
		if !ok {
			s.logger.Debugln("no user attached to this request")
			return "", errors.New("user not found")
		}
		uid = user.ID
	} else {
		uid = client.BelongsTo
	}
	return strconv.FormatUint(uid, 10), nil
}

var _ oauth2server.ClientAuthorizedHandler = (*Server)(nil).ClientAuthorizedHandler

// ClientAuthorizedHandler satisfies the oauth2server ClientAuthorizedHandler interface
func (s *Server) ClientAuthorizedHandler(clientID string, grant oauth2.GrantType) (allowed bool, err error) {
	s.logger.Debugln("ClientAuthorizedHandler called")

	if grant == oauth2.PasswordCredentials {
		return false, errors.New("invalid grant type: password")
	}

	// TODO: what if client is deactivated?!
	client, err := s.db.GetOAuth2Client(context.Background(), clientID)
	if err != nil {
		return false, err
	}

	if grant == oauth2.Implicit && !client.ImplicitAllowed {
		return false, errors.New("client not authorized for implicit grants")
	}

	return true, nil
}

var _ oauth2server.ClientScopeHandler = (*Server)(nil).ClientScopeHandler

// ClientScopeHandler satisfies the oauth2server ClientScopeHandler interface
func (s *Server) ClientScopeHandler(clientID, scope string) (authed bool, err error) {
	logger := s.logger.WithFields(map[string]interface{}{
		"client_id": clientID,
		"scope":     scope,
	})
	logger.Debugln("ClientScopeHandler called")

	c, err := s.db.GetOAuth2Client(context.Background(), clientID)
	if err != nil {
		logger.WithError(err).Errorln("error fetching OAuth2 client for ClientScopeHandler")
		return false, err
	}

	logger = logger.WithField("oauth2_client", c)
	logger.Debugln("OAuth2 Client retrieved in ClientScopeHandler")

	for _, s := range c.Scopes {
		if s == scope || s == "*" {
			authed = true
		}
	}

	logger.
		WithField("authed", authed).
		Debugln("returning from ClientScopeHandler")

	return authed, nil
}
