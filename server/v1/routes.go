package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/metrics/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	gcontext "github.com/gorilla/context"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
)

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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		start := time.Now()
		defer func() {
			s.logger.WithValues(map[string]interface{}{
				"status":        ww.Status(),
				"bytes_written": ww.BytesWritten(),
				"elapsed":       time.Since(start),
			})
		}()

		next.ServeHTTP(ww, r)
	})
}

func (s *Server) setupRouter(metricsHandler metrics.Handler, createClientInitRoute bool) {
	s.router = chi.NewRouter()

	s.router.Use(
		gcontext.ClearHandler,
		middleware.RequestID,
		middleware.Timeout(maxTimeout),
		s.tracingMiddleware,
		s.loggingMiddleware,
		//logging.BuildMiddleware(s.logger),
	)
	// all middlewares must be defined before routes on a mux

	s.router.Get("/_meta_/health", func(res http.ResponseWriter, req *http.Request) { res.WriteHeader(http.StatusOK) })

	if metricsHandler != nil {
		s.logger.Debug("setting metrics handler")
		s.router.Handle("/metrics", metricsHandler)
	}

	s.router.Route("/users", func(userRouter chi.Router) {
		userRouter.With(s.usersService.UserLoginInputContextMiddleware).Post("/login", s.login)
		userRouter.With(s.userCookieAuthenticationMiddleware).Post("/logout", s.logout)

		userRouter.With(
			s.userCookieAuthenticationMiddleware,
			s.usersService.TOTPSecretRefreshInputContextMiddleware,
		).Post("/totp_secret/new", s.usersService.NewTOTPSecret)

		userRouter.With(
			s.userCookieAuthenticationMiddleware,
			s.usersService.PasswordUpdateInputContextMiddleware,
		).Post("/password/new", s.usersService.UpdatePassword)

		usernamePattern := fmt.Sprintf("/{%s:[a-zA-Z0-9]+}", users.URIParamKey)

		userRouter.Get("/", s.usersService.List)                  // List
		userRouter.Get(usernamePattern, s.usersService.Read)      // Read
		userRouter.Delete(usernamePattern, s.usersService.Delete) // Delete
		userRouter.With(s.usersService.UserInputContextMiddleware).
			Post("/", s.usersService.Create) // Create
		// TODO: Update
	})

	s.router.Route("/oauth2", func(oauth2Router chi.Router) {
		if createClientInitRoute {
			s.logger.Debug("creating OAuth2 client initialization route")
			oauth2Router.
				With(
					s.userCookieAuthenticationMiddleware,
					s.oauth2ClientsService.OAuth2ClientCreationInputContextMiddleware,
				).
				Post("/init_client", s.oauth2ClientsService.Create)
		}

		oauth2Router.
			With(s.oauth2ClientsService.OAuth2ClientInfoMiddleware).
			Post("/authorize", func(res http.ResponseWriter, req *http.Request) {
				if err := s.oauth2ClientsService.HandleAuthorizeRequest(res, req); err != nil {
					http.Error(res, err.Error(), http.StatusBadRequest)
				}
			})

		oauth2Router.Post("/token", func(res http.ResponseWriter, req *http.Request) {
			if err := s.oauth2ClientsService.HandleTokenRequest(res, req); err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
			}
		})
	})

	s.router.With(s.oauth2ClientsService.OAuth2TokenAuthenticationMiddleware).Route("/api", func(apiRouter chi.Router) {
		apiRouter.Route("/v1", func(v1Router chi.Router) {

			v1Router.Route("/items", func(itemsRouter chi.Router) {
				sr := fmt.Sprintf("/{%s:[0-9]+}", items.URIParamKey)
				itemsRouter.Get("/", s.itemsService.List)       // List
				itemsRouter.Get("/count", s.itemsService.Count) // Count
				itemsRouter.Get(sr, s.itemsService.Read)        // Read
				itemsRouter.Delete(sr, s.itemsService.Delete)   // Delete
				itemsRouter.With(s.itemsService.ItemInputMiddleware).
					Put(sr, s.itemsService.Update) // Update
				itemsRouter.With(s.itemsService.ItemInputMiddleware).
					Post("/", s.itemsService.Create) // Create
			})

			v1Router.Route("/oauth2", func(oauth2Router chi.Router) {
				oauth2Router.Route("/clients", func(clientRouter chi.Router) {
					sr := fmt.Sprintf("/{%s}", oauth2clients.URIParamKey)
					clientRouter.Get("/", s.oauth2ClientsService.List)     // List
					clientRouter.Get(sr, s.oauth2ClientsService.Read)      // Read
					clientRouter.Delete(sr, s.oauth2ClientsService.Delete) // Delete
					clientRouter.
						With(s.buildRouteCtx(oauth2clients.MiddlewareCtxKey, new(models.OAuth2ClientUpdateInput))).
						Put(sr, s.oauth2ClientsService.Update) // Update
					clientRouter.
						With(s.buildRouteCtx(oauth2clients.MiddlewareCtxKey, new(models.OAuth2ClientCreationInput))).
						Post("/", s.oauth2ClientsService.Create) // Create
				})
			})

		})

	})

}
