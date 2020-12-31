package httpserver

import (
	"fmt"
	"net/http"

	auditservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/audit"
	itemsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/items"
	oauth2clientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/oauth2clients"
	plansservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/plans"
	usersservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/users"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/webhooks"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/heptiolabs/healthcheck"
)

const (
	root             = "/"
	auditRoute       = "/audit"
	searchRoot       = "/search"
	numericIDPattern = "{%s:[0-9]+}"
	maxCORSAge       = 300
)

func (s *Server) setupRouter(metricsHandler metrics.Handler) {
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
		middleware.RealIP,
		middleware.Timeout(maxTimeout),
		buildLoggingMiddleware(s.logger.WithName("middleware"), s.tracer),
		ch.Handler,
	)

	authenticatedRouter := mux.With(s.authService.UserAttributionMiddleware)

	// all middleware must be defined before routes on a mux.

	mux.Route("/_meta_", func(metaRouter chi.Router) {
		health := healthcheck.NewHandler()
		// Expose a liveness check on /live
		metaRouter.Get("/live", health.LiveEndpoint)
		// Expose a readiness check on /ready
		metaRouter.Get("/ready", health.ReadyEndpoint)
	})

	if metricsHandler != nil {
		s.logger.Debug("establishing metrics handler")
		mux.Handle("/metrics", metricsHandler)
	}

	// Frontend routes.
	if sfd := s.frontendSettings.StaticFilesDirectory; sfd != "" {
		staticFileServer, err := s.frontendService.StaticDir(sfd)
		if err != nil {
			s.logger.Error(err, "establishing static file server")
		}

		mux.Get("/*", staticFileServer)
	}

	authenticatedRouter.Get("/auth/status", s.authService.StatusHandler)

	mux.Route("/users", func(userRouter chi.Router) {
		userRouter.With(s.authService.UserLoginInputMiddleware).Post("/login", s.authService.LoginHandler)
		userRouter.With(s.authService.CookieAuthenticationMiddleware).Post("/logout", s.authService.LogoutHandler)
		userRouter.With(s.usersService.UserCreationInputMiddleware).Post(root, s.usersService.CreateHandler)
		userRouter.With(s.usersService.TOTPSecretVerificationInputMiddleware).Post("/totp_secret/verify", s.usersService.TOTPSecretVerificationHandler)

		// need creds beyond this point.
		authedRouter := userRouter.With(s.authService.UserAttributionMiddleware, s.authService.AuthorizationMiddleware)
		authedRouter.With(s.usersService.TOTPSecretRefreshInputMiddleware).Post("/totp_secret/new", s.usersService.NewTOTPSecretHandler)
		authedRouter.With(s.usersService.PasswordUpdateInputMiddleware).Put("/password/new", s.usersService.UpdatePasswordHandler)
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

	authenticatedRouter.With(s.authService.AuthorizationMiddleware).Route("/api/v1", func(v1Router chi.Router) {
		adminRouter := v1Router.With(s.authService.AdminMiddleware)

		// Users
		v1Router.Route("/users", func(usersRouter chi.Router) {
			singleUserRoute := fmt.Sprintf("/"+numericIDPattern, usersservice.UserIDURIParamKey)

			usersRouter.With(s.authService.AdminMiddleware).Get(root, s.usersService.ListHandler)
			usersRouter.With(s.authService.AdminMiddleware).Get("/search", s.usersService.UsernameSearchHandler)

			usersRouter.Get("/self", s.usersService.SelfHandler)
			usersRouter.Route(singleUserRoute, func(singleUserRouter chi.Router) {
				singleUserRouter.With(s.authService.AdminMiddleware).Get(root, s.usersService.ReadHandler)
				singleUserRouter.With(s.authService.AdminMiddleware).Get(auditRoute, s.usersService.AuditEntryHandler)

				singleUserRouter.Delete(root, s.usersService.ArchiveHandler)
			})
		})

		// Plans
		adminRouter.Route("/plans", func(plansRouter chi.Router) {
			singlePlanRoute := fmt.Sprintf("/"+numericIDPattern, plansservice.PlanIDURIParamKey)

			plansRouter.With(s.plansService.CreationInputMiddleware).Post(root, s.plansService.CreateHandler)
			plansRouter.Get(root, s.plansService.ListHandler)
			plansRouter.Route(singlePlanRoute, func(singlePlanRouter chi.Router) {
				singlePlanRouter.Get(root, s.plansService.ReadHandler)
				singlePlanRouter.Get(auditRoute, s.plansService.AuditEntryHandler)
				singlePlanRouter.Delete(root, s.plansService.ArchiveHandler)
				singlePlanRouter.With(s.plansService.UpdateInputMiddleware).Put(root, s.plansService.UpdateHandler)
			})
		})

		// Admin
		adminRouter.Route("/_admin_", func(adminRouter chi.Router) {
			adminRouter.Post("/cycle_cookie_secret", s.authService.CycleCookieSecretHandler)
			adminRouter.With(s.adminService.AccountStatusUpdateInputMiddleware).
				Post("/users/status", s.adminService.UserAccountStatusChangeHandler)

			adminRouter.Route("/audit_log", func(auditRouter chi.Router) {
				entryIDRouteParam := fmt.Sprintf("/"+numericIDPattern, auditservice.LogEntryURIParamKey)
				auditRouter.Get(root, s.auditService.ListHandler)
				auditRouter.Route(entryIDRouteParam, func(singleEntryRouter chi.Router) {
					singleEntryRouter.Get(root, s.auditService.ReadHandler)
				})
			})
		})

		// OAuth2 Clients.
		v1Router.Route("/oauth2/clients", func(clientRouter chi.Router) {
			singleClientRoute := fmt.Sprintf("/"+numericIDPattern, oauth2clientsservice.OAuth2ClientIDURIParamKey)
			clientRouter.Get(root, s.oauth2ClientsService.ListHandler)
			// CreateHandler does not require preexisting authentication.
			// UpdateHandler not supported for OAuth2 clients.

			clientRouter.Route(singleClientRoute, func(singleClientRouter chi.Router) {
				singleClientRouter.Get(root, s.oauth2ClientsService.ReadHandler)
				singleClientRouter.Delete(root, s.oauth2ClientsService.ArchiveHandler)
				singleClientRouter.With(s.authService.AdminMiddleware).Get(auditRoute, s.oauth2ClientsService.AuditEntryHandler)
			})
		})

		// Webhooks.
		v1Router.Route("/webhooks", func(webhookRouter chi.Router) {
			singleWebhookRoute := fmt.Sprintf("/"+numericIDPattern, webhooksservice.WebhookIDURIParamKey)
			webhookRouter.With(s.webhooksService.CreationInputMiddleware).Post(root, s.webhooksService.CreateHandler)
			webhookRouter.Get(root, s.webhooksService.ListHandler)

			webhookRouter.Route(singleWebhookRoute, func(singleWebhookRouter chi.Router) {
				singleWebhookRouter.Get(root, s.webhooksService.ReadHandler)
				singleWebhookRouter.Delete(root, s.webhooksService.ArchiveHandler)
				singleWebhookRouter.With(s.webhooksService.UpdateInputMiddleware).Put(root, s.webhooksService.UpdateHandler)
				singleWebhookRouter.With(s.authService.AdminMiddleware).Get(auditRoute, s.webhooksService.AuditEntryHandler)
			})
		})

		// Items
		itemPath := "items"
		itemsRouteWithPrefix := fmt.Sprintf("/%s", itemPath)
		itemIDRouteParam := fmt.Sprintf("/"+numericIDPattern, itemsservice.ItemIDURIParamKey)
		v1Router.Route(itemsRouteWithPrefix, func(itemsRouter chi.Router) {
			itemsRouter.With(s.itemsService.CreationInputMiddleware).Post(root, s.itemsService.CreateHandler)
			itemsRouter.Get(root, s.itemsService.ListHandler)
			itemsRouter.Get(searchRoot, s.itemsService.SearchHandler)

			itemsRouter.Route(itemIDRouteParam, func(singleItemRouter chi.Router) {
				singleItemRouter.Get(root, s.itemsService.ReadHandler)
				singleItemRouter.Head(root, s.itemsService.ExistenceHandler)
				singleItemRouter.Delete(root, s.itemsService.ArchiveHandler)
				singleItemRouter.With(s.itemsService.UpdateInputMiddleware).Put(root, s.itemsService.UpdateHandler)
				singleItemRouter.With(s.authService.AdminMiddleware).Get(auditRoute, s.itemsService.AuditEntryHandler)
			})
		})
	})

	s.router = mux
}

func (s *Server) logRoutes() {
	if err := chi.Walk(s.router, func(method string, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		s.logger.WithValues(map[string]interface{}{
			"method": method,
			"route":  route,
		}).Debug("route found")

		return nil
	}); err != nil {
		s.logger.Error(err, "logging routes")
	}
}
