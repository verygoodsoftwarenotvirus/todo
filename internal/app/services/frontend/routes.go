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
// the routes used to be like this:
//		router.Get("/login", renderRawStringIntoDashboard(loginPrompt))
// lol.
func (s *Service) SetupRoutes(router routing.Router) {
	initLocalizer()

	router.Get("/", s.homepage)
	router.Get("/dashboard", s.homepage)
	router.Get("/favicon.svg", s.favicon)

	// components
	router.Get("/components/login_prompt", s.loginComponent)

	router.Get("/accounts", s.accountsDashboardView)
	router.Get("/accounts/123", s.accountDashboardView)
	router.Get("/dashboard_pages/accounts", s.accountsDashboardPage)
	router.Get("/dashboard_pages/accounts/123", s.accountDashboardPage)

	router.Get("/api_clients", s.apiClientsDashboardView)
	router.Get("/api_clients/123", s.apiClientDashboardView)
	router.Get("/dashboard_pages/api_clients", s.apiClientsDashboardPage)
	router.Get("/dashboard_pages/api_clients/123", s.apiClientDashboardPage)

	router.Get("/account/webhooks", s.webhooksDashboardView)
	router.Get("/account/webhooks/123", s.webhookDashboardView)
	router.Get("/dashboard_pages/account/webhooks", s.webhooksDashboardPage)
	router.Get("/dashboard_pages/account/webhooks/123", s.webhookDashboardPage)

	router.Get("/items", s.itemsDashboardView)
	router.Get("/items/123", s.itemDashboardView)
	router.Get("/dashboard_pages/items", s.itemsDashboardPage)
	router.Get("/dashboard_pages/items/123", s.itemDashboardPage)

	router.Get("/dashboard_pages/user/settings", s.userSettingsDashboardPage)
	router.Get("/user/settings", s.userSettingsDashboardView)
	router.Get("/dashboard_pages/account/settings", s.accountSettingsDashboardPage)
	router.Get("/account/settings", s.accountSettingsDashboardView)
}
