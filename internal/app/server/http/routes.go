package httpserver

import (
	"fmt"
	"net/http"

	"github.com/heptiolabs/healthcheck"

	plansservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/accountsubscriptionplans"
	auditservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/audit"
	itemsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/items"
	oauth2clientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/oauth2clients"
	usersservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/users"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/webhooks"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing"
)

const (
	root             = "/"
	auditRoute       = "/audit"
	searchRoot       = "/search"
	numericIDPattern = "{%s:[0-9]+}"
)

func (s *Server) setupRouter(router routing.Router, metricsHandler metrics.Handler) {
	router.Route("/_meta_", func(metaRouter routing.Router) {
		health := healthcheck.NewHandler()
		// Expose a liveness check on /live
		metaRouter.Get("/live", health.LiveEndpoint)
		// Expose a readiness check on /ready
		metaRouter.Get("/ready", health.ReadyEndpoint)
	})

	if metricsHandler != nil {
		s.logger.Debug("establishing metrics handler")
		router.Handle("/metrics", metricsHandler)
	}

	// Frontend routes.
	if sfd := s.frontendSettings.StaticFilesDirectory; sfd != "" {
		staticFileServer, err := s.frontendService.StaticDir(sfd)
		if err != nil {
			s.logger.Error(err, "establishing static file server")
		}

		router.Get("/*", staticFileServer)
	}

	authenticatedRouter := router.WithMiddleware(s.authService.UserAttributionMiddleware)
	authenticatedRouter.Get("/auth/status", s.authService.StatusHandler)

	router.Route("/users", func(userRouter routing.Router) {
		userRouter.WithMiddleware(s.authService.UserLoginInputMiddleware).Post("/login", s.authService.LoginHandler)
		userRouter.WithMiddleware(s.authService.CookieAuthenticationMiddleware).Post("/logout", s.authService.LogoutHandler)
		userRouter.WithMiddleware(s.usersService.UserCreationInputMiddleware).Post(root, s.usersService.CreateHandler)
		userRouter.WithMiddleware(s.usersService.TOTPSecretVerificationInputMiddleware).Post("/totp_secret/verify", s.usersService.TOTPSecretVerificationHandler)

		// need creds beyond this point.
		authedRouter := userRouter.WithMiddleware(s.authService.UserAttributionMiddleware, s.authService.AuthorizationMiddleware)
		authedRouter.WithMiddleware(s.usersService.TOTPSecretRefreshInputMiddleware).Post("/totp_secret/new", s.usersService.NewTOTPSecretHandler)
		authedRouter.WithMiddleware(s.usersService.PasswordUpdateInputMiddleware).Put("/authentication/new", s.usersService.UpdatePasswordHandler)
	})

	router.Route("/oauth2", func(oauth2Router routing.Router) {
		oauth2Router.WithMiddleware(
			s.authService.CookieAuthenticationMiddleware,
			s.oauth2ClientsService.CreationInputMiddleware,
		).Post("/client", s.oauth2ClientsService.CreateHandler)

		oauth2Router.WithMiddleware(s.oauth2ClientsService.OAuth2ClientInfoMiddleware).
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

	authenticatedRouter.WithMiddleware(s.authService.AuthorizationMiddleware).Route("/api/v1", func(v1Router routing.Router) {
		adminRouter := v1Router.WithMiddleware(s.authService.AdminMiddleware)

		// Users
		v1Router.Route("/users", func(usersRouter routing.Router) {
			singleUserRoute := fmt.Sprintf("/"+numericIDPattern, usersservice.UserIDURIParamKey)

			usersRouter.WithMiddleware(s.authService.AdminMiddleware).Get(root, s.usersService.ListHandler)
			usersRouter.WithMiddleware(s.authService.AdminMiddleware).Get("/search", s.usersService.UsernameSearchHandler)
			usersRouter.WithMiddleware(s.usersService.AvatarUploadMiddleware).Post("/avatar/upload", s.usersService.AvatarUploadHandler)
			usersRouter.Get("/self", s.usersService.SelfHandler)
			usersRouter.Route(singleUserRoute, func(singleUserRouter routing.Router) {
				singleUserRouter.WithMiddleware(s.authService.AdminMiddleware).Get(root, s.usersService.ReadHandler)
				singleUserRouter.WithMiddleware(s.authService.AdminMiddleware).Get(auditRoute, s.usersService.AuditEntryHandler)

				singleUserRouter.Delete(root, s.usersService.ArchiveHandler)
			})
		})

		// AccountSubscriptionPlans
		adminRouter.Route("/accountsubscriptionplans", func(plansRouter routing.Router) {
			singlePlanRoute := fmt.Sprintf("/"+numericIDPattern, plansservice.PlanIDURIParamKey)

			plansRouter.WithMiddleware(s.plansService.CreationInputMiddleware).Post(root, s.plansService.CreateHandler)
			plansRouter.Get(root, s.plansService.ListHandler)
			plansRouter.Route(singlePlanRoute, func(singlePlanRouter routing.Router) {
				singlePlanRouter.Get(root, s.plansService.ReadHandler)
				singlePlanRouter.Get(auditRoute, s.plansService.AuditEntryHandler)
				singlePlanRouter.Delete(root, s.plansService.ArchiveHandler)
				singlePlanRouter.WithMiddleware(s.plansService.UpdateInputMiddleware).Put(root, s.plansService.UpdateHandler)
			})
		})

		// Admin
		adminRouter.Route("/_admin_", func(adminRouter routing.Router) {
			adminRouter.Post("/cycle_cookie_secret", s.authService.CycleCookieSecretHandler)
			adminRouter.WithMiddleware(s.adminService.AccountStatusUpdateInputMiddleware).
				Post("/users/status", s.adminService.UserAccountStatusChangeHandler)

			adminRouter.Route("/audit_log", func(auditRouter routing.Router) {
				entryIDRouteParam := fmt.Sprintf("/"+numericIDPattern, auditservice.LogEntryURIParamKey)
				auditRouter.Get(root, s.auditService.ListHandler)
				auditRouter.Route(entryIDRouteParam, func(singleEntryRouter routing.Router) {
					singleEntryRouter.Get(root, s.auditService.ReadHandler)
				})
			})
		})

		// OAuth2 Clients.
		v1Router.Route("/oauth2/clients", func(clientRouter routing.Router) {
			singleClientRoute := fmt.Sprintf("/"+numericIDPattern, oauth2clientsservice.OAuth2ClientIDURIParamKey)
			clientRouter.Get(root, s.oauth2ClientsService.ListHandler)

			clientRouter.Route(singleClientRoute, func(singleClientRouter routing.Router) {
				singleClientRouter.Get(root, s.oauth2ClientsService.ReadHandler)
				singleClientRouter.Delete(root, s.oauth2ClientsService.ArchiveHandler)
				singleClientRouter.WithMiddleware(s.authService.AdminMiddleware).Get(auditRoute, s.oauth2ClientsService.AuditEntryHandler)
			})
		})

		// Webhooks.
		v1Router.Route("/webhooks", func(webhookRouter routing.Router) {
			singleWebhookRoute := fmt.Sprintf("/"+numericIDPattern, webhooksservice.WebhookIDURIParamKey)
			webhookRouter.WithMiddleware(s.webhooksService.CreationInputMiddleware).Post(root, s.webhooksService.CreateHandler)
			webhookRouter.Get(root, s.webhooksService.ListHandler)

			webhookRouter.Route(singleWebhookRoute, func(singleWebhookRouter routing.Router) {
				singleWebhookRouter.Get(root, s.webhooksService.ReadHandler)
				singleWebhookRouter.Delete(root, s.webhooksService.ArchiveHandler)
				singleWebhookRouter.WithMiddleware(s.webhooksService.UpdateInputMiddleware).Put(root, s.webhooksService.UpdateHandler)
				singleWebhookRouter.WithMiddleware(s.authService.AdminMiddleware).Get(auditRoute, s.webhooksService.AuditEntryHandler)
			})
		})

		// Items
		itemPath := "items"
		itemsRouteWithPrefix := fmt.Sprintf("/%s", itemPath)
		itemIDRouteParam := fmt.Sprintf("/"+numericIDPattern, itemsservice.ItemIDURIParamKey)
		v1Router.Route(itemsRouteWithPrefix, func(itemsRouter routing.Router) {
			itemsRouter.WithMiddleware(s.itemsService.CreationInputMiddleware).Post(root, s.itemsService.CreateHandler)
			itemsRouter.Get(root, s.itemsService.ListHandler)
			itemsRouter.Get(searchRoot, s.itemsService.SearchHandler)

			itemsRouter.Route(itemIDRouteParam, func(singleItemRouter routing.Router) {
				singleItemRouter.Get(root, s.itemsService.ReadHandler)
				singleItemRouter.Head(root, s.itemsService.ExistenceHandler)
				singleItemRouter.Delete(root, s.itemsService.ArchiveHandler)
				singleItemRouter.WithMiddleware(s.itemsService.UpdateInputMiddleware).Put(root, s.itemsService.UpdateHandler)
				singleItemRouter.WithMiddleware(s.authService.AdminMiddleware).Get(auditRoute, s.itemsService.AuditEntryHandler)
			})
		})
	})

	s.router = router
}
