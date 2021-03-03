package httpserver

import (
	"fmt"
	"net/http"

	"github.com/heptiolabs/healthcheck"

	accountsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/accounts"
	plansservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/accountsubscriptionplans"
	apiclientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/apiclients"
	auditservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/audit"
	itemsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/items"
	usersservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/users"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/webhooks"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing"
)

const (
	root             = "/"
	auditRoute       = "/audit"
	searchRoot       = "/search"
	numericIDPattern = "{%s:[0-9]+}"
)

func buildNumericIDURLChunk(key string) string {
	return fmt.Sprintf("/"+numericIDPattern, key)
}

func buildTokenRestrictionMiddleware(logger logging.Logger, token string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			if token != "" {
				if req.Header.Get("Authorization") != token {
					logger.Info("rejected unauthorized metrics scrape")
					res.WriteHeader(http.StatusUnauthorized)
					return
				}
			}

			next.ServeHTTP(res, req)
		})
	}
}

func (s *Server) setupRouter(router routing.Router, metricsConfig metrics.Config, metricsHandler metrics.Handler) {
	router.Route("/_meta_", func(metaRouter routing.Router) {
		health := healthcheck.NewHandler()
		// Expose a liveness check on /live
		metaRouter.Get("/live", health.LiveEndpoint)
		// Expose a readiness check on /ready
		metaRouter.Get("/ready", health.ReadyEndpoint)
	})

	if metricsHandler != nil {
		s.logger.Debug("establishing metrics handler")
		router.WithMiddleware(buildTokenRestrictionMiddleware(s.logger, metricsConfig.RouteToken)).Handle("/metrics", metricsHandler)
	}

	// Frontend routes.
	if sfd := s.frontendSettings.StaticFilesDirectory; sfd != "" {
		staticFileServer, err := s.frontendService.StaticDir(sfd)
		if err != nil {
			s.logger.Error(err, "establishing static file server")
		}

		router.Get("/*", staticFileServer)
	}

	router.WithMiddleware(s.authService.PASETOCreationInputMiddleware).Post("/paseto", s.authService.PASETOHandler)

	authenticatedRouter := router.WithMiddleware(s.authService.UserAttributionMiddleware)
	authenticatedRouter.Get("/auth/status", s.authService.StatusHandler)

	router.Route("/users", func(userRouter routing.Router) {
		userRouter.WithMiddleware(s.authService.UserLoginInputMiddleware).Post("/login", s.authService.LoginHandler)
		userRouter.WithMiddleware(s.authService.CookieAuthenticationMiddleware).Post("/logout", s.authService.LogoutHandler)
		userRouter.WithMiddleware(s.usersService.UserCreationInputMiddleware).Post(root, s.usersService.CreateHandler)
		userRouter.WithMiddleware(s.usersService.TOTPSecretVerificationInputMiddleware).Post("/totp_secret/verify", s.usersService.TOTPSecretVerificationHandler)

		// need credentials beyond this point
		authedRouter := userRouter.WithMiddleware(s.authService.UserAttributionMiddleware, s.authService.AuthorizationMiddleware)
		authedRouter.WithMiddleware(s.usersService.TOTPSecretRefreshInputMiddleware).Post("/totp_secret/new", s.usersService.NewTOTPSecretHandler)
		authedRouter.WithMiddleware(s.usersService.PasswordUpdateInputMiddleware).Put("/password/new", s.usersService.UpdatePasswordHandler)
	})

	authenticatedRouter.WithMiddleware(s.authService.AuthorizationMiddleware).Route("/api/v1", func(v1Router routing.Router) {
		adminRouter := v1Router.WithMiddleware(s.authService.AdminMiddleware)

		// Account Subscription Plans
		adminRouter.Route("/account_subscription_plans", func(plansRouter routing.Router) {
			plansRouter.Get(root, s.plansService.ListHandler)
			plansRouter.WithMiddleware(s.plansService.CreationInputMiddleware).Post(root, s.plansService.CreateHandler)

			singlePlanRoute := buildNumericIDURLChunk(plansservice.AccountSubscriptionPlanIDURIParamKey)
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
				entryIDRouteParam := buildNumericIDURLChunk(auditservice.LogEntryURIParamKey)
				auditRouter.Get(root, s.auditService.ListHandler)
				auditRouter.Route(entryIDRouteParam, func(singleEntryRouter routing.Router) {
					singleEntryRouter.Get(root, s.auditService.ReadHandler)
				})
			})
		})

		// Users
		v1Router.Route("/users", func(usersRouter routing.Router) {
			usersRouter.WithMiddleware(s.authService.AdminMiddleware).Get(root, s.usersService.ListHandler)
			usersRouter.WithMiddleware(s.authService.AdminMiddleware).Get("/search", s.usersService.UsernameSearchHandler)
			usersRouter.WithMiddleware(s.usersService.AvatarUploadMiddleware).Post("/avatar/upload", s.usersService.AvatarUploadHandler)
			usersRouter.Get("/self", s.usersService.SelfHandler)

			singleUserRoute := buildNumericIDURLChunk(usersservice.UserIDURIParamKey)
			usersRouter.Route(singleUserRoute, func(singleUserRouter routing.Router) {
				singleUserRouter.WithMiddleware(s.authService.AdminMiddleware).Get(root, s.usersService.ReadHandler)
				singleUserRouter.WithMiddleware(s.authService.AdminMiddleware).Get(auditRoute, s.usersService.AuditEntryHandler)

				singleUserRouter.Delete(root, s.usersService.ArchiveHandler)
			})
		})

		// Accounts
		v1Router.Route("/accounts", func(accountsRouter routing.Router) {
			accountsRouter.WithMiddleware(s.accountsService.CreationInputMiddleware).Post(root, s.accountsService.CreateHandler)
			accountsRouter.Get(root, s.accountsService.ListHandler)

			singleUserRoute := buildNumericIDURLChunk(accountsservice.UserIDURIParamKey)
			singleAccountRoute := buildNumericIDURLChunk(accountsservice.AccountIDURIParamKey)
			accountsRouter.Route(singleAccountRoute, func(singleAccountRouter routing.Router) {
				singleAccountRouter.Get(root, s.accountsService.ReadHandler)
				singleAccountRouter.WithMiddleware(s.accountsService.UpdateInputMiddleware).Put(root, s.accountsService.UpdateHandler)
				singleAccountRouter.Delete(root, s.accountsService.ArchiveHandler)
				singleAccountRouter.WithMiddleware(s.authService.AdminMiddleware).Get(auditRoute, s.accountsService.AuditEntryHandler)

				singleAccountRouter.WithMiddleware(s.accountsService.AddMemberInputMiddleware).
					Post("/member", s.accountsService.AddUserHandler)
				singleAccountRouter.Post("/default", s.accountsService.MarkAsDefaultHandler)
				singleAccountRouter.WithMiddleware(s.accountsService.ModifyMemberPermissionsInputMiddleware).
					Patch("/members"+singleUserRoute+"/permissions", s.accountsService.ModifyMemberPermissionsHandler)
				singleAccountRouter.Delete("/members"+singleUserRoute, s.accountsService.RemoveUserHandler)
				singleAccountRouter.WithMiddleware(s.accountsService.AccountTransferInputMiddleware).
					Post("/transfer", s.accountsService.TransferAccountOwnershipHandler)
			})
		})

		// API Clients
		v1Router.Route("/api_clients", func(clientRouter routing.Router) {
			clientRouter.Get(root, s.apiClientsService.ListHandler)
			clientRouter.WithMiddleware(s.authService.PermissionRestrictionMiddleware(permissions.CanCreateAPIClients), s.apiClientsService.CreationInputMiddleware).Post(root, s.apiClientsService.CreateHandler)

			singleClientRoute := buildNumericIDURLChunk(apiclientsservice.APIClientIDURIParamKey)
			clientRouter.Route(singleClientRoute, func(singleClientRouter routing.Router) {
				singleClientRouter.WithMiddleware(s.authService.PermissionRestrictionMiddleware(permissions.CanReadAPIClients)).Get(root, s.apiClientsService.ReadHandler)
				singleClientRouter.WithMiddleware(s.authService.PermissionRestrictionMiddleware(permissions.CanArchiveAPIClients)).Delete(root, s.apiClientsService.ArchiveHandler)
				singleClientRouter.WithMiddleware(s.authService.AdminMiddleware).Get(auditRoute, s.apiClientsService.AuditEntryHandler)
			})
		})

		// Webhooks
		v1Router.Route("/webhooks", func(webhookRouter routing.Router) {
			webhookRouter.WithMiddleware(
				s.authService.PermissionRestrictionMiddleware(permissions.CanCreateWebhooks),
				s.webhooksService.CreationInputMiddleware,
			).Post(root, s.webhooksService.CreateHandler)
			webhookRouter.Get(root, s.webhooksService.ListHandler)

			singleWebhookRoute := buildNumericIDURLChunk(webhooksservice.WebhookIDURIParamKey)
			webhookRouter.Route(singleWebhookRoute, func(singleWebhookRouter routing.Router) {
				singleWebhookRouter.WithMiddleware(s.authService.PermissionRestrictionMiddleware(permissions.CanReadWebhooks)).
					Get(root, s.webhooksService.ReadHandler)
				singleWebhookRouter.WithMiddleware(s.authService.PermissionRestrictionMiddleware(permissions.CanArchiveWebhooks)).
					Delete(root, s.webhooksService.ArchiveHandler)
				singleWebhookRouter.WithMiddleware(
					s.authService.PermissionRestrictionMiddleware(permissions.CanUpdateWebhooks),
					s.webhooksService.UpdateInputMiddleware,
				).Put(root, s.webhooksService.UpdateHandler)
				singleWebhookRouter.WithMiddleware(s.authService.AdminMiddleware).
					Get(auditRoute, s.webhooksService.AuditEntryHandler)
			})
		})

		// Items
		itemPath := "items"
		itemsRouteWithPrefix := fmt.Sprintf("/%s", itemPath)
		itemIDRouteParam := buildNumericIDURLChunk(itemsservice.ItemIDURIParamKey)
		v1Router.Route(itemsRouteWithPrefix, func(itemsRouter routing.Router) {
			itemsRouter.WithMiddleware(s.itemsService.CreationInputMiddleware).
				Post(root, s.itemsService.CreateHandler)
			itemsRouter.Get(root, s.itemsService.ListHandler)
			itemsRouter.Get(searchRoot, s.itemsService.SearchHandler)

			itemsRouter.Route(itemIDRouteParam, func(singleItemRouter routing.Router) {
				singleItemRouter.WithMiddleware(s.authService.PermissionRestrictionMiddleware(permissions.CanReadItems)).
					Get(root, s.itemsService.ReadHandler)
				singleItemRouter.WithMiddleware(s.authService.PermissionRestrictionMiddleware(permissions.CanReadItems)).
					Head(root, s.itemsService.ExistenceHandler)
				singleItemRouter.WithMiddleware(s.authService.PermissionRestrictionMiddleware(permissions.CanArchiveItems)).
					Delete(root, s.itemsService.ArchiveHandler)
				singleItemRouter.WithMiddleware(
					s.authService.PermissionRestrictionMiddleware(permissions.CanUpdateItems),
					s.itemsService.UpdateInputMiddleware,
				).Put(root, s.itemsService.UpdateHandler)
				singleItemRouter.WithMiddleware(s.authService.AdminMiddleware).
					Get(auditRoute, s.itemsService.AuditEntryHandler)
			})
		})
	})

	s.router = router
}
