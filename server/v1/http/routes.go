package httpserver

import (
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics"
	itemsservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	oauth2clientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	usersservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/webhooks"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/heptiolabs/healthcheck"
)

const (
	root             = "/"
	searchRoot       = "/search"
	numericIDPattern = "/{%s:[0-9]+}"
	maxCORSAge       = 300
)

func (s *Server) setupRouter(cfg *config.ServerConfig, metricsHandler metrics.Handler) {
	if cfg == nil {
		panic("config should not be nil")
	}

	// Basic CORS, for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	ch := cors.New(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts,
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Provider",
			"X-CSRF-Token",
		},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           maxCORSAge,
	})

	mux := chi.NewRouter()
	mux.Use(
		middleware.RequestID,
		middleware.Timeout(maxTimeout),
		buildLoggingMiddleware(s.logger.WithName("middleware")),
		ch.Handler,
	)

	plainRouter := mux.With()
	authenticatedRouter := mux.With(s.authService.AuthenticationMiddleware)

	// all middleware must be defined before routes on a mux.

	plainRouter.Route("/_meta_", func(metaRouter chi.Router) {
		health := healthcheck.NewHandler()
		// Expose a liveness check on /live
		metaRouter.Get("/live", health.LiveEndpoint)
		// Expose a readiness check on /ready
		metaRouter.Get("/ready", health.ReadyEndpoint)
	})

	if metricsHandler != nil {
		s.logger.Debug("establishing metrics handler")
		plainRouter.Handle("/metrics", metricsHandler)
	}

	// Frontend routes.
	if s.config.Frontend.StaticFilesDirectory != "" {
		staticFileServer, err := s.frontendService.StaticDir(cfg.Frontend.StaticFilesDirectory)
		if err != nil {
			s.logger.Error(err, "establishing static file server")
		}
		plainRouter.Get("/*", staticFileServer)
	}

	authenticatedRouter.Get("/auth/status", s.authService.StatusHandler)

	mux.With(
		s.authService.AuthorizationMiddleware(true),
		s.authService.AdminMiddleware,
	).Route("/_admin_", func(adminRouter chi.Router) {
		adminRouter.Post("/cycle_cookie_secret", s.authService.CycleSecretHandler)

		entryIDRouteParam := fmt.Sprintf(numericIDPattern, itemsservice.URIParamKey)
		adminRouter.Get(entryIDRouteParam, s.auditService.ReadHandler)

		adminRouter.Get("/audit_log", s.auditService.ListHandler)
	})

	mux.Route("/users", func(userRouter chi.Router) {
		userRouter.With(s.authService.UserLoginInputMiddleware).Post("/login", s.authService.LoginHandler)
		userRouter.With(s.authService.CookieAuthenticationMiddleware).Post("/logout", s.authService.LogoutHandler)
		userRouter.With(s.usersService.UserInputMiddleware).Post(root, s.usersService.CreateHandler)
		userRouter.With(s.usersService.TOTPSecretVerificationInputMiddleware).Post("/totp_secret/verify", s.usersService.TOTPSecretVerificationHandler)
		userRouter.With(
			s.authService.AuthorizationMiddleware(true),
			s.usersService.TOTPSecretRefreshInputMiddleware,
		).Post("/totp_secret/new", s.usersService.NewTOTPSecretHandler)
		userRouter.With(
			s.authService.AuthorizationMiddleware(true),
			s.usersService.PasswordUpdateInputMiddleware,
		).Put("/password/new", s.usersService.UpdatePasswordHandler)
	})

	mux.Route("/oauth2", func(oauth2Router chi.Router) {
		oauth2Router.With(
			s.authService.CookieAuthenticationMiddleware,
			s.oauth2ClientsService.CreationInputMiddleware,
		).Post("/client", s.oauth2ClientsService.CreateHandler)

		oauth2Router.With(s.oauth2ClientsService.OAuth2ClientInfoMiddleware).
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

	mux.With(s.authService.AuthorizationMiddleware(true)).
		Route("/api/v1", func(v1Router chi.Router) {
			// Users
			v1Router.Route("/users", func(usersRouter chi.Router) {
				userIDPattern := fmt.Sprintf(numericIDPattern, usersservice.URIParamKey)

				usersRouter.With(s.authService.AdminMiddleware).Get(root, s.usersService.ListHandler)
				usersRouter.Get("/self", s.usersService.SelfHandler)
				usersRouter.With(s.authService.AdminMiddleware).Get(userIDPattern, s.usersService.ReadHandler)
				usersRouter.Delete(userIDPattern, s.usersService.ArchiveHandler)
			})

			// OAuth2 Clients.
			v1Router.Route("/oauth2/clients", func(clientRouter chi.Router) {
				sr := fmt.Sprintf(numericIDPattern, oauth2clientsservice.URIParamKey)
				// CreateHandler is not bound to an OAuth2 authentication token.
				// UpdateHandler not supported for OAuth2 clients.
				clientRouter.Get(sr, s.oauth2ClientsService.ReadHandler)
				clientRouter.Delete(sr, s.oauth2ClientsService.ArchiveHandler)
				clientRouter.Get(root, s.oauth2ClientsService.ListHandler)
			})

			// Webhooks.
			v1Router.Route("/webhooks", func(webhookRouter chi.Router) {
				sr := fmt.Sprintf(numericIDPattern, webhooksservice.URIParamKey)
				webhookRouter.With(s.webhooksService.CreationInputMiddleware).Post(root, s.webhooksService.CreateHandler)
				webhookRouter.Get(sr, s.webhooksService.ReadHandler)
				webhookRouter.With(s.webhooksService.UpdateInputMiddleware).Put(sr, s.webhooksService.UpdateHandler)
				webhookRouter.Delete(sr, s.webhooksService.ArchiveHandler)
				webhookRouter.Get(root, s.webhooksService.ListHandler)
			})

			// Items
			itemPath := "items"
			itemsRouteWithPrefix := fmt.Sprintf("/%s", itemPath)
			itemIDRouteParam := fmt.Sprintf(numericIDPattern, itemsservice.URIParamKey)
			v1Router.Route(itemsRouteWithPrefix, func(itemsRouter chi.Router) {
				itemsRouter.With(s.itemsService.CreationInputMiddleware).Post(root, s.itemsService.CreateHandler)
				itemsRouter.Route(itemIDRouteParam, func(singleItemRouter chi.Router) {
					singleItemRouter.Get(root, s.itemsService.ReadHandler)
					singleItemRouter.With(s.itemsService.UpdateInputMiddleware).Put(root, s.itemsService.UpdateHandler)
					singleItemRouter.Delete(root, s.itemsService.ArchiveHandler)
					singleItemRouter.Head(root, s.itemsService.ExistenceHandler)
				})
				itemsRouter.Get(root, s.itemsService.ListHandler)
				itemsRouter.Get(searchRoot, s.itemsService.SearchHandler)
			})
		})

	s.router = mux
}
