package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
				s.logger.Errorf("error encountered decoding request body: %v", err)
				res.WriteHeader(http.StatusBadRequest)
				return
			}
			next.ServeHTTP(res, req.WithContext(context.WithValue(req.Context(), key, x)))
		})
	}
}

func (s *Server) buildTracingMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
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
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	f := &middleware.DefaultLogFormatter{Logger: s.logger, NoColor: true}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entry := f.NewLogEntry(r)
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		start := time.Now()
		defer func() {
			entry.Write(ww.Status(), ww.BytesWritten(), time.Since(start))
		}()

		next.ServeHTTP(ww, middleware.WithLogEntry(r, entry))
	})
}

func (s *Server) setupRoutes() {
	s.router = chi.NewRouter()

	s.router.Use(
		gcontext.ClearHandler,
		middleware.RequestID,
		middleware.DefaultLogger,
		middleware.Timeout(maxTimeout),
		s.buildTracingMiddleware(),
	)

	if s.DebugMode {
		s.router.Use(middleware.SetHeader("Access-Control-Allow-Origin", "*"))
		s.router.Get("/_debug_/stats", s.stats)
	}

	s.router.Get("/_meta_/health", func(res http.ResponseWriter, req *http.Request) { res.WriteHeader(http.StatusOK) })

	s.router.Route("/users", func(userRouter chi.Router) {
		userRouter.With(s.usersService.UserLoginInputContextMiddleware).Post("/login", s.Login)
		userRouter.With(s.UserCookieAuthenticationMiddleware).Post("/logout", s.Logout)

		userRouter.With(
			s.UserCookieAuthenticationMiddleware,
			s.usersService.TOTPSecretRefreshInputContextMiddleware,
		).Post("/totp_secret/new", s.usersService.NewTOTPSecret)

		userRouter.With(
			s.UserCookieAuthenticationMiddleware,
			s.usersService.PasswordUpdateInputContextMiddleware,
		).Post("/password/new", s.usersService.UpdatePassword)

		usernamePattern := fmt.Sprintf("/{%s:[a-zA-Z0-9]+}", users.URIParamKey)

		userRouter.Get("/", s.usersService.List)                  // List
		userRouter.Get(usernamePattern, s.usersService.Read)      // Read
		userRouter.Delete(usernamePattern, s.usersService.Delete) // Delete
		userRouter.With(s.usersService.UserInputContextMiddleware).
			Post("/", s.usersService.Create) // Create
		// userRouter.With(s.usersService.UserInputContextMiddleware).Put(sr, s.usersService.Update)   // Update
	})

	s.router.Route("/oauth2", func(oauth2Router chi.Router) {
		oauth2Router.
			With(s.OAuth2ClientInfoMiddleware).
			Post("/authorize", func(res http.ResponseWriter, req *http.Request) {
				if err := s.oauth2Handler.HandleAuthorizeRequest(res, req); err != nil {
					http.Error(res, err.Error(), http.StatusBadRequest)
				}
			})

		oauth2Router.Post("/token", func(res http.ResponseWriter, req *http.Request) {
			if err := s.oauth2Handler.HandleTokenRequest(res, req); err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
			}
		})
	})

	s.router.
		With(s.OauthTokenAuthenticationMiddleware).
		Route("/api", func(apiRouter chi.Router) {
			apiRouter.Route("/v1", func(v1Router chi.Router) {

				v1Router.Route("/items", func(itemsRouter chi.Router) {
					sr := fmt.Sprintf("/{%s:[0-9]+}", items.URIParamKey)
					itemsRouter.Get("/", s.itemsService.List)                                   // List
					itemsRouter.Get("/count", s.itemsService.Count)                             // Count
					itemsRouter.Get(sr, s.itemsService.BuildReadHandler(chiItemIDFetcher))      // Read
					itemsRouter.Delete(sr, s.itemsService.BuildDeleteHandler(chiItemIDFetcher)) // Delete
					itemsRouter.With(s.itemsService.ItemInputMiddleware).
						Put(sr, s.itemsService.BuildUpdateHandler(chiItemIDFetcher)) // Update
					itemsRouter.With(s.itemsService.ItemInputMiddleware).
						Post("/", s.itemsService.Create) // Create
				})

				v1Router.Route("/oauth2", func(oauth2Router chi.Router) {
					oauth2Router.Route("/clients", func(clientRouter chi.Router) {
						sr := fmt.Sprintf("/{%s}", oauth2clients.URIParamKey)
						clientRouter.Get("/", s.oauth2ClientsService.List)                                           // List
						clientRouter.Get(sr, s.oauth2ClientsService.BuildReadHandler(chiOAuth2ClientIDFetcher))      // Read
						clientRouter.Delete(sr, s.oauth2ClientsService.BuildDeleteHandler(chiOAuth2ClientIDFetcher)) // Delete
						clientRouter.
							With(s.buildRouteCtx(oauth2clients.MiddlewareCtxKey, new(models.OAuth2ClientUpdateInput))).
							Put(sr, s.oauth2ClientsService.BuildUpdateHandler(chiOAuth2ClientIDFetcher)) // Update
						clientRouter.
							With(s.buildRouteCtx(oauth2clients.MiddlewareCtxKey, new(models.OAuth2ClientCreationInput))).
							Post("/", s.oauth2ClientsService.Create) // Create
					})
				})

			})
		})
}
