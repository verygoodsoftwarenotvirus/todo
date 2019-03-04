package oauth2clients

import (
	"context"
	"encoding/json"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
)

// OAuth2ClientCreationInputContextMiddleware is a middleware for attaching OAuth2 client info to a request
func (s *Service) OAuth2ClientCreationInputContextMiddleware(next http.Handler) http.Handler {
	x := new(models.OAuth2ClientCreationInput)

	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		s.logger.Debug("OAuth2ClientCreationInputContextMiddleware called")
		if err := json.NewDecoder(req.Body).Decode(x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(req.Context(), MiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// OAuth2TokenAuthenticationMiddleware authenticates Oauth tokens
func (s *Service) OAuth2TokenAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		s.logger.Debug("OAuth2TokenAuthenticationMiddleware called")

		token, err := s.oauth2Handler.ValidationBearerToken(req)
		if err != nil || token == nil {
			s.logger.Error(err, "error validating bearer token")
			http.Error(res, "invalid token", http.StatusUnauthorized)
			return
		}

		// ignoring this error because the User ID source should only ever provide uints
		clientID := token.GetClientID()
		logger := s.logger.WithValues(map[string]interface{}{
			"client_id": clientID,
		})

		c, err := s.database.GetOAuth2ClientByClientID(ctx, clientID)
		if err != nil {
			logger.Error(err, "error fetching OAuth2 Client")
			http.Error(res, errors.Wrap(err, "error fetching client ID").Error(), http.StatusUnauthorized)
			// http.Redirect(res, req, "/login", http.StatusUnauthorized)
			return
		}

		req = req.WithContext( // attach both the user ID and the client object to the request. it might seem superfluous,
			context.WithValue( // but some things should only need to know to look for user IDs, and not trouble themselves
				context.WithValue( // with foolish concerns of OAuth2 clients and their fields
					ctx, models.UserIDKey, c.BelongsTo,
				),
				models.OAuth2ClientKey,
				c,
			),
		)
		next.ServeHTTP(res, req)
	})
}

// BuildAuthenticationMiddleware provides a way of building middleware with varying behaviors
func (s *Service) BuildAuthenticationMiddleware(reject bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			logger := s.logger.WithValue("rejecting", reject)
			logger.Debug("OAuth2TokenAuthenticationMiddleware called")

			token, err := s.oauth2Handler.ValidationBearerToken(req)
			if err != nil || token == nil {
				logger.Error(err, "error validating bearer token")
				if reject {
					http.Error(res, "invalid token", http.StatusUnauthorized)
				} else {
					next.ServeHTTP(res, req)
				}
				return
			}

			// ignoring this error because the User ID source should only ever provide uints
			clientID := token.GetClientID()
			logger = logger.WithValues(map[string]interface{}{
				"client_id": clientID,
			})

			c, err := s.database.GetOAuth2ClientByClientID(ctx, clientID)
			if err != nil {
				logger.Error(err, "error fetching OAuth2 Client")
				if reject {
					http.Error(res, errors.Wrap(err, "error fetching client ID").Error(), http.StatusUnauthorized)
					// http.Redirect(res, req, "/login", http.StatusUnauthorized)
				} else {
					next.ServeHTTP(res, req)
				}
				return
			}

			req = req.WithContext( // attach both the user ID and the client object to the request. it might seem superfluous,
				context.WithValue( // but some things should only need to know to look for user IDs, and not trouble themselves
					context.WithValue( // with foolish concerns of OAuth2 clients and their fields
						ctx, models.UserIDKey, c.BelongsTo,
					),
					models.OAuth2ClientKey,
					c,
				),
			)
			next.ServeHTTP(res, req)
		})
	}
}

// OAuth2ClientInfoMiddleware fetches clientOAuth2Client info from requests and attaches it eplicitly to a request
func (s *Service) OAuth2ClientInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		s.logger.Debug("OauthInfoMiddleware called")

		if v := req.URL.Query().Get(oauth2ClientIDURIParamKey); v != "" {
			logger := s.logger.WithValue("oauth2_client_id", v)

			client, err := s.database.GetOAuth2ClientByClientID(ctx, v)
			if err != nil {
				logger.Error(err, "error fetching OAuth2 client")
				http.Error(res, err.Error(), http.StatusInternalServerError)
			}
			req = req.WithContext(context.WithValue(ctx, models.OAuth2ClientKey, client))
		}

		next.ServeHTTP(res, req)
	})
}

func (s *Service) fetchOAuth2ClientFromRequest(req *http.Request) *models.OAuth2Client {
	s.logger.Debug("fetchOAuth2ClientFromRequest called")
	ctx := req.Context()
	client, ok := ctx.Value(models.OAuth2ClientKey).(*models.OAuth2Client)
	if !ok {
		return nil
	}
	return client
}

func (s *Service) fetchOAuth2ClientScopesFromRequest(req *http.Request) []string {
	s.logger.Debug("fetchOAuth2ClientScopesFromRequest called")
	ctx := req.Context()
	scopes, ok := ctx.Value(scopesKey).([]string)
	if !ok {
		return nil
	}
	return scopes
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
