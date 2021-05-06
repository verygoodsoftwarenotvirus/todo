package frontend

import (
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing"
)

func fetchTableColumns(messageID string) []string {
	out := []string{}

	for _, x := range strings.Split(getSimpleLocalizedString(messageID), ",") {
		out = append(out, strings.TrimSpace(x))
	}

	return out
}

var defaultFuncMap = map[string]interface{}{
	"relativeTime":        relativeTime,
	"relativeTimeFromPtr": relativeTimeFromPtr,
	"translate":           getSimpleLocalizedString,
}

// SetupRoutes sets up the routes.
func (s *Service) SetupRoutes(router routing.Router) {
	initLocalizer()

	router.Get("/", s.homepage)
	router.Get("/dashboard", s.homepage)
	router.Get("/favicon.svg", s.favicon)

	// auth stuff
	router.Get("/login", s.loginView)
	router.Get("/components/login_prompt", s.loginComponent)
	router.WithMiddleware(s.authService.UserLoginInputMiddleware).Post("/auth/submit_login", s.handleLoginSubmission)

	router.Get("/register", s.registrationView)
	router.Get("/components/registration_prompt", s.registrationComponent)
	router.WithMiddleware(s.usersService.UserRegistrationInputMiddleware).Post("/auth/submit_registration", s.handleRegistrationSubmission)

	attributedRouter := router.WithMiddleware(s.authService.UserAttributionMiddleware)

	attributedRouter.Get("/accounts", s.accountsDashboardView)
	attributedRouter.Get("/accounts/123", s.accountDashboardView)
	attributedRouter.Get("/dashboard_pages/accounts", s.accountsTableView)
	attributedRouter.Get("/dashboard_pages/accounts/123", s.accountsEditorView)

	attributedRouter.Get("/api_clients", s.apiClientsDashboardView)
	attributedRouter.Get("/api_clients/123", s.apiClientDashboardView)
	attributedRouter.Get("/dashboard_pages/api_clients", s.apiClientsTableView)
	attributedRouter.Get("/dashboard_pages/api_clients/123", s.apiClientsEditorView)

	attributedRouter.Get("/account/webhooks", s.webhooksDashboardView)
	attributedRouter.Get("/account/webhooks/123", s.webhookDashboardView)
	attributedRouter.Get("/dashboard_pages/account/webhooks", s.webhooksTableView)
	attributedRouter.Get("/dashboard_pages/account/webhooks/123", s.webhooksEditorView)

	attributedRouter.Get("/dashboard_pages/user/settings", s.userSettingsView)
	attributedRouter.Get("/user/settings", s.userSettingsDashboardView)
	attributedRouter.Get("/dashboard_pages/account/settings", s.accountSettingsView)
	attributedRouter.Get("/account/settings", s.accountSettingsDashboardView)

	attributedRouter.Get("/items", s.itemsDashboardView)
	attributedRouter.Get("/items/123", s.itemDashboardView)
	attributedRouter.Get("/dashboard_pages/items", s.itemsTableView)
	attributedRouter.Get("/dashboard_pages/items/123", s.itemsEditorView)
}
