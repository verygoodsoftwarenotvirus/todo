package oauth2clients

import (
	"context"
	"net/http"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
)

// const (
// scopesKey models.ContextKey = "scopes"
// )

// CreationInputMiddleware is a middleware for attaching OAuth2 client info to a request
func (s *Service) CreationInputMiddleware(next http.Handler) http.Handler {
	x := new(models.OAuth2ClientCreationInput)

	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		s.logger.Debug("OAuth2ClientCreationInputContextMiddleware called")
		if err := s.encoderDecoder.DecodeRequest(req, x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(req.Context(), MiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// RequestIsAuthenticated RENAMEME returns whether or not the request is authenticated
func (s *Service) RequestIsAuthenticated(req *http.Request) (*models.OAuth2Client, error) {
	ctx := req.Context()
	logger := s.logger.WithValue("function_name", "RequestIsAuthenticated")
	logger.Debug("RequestIsAuthenticated called")

	token, err := s.oauth2Handler.ValidationBearerToken(req)
	if err != nil {
		return nil, errors.Wrap(err, "validating bearer token")
	}

	// ignoring this error because the User ID source should only ever provide uints
	clientID := token.GetClientID()
	logger = logger.WithValue("client_id", clientID)

	c, err := s.database.GetOAuth2ClientByClientID(ctx, clientID)
	if err != nil {
		logger.Error(err, "error fetching OAuth2 Client")
		return nil, err
	}

	scope := determineScope(req)
	hasScope := c.HasScope(scope)

	logger = logger.WithValue("scope", scope).
		WithValue("scopes", c.Scopes).
		WithValue("has_scope", hasScope)

	if !hasScope {
		logger.Info("rejecting client for invalid scope")
		return nil, errors.New("client not authorized for scope")
	}

	logger.Info("accepting client for valid scope")
	return c, nil
}

func determineScope(req *http.Request) string {
	if strings.HasPrefix(req.URL.Path, "/api/v1/") {
		x := strings.TrimPrefix(req.URL.Path, "/api/v1/")
		if y := strings.Split(x, "/"); len(y) > 0 {
			x = y[0]
		}
		return x
	}
	return ""
}

// OAuth2TokenAuthenticationMiddleware authenticates Oauth tokens
func (s *Service) OAuth2TokenAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		s.logger.Debug("OAuth2TokenAuthenticationMiddleware called")

		c, err := s.RequestIsAuthenticated(req)
		if err != nil {
			s.logger.Error(err, "error authenticated token-authed request")
			http.Error(res, "invalid token", http.StatusUnauthorized)
			return
		}

		// attach both the user ID and the client object to the request. it might seem
		// superfluous, but some things should only need to know to look for user IDs
		ctx2 := context.WithValue(ctx, models.UserIDKey, c.BelongsTo)
		ctx3 := context.WithValue(ctx2, models.OAuth2ClientKey, c)
		req2 := req.WithContext(ctx3)

		next.ServeHTTP(res, req2)
	})
}

// OAuth2ClientInfoMiddleware fetches clientOAuth2Client info from requests and attaches it explicitly to a request
func (s *Service) OAuth2ClientInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		s.logger.Debug("OauthInfoMiddleware called")

		if v := req.URL.Query().Get(oauth2ClientIDURIParamKey); v != "" {
			logger := s.logger.WithValue("oauth2_client_id", v)

			client, err := s.database.GetOAuth2ClientByClientID(ctx, v)
			if err != nil {
				logger.Error(err, "error fetching OAuth2 client")
				http.Error(res, "invalid request", http.StatusUnauthorized)
				return
			}
			req = req.WithContext(context.WithValue(ctx, models.OAuth2ClientKey, client))
		}

		next.ServeHTTP(res, req)
	})
}

func (s *Service) fetchOAuth2ClientFromRequest(req *http.Request) *models.OAuth2Client {
	logger := s.logger.WithValue("function_name", "fetchOAuth2ClientFromRequest")
	logger.Debug("called")

	ctx := req.Context()
	client, ok := ctx.Value(models.OAuth2ClientKey).(*models.OAuth2Client)

	if !ok {
		logger.Debug("returning nil, no oauth2 client found")
		return nil
	}
	return client
}

func (s *Service) fetchOAuth2ClientIDFromRequest(req *http.Request) string {
	s.logger.Debug("fetchOAuth2ClientIDFromRequest called")
	ctx := req.Context()
	clientID, ok := ctx.Value(clientIDKey).(string)
	if !ok {
		return ""
	}
	return clientID
}
