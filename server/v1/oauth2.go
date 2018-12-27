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
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"

	// "github.com/sirupsen/logrus"
	"gopkg.in/oauth2.v3"
	oauth2errors "gopkg.in/oauth2.v3/errors"
	oauth2manage "gopkg.in/oauth2.v3/manage"
	oauth2models "gopkg.in/oauth2.v3/models"
	oauth2server "gopkg.in/oauth2.v3/server"
	oauth2store "gopkg.in/oauth2.v3/store"
)

const (
	scopesSeparator                             = ","
	scopesKey                 models.ContextKey = "scopes"
	clientIDKey               models.ContextKey = "client_id"
	oauth2ClientIDURIParamKey                   = "client_id"
)

func (s *Server) initializeOauth2Server() {
	manager := oauth2manage.NewDefaultManager()
	// token memory store
	tokenStore, err := oauth2store.NewMemoryTokenStore()
	manager.MustTokenStorage(tokenStore, err)

	// client memory store
	s.oauth2ClientStore = oauth2store.NewClientStore()

	paginating := true
	for page := 1; paginating; page++ {
		clientList, err := s.db.GetOauth2Clients(
			&models.QueryFilter{
				Page:  uint64(page),
				Limit: 50,
			},
		)

		if clientList != nil && len(clientList.Clients) == 0 || err == sql.ErrNoRows {
			paginating = false
		} else if err != nil {
			s.logger.Fatalln("error encountered querying oauth clients to add to the clientStore: ", err)
		}

		for _, client := range clientList.Clients {
			s.logger.Debugf("loading client %q", client.ClientID)
			if err := s.oauth2ClientStore.Set(client.ClientID, &oauth2models.Client{
				ID:     client.ClientID,
				Secret: client.ClientSecret,
				Domain: client.RedirectURI,
				UserID: strconv.FormatUint(client.BelongsTo, 10),
			}); err != nil {
				s.logger.Fatalln("error encountered loading oauth clients to the clientStore: ", err)
			}

		}
	}

	manager.MapClientStorage(s.oauth2ClientStore)

	s.setOauth2Defaults(manager)

	s.oauth2ClientsService = oauth2clients.NewOauth2ClientsService(
		oauth2clients.Oauth2ClientsServiceConfig{
			Logger:        s.logger,
			Authenticator: s.authenticator,
			Database:      s.db,
			ClientStore:   s.oauth2ClientStore,
			TokenStore:    tokenStore,
		},
	)
}

func (s *Server) setOauth2Defaults(manager *oauth2manage.Manager) {
	s.oauth2Handler = oauth2server.NewDefaultServer(manager)

	s.oauth2Handler.SetAllowGetAccessRequest(true)
	// s.oauth2Handler.SetAccessTokenExpHandler(s.AccessTokenExpirationHandler)
	s.oauth2Handler.SetClientAuthorizedHandler(s.ClientAuthorizedHandler)
	s.oauth2Handler.SetClientScopeHandler(s.ClientScopeHandler)
	s.oauth2Handler.SetClientInfoHandler(oauth2server.ClientFormHandler)
	s.oauth2Handler.SetUserAuthorizationHandler(s.UserAuthorizationHandler)
	s.oauth2Handler.SetAuthorizeScopeHandler(s.AuthorizeScopeHandler)
	s.oauth2Handler.Config.AllowedGrantTypes = []oauth2.GrantType{
		oauth2.AuthorizationCode, oauth2.ClientCredentials, oauth2.Refreshing, oauth2.Implicit,
	}

	s.oauth2Handler.SetInternalErrorHandler(func(err error) (re *oauth2errors.Response) {
		s.logger.Errorln("Internal Error:", err.Error())
		return
	})

	s.oauth2Handler.SetResponseErrorHandler(func(re *oauth2errors.Response) {
		s.logger.WithFields(map[string]interface{}{
			"error":       re.Error,
			"error_code":  re.ErrorCode,
			"description": re.Description,
			"URI":         re.URI,
			"status_code": re.StatusCode,
			"header":      re.Header,
		})
	})
}

func (s *Server) OauthTokenAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		token, err := s.oauth2Handler.ValidationBearerToken(req)
		if err != nil || token == nil {
			http.Error(res, err.Error(), http.StatusUnauthorized)
			return
		}

		cid := token.GetClientID()
		c, err := s.db.GetOauth2Client(cid)
		if err != nil {
			http.Error(res, fmt.Sprintf("error fetching client ID: %s", err.Error()), http.StatusUnauthorized)
			return
		}

		req = req.WithContext(context.WithValue(req.Context(), models.UserIDKey, c.BelongsTo))
		next.ServeHTTP(res, req)
	})
}

func (s *Server) userIDFetcher(req *http.Request) uint64 {
	x, ok := req.Context().Value(models.UserIDKey).(uint64)
	if !ok {
		s.logger.Errorln("no input attached to request")
	}
	return x
}

