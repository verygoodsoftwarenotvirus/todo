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

func (s *Server) buildCookieMiddleware(rejectIfNotFound bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			logger := s.logger.WithValue("rejecting", rejectIfNotFound)
			logger.Debug("buildCookieMiddleware triggered")

			user, err := s.fetchUserFromRequest(req)
			if err != nil && rejectIfNotFound {
				logger.Error(err, "error encountered fetching user")
				s.internalServerError(res, req, err)
				return
			}

			if user != nil {
				req = req.WithContext(context.WithValue(
					context.WithValue(req.Context(), models.UserKey, user),
					models.UserIDKey,
					user.ID,
				))
			} else if rejectIfNotFound {
				logger.Debug("redirecting to login")
				http.Redirect(res, req, "/login", http.StatusUnauthorized)
				return
			}

			logger.Debug("moving onto next middleware from buildCookieMiddleware")
			next.ServeHTTP(res, req)
		})
	}
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
