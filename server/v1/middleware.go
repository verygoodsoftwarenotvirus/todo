package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"

	"github.com/go-chi/chi/middleware"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
)

func (s *Server) apiAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		s.logger.Debug("apiAuthenticationMiddleware called")

		// First we check to see if there is an OAuth2 token for a valid client attached to the request.
		// We do this first because it is presumed to be the primary means by which requests are made to the server.
		oauth2Client, err := s.oauth2ClientsService.RequestIsAuthenticated(req)
		if err != nil || oauth2Client == nil {

			// In the event there's not a valid OAuth2 token attached to the request, or there is some other OAuth2 issue,
			// we next check to see if a valid cookie is attached to the request
			cookieAuth, cerr := s.decodeCookieFromRequest(req)
			if cerr != nil || cookieAuth == nil {
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

// userCookieAuthenticationMiddleware checks for a user cookie
func (s *Server) userCookieAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		s.logger.Debug("userCookieAuthenticationMiddleware triggered")

		user, err := s.fetchUserFromRequest(req)
		if err != nil {
			s.logger.Error(err, "error encountered fetching user")
			s.internalServerError(res, req, err)
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

func (s *Server) buildRouteCtx(key models.ContextKey, x interface{}) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			if err := json.NewDecoder(req.Body).Decode(x); err != nil {
				s.logger.Error(err, "error encountered decoding request body")
				res.WriteHeader(http.StatusBadRequest)
				return
			}
			next.ServeHTTP(res, req.WithContext(context.WithValue(req.Context(), key, x)))
		})
	}
}

func (s *Server) tracingMiddleware(next http.Handler) http.Handler {
	return nethttp.Middleware(
		s.tracer,
		next,
		nethttp.MWComponentName("todo-server"),
		nethttp.MWSpanObserver(func(span opentracing.Span, req *http.Request) {
			span.SetTag("http.method", req.Method)
			span.SetTag("http.uri", req.URL.EscapedPath())
		}),
		nethttp.OperationNameFunc(func(req *http.Request) string {
			return fmt.Sprintf("%s %s", req.Method, req.URL.Path)
		}),
	)
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ww := middleware.NewWrapResponseWriter(res, req.ProtoMajor)

		start := time.Now()
		defer func() {
			s.logger.WithValues(map[string]interface{}{
				"status":        ww.Status(),
				"bytes_written": ww.BytesWritten(),
				"elapsed":       time.Since(start),
			})
		}()

		next.ServeHTTP(ww, req)
	})
}

func (s *Server) dualAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		oac := ctx.Value(oauth2clients.MiddlewareCtxKey)
		u := ctx.Value(users.MiddlewareCtxKey)

		x, ok1 := oac.(*models.OAuth2Client)
		y, ok2 := u.(*models.User)

		logger := s.logger.WithValues(map[string]interface{}{
			"oauth2_client":    x,
			"oauth2_client_ok": ok1,
			"user":             y,
			"user_ok":          ok2,
		})

		if (ok1 && x != nil) || (ok2 && y != nil) {
			logger.Debug("errythang good")
			next.ServeHTTP(res, req)
		} else {
			http.Redirect(res, req, "/login", http.StatusUnauthorized)
		}
	})
}
