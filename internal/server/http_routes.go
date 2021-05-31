package server

import (
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/accounts"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/apiclients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/users"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/webhooks"

	"github.com/heptiolabs/healthcheck"
)

const (
	root             = "/"
	auditRoute       = "/audit"
	searchRoot       = "/search"
	numericIDPattern = "{%s:[0-9]+}"
)

func buildNumericIDURLChunk(key string) string {
	return fmt.Sprintf(root+numericIDPattern, key)
}

func (s *HTTPServer) setupRouter(ctx context.Context, router routing.Router, metricsHandler metrics.Handler) {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	router.Route("/_meta_", func(metaRouter routing.Router) {
		health := healthcheck.NewHandler()
		// Expose a liveness check on /live
		metaRouter.Get("/live", health.LiveEndpoint)
		// Expose a readiness check on /ready
		metaRouter.Get("/ready", health.ReadyEndpoint)
	})

	if metricsHandler != nil {
		s.logger.Debug("establishing metrics handler")
		router.HandleFunc("/metrics", metricsHandler.ServeHTTP)
	}

	// Frontend routes.
	s.frontendService.SetupRoutes(router)

	router.Post("/paseto", s.authService.PASETOHandler)

	authenticatedRouter := router.WithMiddleware(s.authService.UserAttributionMiddleware)
	authenticatedRouter.Get("/auth/status", s.authService.StatusHandler)

	router.Route("/users", func(userRouter routing.Router) {
		userRouter.Post("/login", s.authService.BeginSessionHandler)
		userRouter.WithMiddleware(s.authService.UserAttributionMiddleware, s.authService.CookieRequirementMiddleware).Post("/logout", s.authService.EndSessionHandler)
		userRouter.Post(root, s.usersService.CreateHandler)
		userRouter.Post("/totp_secret/verify", s.usersService.TOTPSecretVerificationHandler)

		// need credentials beyond this point
		authedRouter := userRouter.WithMiddleware(s.authService.UserAttributionMiddleware, s.authService.AuthorizationMiddleware)
		authedRouter.Post("/account/select", s.authService.ChangeActiveAccountHandler)
		authedRouter.Post("/totp_secret/new", s.usersService.NewTOTPSecretHandler)
		authedRouter.Put("/password/new", s.usersService.UpdatePasswordHandler)
	})

	authenticatedRouter.WithMiddleware(s.authService.AuthorizationMiddleware).Route("/api/v1", func(v1Router routing.Router) {
		adminRouter := v1Router.WithMiddleware(s.authService.ServiceAdminMiddleware)

		// Admin
		adminRouter.Route("/admin", func(adminRouter routing.Router) {
			adminRouter.
				WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.CycleCookieSecretPermission)).
				Post("/cycle_cookie_secret", s.authService.CycleCookieSecretHandler)
			adminRouter.
				WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.UpdateUserStatusPermission)).
				Post("/users/status", s.adminService.UserReputationChangeHandler)

			adminRouter.Route("/audit_log", func(auditRouter routing.Router) {
				entryIDRouteParam := buildNumericIDURLChunk(audit.LogEntryURIParamKey)
				auditRouter.
					WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadAllAuditLogEntriesPermission)).
					Get(root, s.auditService.ListHandler)
				auditRouter.Route(entryIDRouteParam, func(singleEntryRouter routing.Router) {
					singleEntryRouter.
						WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadAllAuditLogEntriesPermission)).
						Get(root, s.auditService.ReadHandler)
				})
			})
		})

		// Users
		v1Router.Route("/users", func(usersRouter routing.Router) {
			usersRouter.
				WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadUserPermission)).
				Get(root, s.usersService.ListHandler)
			usersRouter.
				WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.SearchUserPermission)).
				Get("/search", s.usersService.UsernameSearchHandler)
			usersRouter.Post("/avatar/upload", s.usersService.AvatarUploadHandler)
			usersRouter.Get("/self", s.usersService.SelfHandler)

			singleUserRoute := buildNumericIDURLChunk(users.UserIDURIParamKey)
			usersRouter.Route(singleUserRoute, func(singleUserRouter routing.Router) {
				singleUserRouter.
					WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadUserPermission)).
					Get(root, s.usersService.ReadHandler)
				singleUserRouter.
					WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadUserAuditLogEntriesPermission)).
					Get(auditRoute, s.usersService.AuditEntryHandler)

				singleUserRouter.Delete(root, s.usersService.ArchiveHandler)
			})
		})

		// Accounts
		v1Router.Route("/accounts", func(accountsRouter routing.Router) {
			accountsRouter.Post(root, s.accountsService.CreateHandler)
			accountsRouter.Get(root, s.accountsService.ListHandler)

			singleUserRoute := buildNumericIDURLChunk(accounts.UserIDURIParamKey)
			singleAccountRoute := buildNumericIDURLChunk(accounts.AccountIDURIParamKey)
			accountsRouter.Route(singleAccountRoute, func(singleAccountRouter routing.Router) {
				singleAccountRouter.Get(root, s.accountsService.ReadHandler)
				singleAccountRouter.Put(root, s.accountsService.UpdateHandler)
				singleAccountRouter.
					WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ArchiveAccountPermission)).
					Delete(root, s.accountsService.ArchiveHandler)
				singleAccountRouter.
					WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadAccountAuditLogEntriesPermission)).
					Get(auditRoute, s.accountsService.AuditEntryHandler)

				singleAccountRouter.Post("/default", s.accountsService.MarkAsDefaultAccountHandler)
				singleAccountRouter.
					WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.RemoveMemberAccountPermission)).
					Delete("/members"+singleUserRoute, s.accountsService.RemoveMemberHandler)
				singleAccountRouter.
					WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.AddMemberAccountPermission)).
					Post("/member", s.accountsService.AddMemberHandler)
				singleAccountRouter.
					WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ModifyMemberPermissionsForAccountPermission)).
					Patch("/members"+singleUserRoute+"/permissions", s.accountsService.ModifyMemberPermissionsHandler)
				singleAccountRouter.Post("/transfer", s.accountsService.TransferAccountOwnershipHandler)
			})
		})

		// API Clients
		v1Router.Route("/api_clients", func(clientRouter routing.Router) {
			clientRouter.
				WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadAPIClientsPermission)).
				Get(root, s.apiClientsService.ListHandler)
			clientRouter.
				WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.CreateAPIClientsPermission)).
				Post(root, s.apiClientsService.CreateHandler)

			singleClientRoute := buildNumericIDURLChunk(apiclients.APIClientIDURIParamKey)
			clientRouter.Route(singleClientRoute, func(singleClientRouter routing.Router) {
				singleClientRouter.
					WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadAPIClientsPermission)).
					Get(root, s.apiClientsService.ReadHandler)
				singleClientRouter.
					WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ArchiveAPIClientsPermission)).
					Delete(root, s.apiClientsService.ArchiveHandler)
				singleClientRouter.
					WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadAPIClientAuditLogEntriesPermission)).
					Get(auditRoute, s.apiClientsService.AuditEntryHandler)
			})
		})

		// Webhooks
		v1Router.Route("/webhooks", func(webhookRouter routing.Router) {
			singleWebhookRoute := buildNumericIDURLChunk(webhooks.WebhookIDURIParamKey)
			webhookRouter.
				WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadWebhooksPermission)).
				Get(root, s.webhooksService.ListHandler)

			webhookRouter.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.CreateWebhooksPermission)).
				Post(root, s.webhooksService.CreateHandler)

			webhookRouter.Route(singleWebhookRoute, func(singleWebhookRouter routing.Router) {
				singleWebhookRouter.
					WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadWebhooksPermission)).
					Get(root, s.webhooksService.ReadHandler)

				singleWebhookRouter.
					WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ArchiveWebhooksPermission)).
					Delete(root, s.webhooksService.ArchiveHandler)
				singleWebhookRouter.
					WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.UpdateWebhooksPermission)).
					Put(root, s.webhooksService.UpdateHandler)
				singleWebhookRouter.
					WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadWebhooksAuditLogEntriesPermission)).
					Get(auditRoute, s.webhooksService.AuditEntryHandler)
			})
		})

		// Items
		itemPath := "items"
		itemsRouteWithPrefix := fmt.Sprintf("/%s", itemPath)
		itemIDRouteParam := buildNumericIDURLChunk(items.ItemIDURIParamKey)
		v1Router.Route(itemsRouteWithPrefix, func(itemsRouter routing.Router) {
			itemsRouter.
				WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.CreateItemsPermission)).
				Post(root, s.itemsService.CreateHandler)
			itemsRouter.
				WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadItemsPermission)).
				Get(root, s.itemsService.ListHandler)
			itemsRouter.
				WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadItemsPermission)).
				Get(searchRoot, s.itemsService.SearchHandler)

			itemsRouter.Route(itemIDRouteParam, func(singleItemRouter routing.Router) {
				singleItemRouter.
					WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadItemsPermission)).
					Get(root, s.itemsService.ReadHandler)
				singleItemRouter.
					WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadItemsPermission)).
					Head(root, s.itemsService.ExistenceHandler)
				singleItemRouter.
					WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ArchiveItemsPermission)).
					Delete(root, s.itemsService.ArchiveHandler)
				singleItemRouter.
					WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.UpdateItemsPermission)).
					Put(root, s.itemsService.UpdateHandler)
				singleItemRouter.
					WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadItemsAuditLogEntriesPermission)).
					Get(auditRoute, s.itemsService.AuditEntryHandler)
			})
		})
	})

	s.router = router
}