func (s *Server) Oauth2ClientInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		s.logger.Debugln("OauthInfoMiddleware triggered")

		values := req.URL.Query()
		if v := values.Get(oauth2ClientIDURIParamKey); v != "" {
			s.logger.Debugf("fetching oauth2 client %s from database", v)
			client, err := s.db.GetOauth2Client(v)
			if err != nil {
				s.logger.Errorln("error fetching ")
				http.Error(res, err.Error(), http.StatusInternalServerError)
			}
			req = req.WithContext(context.WithValue(req.Context(), models.Oauth2ClientKey, client))
		}

		next.ServeHTTP(res, req)
	})
}

func (s *Server) fetchOauth2ClientFromRequest(req *http.Request) *models.Oauth2Client {
	client, ok := req.Context().Value(models.Oauth2ClientKey).(*models.Oauth2Client)
	if !ok {
		return nil
	}
	return client
}

func (s *Server) fetchOauth2ClientScopesFromRequest(req *http.Request) []string {
	scopes, ok := req.Context().Value(scopesKey).([]string)
	if !ok {
		return nil
	}
	return scopes
}

func (s *Server) fetchOauth2ClientIDFromRequest(req *http.Request) string {
	clientID, ok := req.Context().Value(clientIDKey).(string)
	if !ok {
		return ""
	}
	return clientID
}

// gopkg.in/oauth2.v3/server specific implementations

var _ oauth2server.AuthorizeScopeHandler = (*Server)(nil).AuthorizeScopeHandler

func (s *Server) AuthorizeScopeHandler(res http.ResponseWriter, req *http.Request) (scope string, err error) {
	s.logger.Debugln("AuthorizeScopeHandler called")
	client := s.fetchOauth2ClientFromRequest(req)
	if client == nil {
		clientID := s.fetchOauth2ClientIDFromRequest(req)
		if clientID != "" {
			client, err := s.db.GetOauth2Client(clientID)
			if err != nil {
				return "", err
			}

			req = req.WithContext(context.WithValue(req.Context(), models.Oauth2ClientKey, client))
			return strings.Join(client.Scopes, scopesSeparator), nil
		}
	} else {
		return strings.Join(client.Scopes, scopesSeparator), nil
	}

	return "", errors.New("no scope information found")
}

var _ oauth2server.UserAuthorizationHandler = (*Server)(nil).UserAuthorizationHandler

func (s *Server) UserAuthorizationHandler(res http.ResponseWriter, req *http.Request) (userID string, err error) {
	s.logger.Debugln("UserAuthorizationHandler called")
	ctx := req.Context()
	var uid uint64
	if client, clientOk := ctx.Value(models.Oauth2ClientKey).(*models.Oauth2Client); !clientOk {
		user, ok := ctx.Value(models.UserKey).(*models.User)
		if !ok {
			return "", errors.New("user not found")
		}
		uid = user.ID
	} else {
		uid = client.BelongsTo
	}
	return strconv.FormatUint(uid, 10), nil
}

// var _ oauth2server.AccessTokenExpHandler = (*Server)(nil).AccessTokenExpirationHandler

// func (s *Server) AccessTokenExpirationHandler(w http.ResponseWriter, r *http.Request) (time.Duration, error) {
// 	return 10 * time.Minute, nil
// }

var _ oauth2server.ClientAuthorizedHandler = (*Server)(nil).ClientAuthorizedHandler

func (s *Server) ClientAuthorizedHandler(clientID string, grant oauth2.GrantType) (allowed bool, err error) {
	s.logger.Debugln("ClientAuthorizedHandler called")
	// AuthorizationCode   GrantType = "authorization_code"
	// ClientCredentials   GrantType = "client_credentials"
	// Refreshing          GrantType = "refresh_token"
	// Implicit            GrantType = "__implicit"
	// PasswordCredentials GrantType = "password"

	if grant == oauth2.PasswordCredentials {
		return false, errors.New("invalid grant type: password")
	}

	// TODO: what if client is deactivated?!
	client, err := s.db.GetOauth2Client(clientID)
	if err != nil {
		return false, err
	}

	if grant == oauth2.Implicit && !client.ImplicitAllowed {
		return false, errors.New("client not authorized for implicit grants")
	}

	return true, nil
}

var _ oauth2server.ClientScopeHandler = (*Server)(nil).ClientScopeHandler

func (s *Server) ClientScopeHandler(clientID, scope string) (allowed bool, err error) {
	s.logger.Debugln("ClientScopeHandler called")
	if c, err := s.db.GetOauth2Client(clientID); err != nil {
		return false, err
	} else {
		for _, s := range c.Scopes {
			if s == scope || s == "*" {
				return true, nil
			}
		}
	}
	return
}
