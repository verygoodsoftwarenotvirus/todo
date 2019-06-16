package oauth2clients

import (
	"context"
	"net/http"
	"strings"

	"go.opencensus.io/trace"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
)

const (
	scopesSeparator = ","
)

// CreationInputMiddleware is a middleware for attaching OAuth2 client info to a request
func (s *Service) CreationInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "CreationInputMiddleware")
		defer span.End()
		x := new(models.OAuth2ClientCreationInput)

		if err := s.encoderDecoder.DecodeRequest(req, x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, MiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// ExtractOAuth2ClientFromRequest extracts OAuth2 client data from a request
func (s *Service) ExtractOAuth2ClientFromRequest(ctx context.Context, req *http.Request) (*models.OAuth2Client, error) {
	ctx, span := trace.StartSpan(ctx, "ExtractOAuth2ClientFromRequest")
	defer span.End()

	logger := s.logger.WithValue("function_name", "ExtractOAuth2ClientFromRequest")

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
		WithValue("scopes", strings.Join(c.Scopes, scopesSeparator))

	if !hasScope {
		logger.Info("rejecting client for invalid scope")
		return nil, errors.New("client not authorized for scope")
	}

	return c, nil
}

const apiPathPrefix = "/api/v1/"

func determineScope(req *http.Request) string {
	if strings.HasPrefix(req.URL.Path, apiPathPrefix) {
		x := strings.TrimPrefix(req.URL.Path, apiPathPrefix)
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
		ctx, span := trace.StartSpan(req.Context(), "OAuth2TokenAuthenticationMiddleware")
		defer span.End()

		c, err := s.ExtractOAuth2ClientFromRequest(ctx, req)
		if err != nil {
			s.logger.Error(err, "error authenticated token-authed request")
			http.Error(res, "invalid token", http.StatusUnauthorized)
			return
		}

		attachUserIDToSpan(span, c.BelongsTo)
		attachOAuth2ClientDatabaseIDToSpan(span, c.ID)
		attachOAuth2ClientIDToSpan(span, c.ClientID)

		// attach both the user ID and the client object to the request. it might seem
		// superfluous, but some things should only need to know to look for user IDs
		ctx2 := context.WithValue(ctx, models.OAuth2ClientKey, c)
		ctx3 := context.WithValue(ctx2, models.UserIDKey, c.BelongsTo)

		next.ServeHTTP(res, req.WithContext(ctx3))
	})
}

// OAuth2ClientInfoMiddleware fetches clientOAuth2Client info from requests and attaches it explicitly to a request
func (s *Service) OAuth2ClientInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := trace.StartSpan(req.Context(), "OAuth2ClientInfoMiddleware")
		defer span.End()

		if v := req.URL.Query().Get(oauth2ClientIDURIParamKey); v != "" {
			logger := s.logger.WithValue("oauth2_client_id", v)

			client, err := s.database.GetOAuth2ClientByClientID(ctx, v)
			if err != nil {
				logger.Error(err, "error fetching OAuth2 client")
				http.Error(res, "invalid request", http.StatusUnauthorized)
				return
			}

			attachOAuth2ClientIDToSpan(span, client.ClientID)
			attachOAuth2ClientDatabaseIDToSpan(span, client.ID)
			attachUserIDToSpan(span, client.BelongsTo)

			req = req.WithContext(context.WithValue(ctx, models.OAuth2ClientKey, client))
			req = req.WithContext(context.WithValue(ctx, models.UserIDKey, client.BelongsTo))
		}

		next.ServeHTTP(res, req)
	})
}

func (s *Service) fetchOAuth2ClientFromRequest(req *http.Request) *models.OAuth2Client {
	client, ok := req.Context().Value(models.OAuth2ClientKey).(*models.OAuth2Client)
	_ = ok // we don't really care, but the linters do
	return client
}

func (s *Service) fetchOAuth2ClientIDFromRequest(req *http.Request) string {
	clientID, ok := req.Context().Value(clientIDKey).(string)
	_ = ok // we don't really care, but the linters do
	return clientID
}
