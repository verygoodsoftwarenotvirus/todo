package auth

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

// CookieAuthenticationMiddleware checks every request for a user cookie
func (s *Service) CookieAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		s.logger.Debug("userCookieAuthenticationMiddleware triggered")

		user, err := s.FetchUserFromRequest(req)
		if err != nil {
			s.logger.Error(err, "error encountered fetching user")
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		if user != nil {
			req = req.WithContext(context.WithValue(
				context.WithValue(req.Context(), models.UserKey, user),
				models.UserIDKey,
				user.ID,
			))

			next.ServeHTTP(res, req)
			return
		}
		http.Redirect(res, req, "/login", http.StatusUnauthorized)
	})
}

// AuthenticationMiddleware authenticates based on either an oauth2 token or a cookie
func (s *Service) AuthenticationMiddleware(allowValidCookieInLieuOfAValidToken bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			s.logger.Debug("apiAuthenticationMiddleware called")

			// First we check to see if there is an OAuth2 token for a valid client attached to the request.
			// We do this first because it is presumed to be the primary means by which requests are made to the httpServer.
			oauth2Client, err := s.oauth2ClientsService.RequestIsAuthenticated(req)
			if err != nil || oauth2Client == nil && allowValidCookieInLieuOfAValidToken {

				// In the event there's not a valid OAuth2 token attached to the request, or there is some other OAuth2 issue,
				// we next check to see if a valid cookie is attached to the request
				cookieAuth, cookieErr := s.DecodeCookieFromRequest(req)
				if cookieErr != nil || cookieAuth == nil {
					// If your request gets here, you're likely either trying to get here, or desperately trying to get anywhere
					s.logger.Error(err, "error authenticated token-authed request")
					http.Error(res, "invalid token", http.StatusUnauthorized)
					return
				}

				s.logger.WithValue("user_id", cookieAuth.UserID).Debug("attaching userID to request")
				ctx = context.WithValue(ctx, models.UserIDKey, cookieAuth.UserID)
				req = req.WithContext(ctx)
				next.ServeHTTP(res, req)
				return
			}

			// We found a valid OAuth2 client in the request, let's attach it and move on with our lives
			if oauth2Client != nil {
				// Attach both the user ID and the client object to the request. It might seem superfluous,
				// but some things should only need to know to look for user IDs, and not trouble themselves
				// with foolish concerns of OAuth2 clients and their fields
				ctx = context.WithValue(ctx, models.UserIDKey, oauth2Client.BelongsTo)
				ctx = context.WithValue(ctx, models.OAuth2ClientKey, oauth2Client)
				req = req.WithContext(ctx)
				next.ServeHTTP(res, req)
				return
			}

			http.Redirect(res, req, "/login", http.StatusUnauthorized)
			return
		})
	}
}