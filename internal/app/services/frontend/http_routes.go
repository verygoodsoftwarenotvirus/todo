package frontend

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing"
)

const (
	numericIDPattern = "{%s:[0-9]+}"
)

// SetupRoutes sets up the routes.
func (s *Service) SetupRoutes(router routing.Router) {
	router = router.WithMiddleware(s.authService.UserAttributionMiddleware)

	router.Get("/", s.homepage)
	router.Get("/dashboard", s.homepage)
	router.Get("/favicon.svg", s.favicon)

	// auth stuff
	router.Get("/login", s.buildLoginView(true))
	router.Get("/components/login_prompt", s.buildLoginView(false))
	router.Post("/auth/submit_login", s.handleLoginSubmission)

	router.Get("/register", s.registrationView)
	router.Get("/components/registration_prompt", s.registrationComponent)
	router.Post("/auth/submit_registration", s.handleRegistrationSubmission)
	router.Post("/auth/verify_two_factor_secret", s.handleTOTPVerificationSubmission)

	singleAccountPattern := fmt.Sprintf(numericIDPattern, accountIDURLParamKey)
	router.Get("/accounts", s.buildAccountsView(true))
	router.Get(fmt.Sprintf("/accounts/%s", singleAccountPattern), s.buildAccountView(true))
	router.Get("/dashboard_pages/accounts", s.buildAccountsView(false))
	router.Get(fmt.Sprintf("/dashboard_pages/accounts/%s", singleAccountPattern), s.buildAccountView(false))

	singleAPIClientPattern := fmt.Sprintf(numericIDPattern, apiClientIDURLParamKey)
	router.Get("/api_clients", s.buildAPIClientsTableView(true))
	router.Get(fmt.Sprintf("/api_clients/%s", singleAPIClientPattern), s.buildAPIClientEditorView(true))
	router.Get("/dashboard_pages/api_clients", s.buildAPIClientsTableView(false))
	router.Get(fmt.Sprintf("/dashboard_pages/api_clients/%s", singleAPIClientPattern), s.buildAPIClientEditorView(false))

	singleWebhookPattern := fmt.Sprintf(numericIDPattern, webhookIDURLParamKey)
	router.Get("/account/webhooks", s.buildWebhooksTableView(true))
	router.Get(fmt.Sprintf("/account/webhooks/%s", singleWebhookPattern), s.buildWebhookEditorView(true))
	router.Get("/dashboard_pages/account/webhooks", s.buildWebhooksTableView(false))
	router.Get(fmt.Sprintf("/dashboard_pages/account/webhooks/%s", singleWebhookPattern), s.buildWebhookEditorView(false))

	router.Get("/user/settings", s.buildUserSettingsView(true))
	router.Get("/dashboard_pages/user/settings", s.buildUserSettingsView(false))
	router.Get("/account/settings", s.buildAccountSettingsView(true))
	router.Get("/dashboard_pages/account/settings", s.buildAccountSettingsView(false))

	singleItemPattern := fmt.Sprintf(numericIDPattern, itemIDURLParamKey)
	router.Get("/items", s.buildItemsTableView(true))
	router.Get(fmt.Sprintf("/items/%s", singleItemPattern), s.buildItemEditorView(true))
	router.Get("/dashboard_pages/items", s.buildItemsTableView(false))
	router.Get(fmt.Sprintf("/dashboard_pages/items/%s", singleItemPattern), s.buildItemEditorView(false))
}
