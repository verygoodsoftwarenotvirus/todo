package oauth2clients

import (
	"context"
	"errors"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	scopesSeparator = ","
)

var errClientUnauthorizedForScope = errors.New("client not authorized for scope")

// CreationInputMiddleware is a middleware for attaching OAuth2 client info to a request.
func (s *Service) CreationInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := tracing.StartSpan(req.Context(), "oauth2clients.service.CreationInputMiddleware")
		defer span.End()
		x := new(types.OAuth2ClientCreationInput)

		// decode value from request.
		if err := s.encoderDecoder.DecodeRequest(req, x); err != nil {
			s.logger.Error(err, "error encountered decoding request body")
			s.encoderDecoder.EncodeErrorResponse(res, "invalid request content", http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, creationMiddlewareCtxKey, x)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// OAuth2TokenAuthenticationMiddleware authenticates Oauth tokens.
func (s *Service) OAuth2TokenAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := tracing.StartSpan(req.Context(), "oauth2clients.service.OAuth2TokenAuthenticationMiddleware")
		defer span.End()

		c, err := s.ExtractOAuth2ClientFromRequest(ctx, req)
		if err != nil {
			s.logger.Error(err, "error authenticated token-authed request")
			http.Error(res, "invalid token", http.StatusUnauthorized)
			return
		}

		tracing.AttachUserIDToSpan(span, c.BelongsToUser)
		tracing.AttachOAuth2ClientIDToSpan(span, c.ClientID)
		tracing.AttachOAuth2ClientDatabaseIDToSpan(span, c.ID)

		// attach the client object to the request.
		ctx = context.WithValue(ctx, types.OAuth2ClientKey, c)

		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// OAuth2ClientInfoMiddleware fetches clientOAuth2Client info from requests and attaches it explicitly to a request.
func (s *Service) OAuth2ClientInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := tracing.StartSpan(req.Context(), "oauth2clients.service.OAuth2ClientInfoMiddleware")
		defer span.End()

		if v := req.URL.Query().Get(oauth2ClientIDURIParamKey); v != "" {
			logger := s.logger.WithValue("oauth2_client_id", v)

			client, err := s.clientDataManager.GetOAuth2ClientByClientID(ctx, v)
			if err != nil {
				logger.Error(err, "error fetching OAuth2 client")
				http.Error(res, "invalid request", http.StatusUnauthorized)
				return
			}

			tracing.AttachUserIDToSpan(span, client.BelongsToUser)
			tracing.AttachOAuth2ClientIDToSpan(span, client.ClientID)
			tracing.AttachOAuth2ClientDatabaseIDToSpan(span, client.ID)

			ctx = context.WithValue(ctx, types.OAuth2ClientKey, client)

			req = req.WithContext(ctx)
		}

		next.ServeHTTP(res, req)
	})
}

func (s *Service) fetchOAuth2ClientFromRequest(req *http.Request) *types.OAuth2Client {
	if client, ok := req.Context().Value(types.OAuth2ClientKey).(*types.OAuth2Client); ok {
		return client
	}
	return nil
}

func (s *Service) fetchOAuth2ClientIDFromRequest(req *http.Request) string {
	if clientID, ok := req.Context().Value(clientIDKey).(string); ok {
		return clientID
	}
	return ""
}
