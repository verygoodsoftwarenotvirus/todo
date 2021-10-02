package frontend

import (
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing"
)

const (
	numericIDPattern                 = "{%s:[0-9]+}"
	unauthorizedRedirectResponseCode = http.StatusSeeOther
)

// SetupRoutes sets up the routes.
func (s *service) SetupRoutes(router routing.Router) {
	router = router.WithMiddleware(s.authService.UserAttributionMiddleware)

	router.Get("/", s.homepage)
	router.Get("/dashboard", s.homepage)

	// statics
	router.Get("/favicon.svg", s.favicon)

	// auth stuff
	router.Get("/login", s.buildLoginView(true))
	router.Get("/components/login_prompt", s.buildLoginView(false))
	router.Post("/auth/submit_login", s.handleLoginSubmission)
	router.Post("/logout", s.handleLogoutSubmission)

	router.Get("/register", s.registrationView)
	router.Get("/components/registration_prompt", s.registrationComponent)
	router.Post("/auth/submit_registration", s.handleRegistrationSubmission)
	router.Post("/auth/verify_two_factor_secret", s.handleTOTPVerificationSubmission)

	singleAccountPattern := fmt.Sprintf(numericIDPattern, accountIDURLParamKey)
	router.Get("/accounts", s.buildAccountsTableView(true))
	router.Get(fmt.Sprintf("/accounts/%s", singleAccountPattern), s.buildAccountEditorView(true))
	router.Get("/dashboard_pages/accounts", s.buildAccountsTableView(false))
	router.Get(fmt.Sprintf("/dashboard_pages/accounts/%s", singleAccountPattern), s.buildAccountEditorView(false))

	singleAPIClientPattern := fmt.Sprintf(numericIDPattern, apiClientIDURLParamKey)
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadAPIClientsPermission)).
		Get("/api_clients", s.buildAPIClientsTableView(true))
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadAPIClientsPermission)).
		Get("/dashboard_pages/api_clients", s.buildAPIClientsTableView(false))
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadAPIClientsPermission)).
		Get(fmt.Sprintf("/api_clients/%s", singleAPIClientPattern), s.buildAPIClientEditorView(true))
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadAPIClientsPermission)).
		Get(fmt.Sprintf("/dashboard_pages/api_clients/%s", singleAPIClientPattern), s.buildAPIClientEditorView(false))

	singleWebhookPattern := fmt.Sprintf(numericIDPattern, webhookIDURLParamKey)
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadWebhooksPermission)).
		Get("/account/webhooks", s.buildWebhooksTableView(true))
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadWebhooksPermission)).
		Get("/dashboard_pages/account/webhooks", s.buildWebhooksTableView(false))
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.UpdateWebhooksPermission)).
		Get(fmt.Sprintf("/account/webhooks/%s", singleWebhookPattern), s.buildWebhookEditorView(true))
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.UpdateWebhooksPermission)).
		Get(fmt.Sprintf("/dashboard_pages/account/webhooks/%s", singleWebhookPattern), s.buildWebhookEditorView(false))

	router.Get("/user/settings", s.buildUserSettingsView(true))
	router.Get("/dashboard_pages/user/settings", s.buildUserSettingsView(false))
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.UpdateAccountPermission)).
		Get("/account/settings", s.buildAccountSettingsView(true))
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.UpdateAccountPermission)).
		Get("/dashboard_pages/account/settings", s.buildAccountSettingsView(false))
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.SearchUserPermission)).
		Get("/admin/users/search", s.buildUsersTableView(true, true))
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.SearchUserPermission)).
		Get("/dashboard_pages/admin/users/search", s.buildUsersTableView(false, true))
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadUserPermission)).
		Get("/admin/users", s.buildUsersTableView(true, false))
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadUserPermission)).
		Get("/dashboard_pages/admin/users", s.buildUsersTableView(false, false))
	router.WithMiddleware(s.authService.ServiceAdminMiddleware).
		Get("/admin/settings", s.buildAdminSettingsView(true))
	router.WithMiddleware(s.authService.ServiceAdminMiddleware).
		Get("/dashboard_pages/admin/settings", s.buildAdminSettingsView(false))

	singleItemPattern := fmt.Sprintf(numericIDPattern, itemIDURLParamKey)
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadItemsPermission)).
		Get("/items", s.buildItemsTableView(true))
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ReadItemsPermission)).
		Get("/dashboard_pages/items", s.buildItemsTableView(false))
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.CreateItemsPermission)).
		Get("/items/new", s.buildItemCreatorView(true))
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.CreateItemsPermission)).
		Post("/items/new/submit", s.handleItemCreationRequest)
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ArchiveItemsPermission)).
		Delete(fmt.Sprintf("/dashboard_pages/items/%s", singleItemPattern), s.handleItemArchiveRequest)
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.ArchiveItemsPermission)).
		Get("/dashboard_pages/items/new", s.buildItemCreatorView(false))
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.UpdateItemsPermission)).
		Get(fmt.Sprintf("/items/%s", singleItemPattern), s.buildItemEditorView(true))
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.UpdateItemsPermission)).
		Put(fmt.Sprintf("/dashboard_pages/items/%s", singleItemPattern), s.handleItemUpdateRequest)
	router.WithMiddleware(s.authService.PermissionFilterMiddleware(authorization.UpdateItemsPermission)).
		Get(fmt.Sprintf("/dashboard_pages/items/%s", singleItemPattern), s.buildItemEditorView(false))
}
