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
// 		router.Get("/login", renderRawStringIntoDashboard(loginPrompt))
//
// 		router.Get("/api_clients", renderRawStringIntoDashboard(buildDashboardSubpageString("API Clients", buildAPIClientsTable())))
// 		router.Get("/api_clients/123", renderRawStringIntoDashboard(buildAPIClientViewer(fakes.BuildFakeAPIClient())))
// 		router.Get("/dashboard_pages/api_clients", apiClientsDashboardPage)
// 		router.Get("/dashboard_pages/api_clients/123", renderHTMLTemplateToResponse(buildAPIClientViewer(fakes.BuildFakeAPIClient())))
//
// 		router.Get("/accounts", renderRawStringIntoDashboard(buildDashboardSubpageString("Accounts", buildAccountsTable())))
// 		router.Get("/accounts/123", renderRawStringIntoDashboard(buildAccountViewer(fakes.BuildFakeAccount())))
// 		router.Get("/dashboard_pages/accounts", accountsDashboardPage)
// 		router.Get("/dashboard_pages/accounts/123", renderHTMLTemplateToResponse(buildAccountViewer(fakes.BuildFakeAccount())))
//
// 		router.Get("/account/webhooks", renderRawStringIntoDashboard(buildDashboardSubpageString("Webhooks", buildWebhooksTable())))
// 		router.Get("/account/webhooks/123", renderRawStringIntoDashboard(buildWebhookViewer(fakes.BuildFakeWebhook())))
// 		router.Get("/dashboard_pages/account/webhooks", webhooksDashboardPage)
// 		router.Get("/dashboard_pages/account/webhooks/123", renderHTMLTemplateToResponse(buildWebhookViewer(fakes.BuildFakeWebhook())))
//
// lol.
func (s *Service) SetupRoutes(router routing.Router) {
	initLocalizer()

	router.Get("/", s.homepage)
	router.Get("/dashboard", s.homepage)
	router.Get("/favicon.svg", s.favicon)

	// components
	router.Get("/components/login_prompt", s.loginComponent)

	router.Get("/items", s.itemsDashboardView)
	router.Get("/items/123", s.itemDashboardView)
	router.Get("/dashboard_pages/items", s.itemsDashboardPage)
	router.Get("/dashboard_pages/items/123", s.itemDashboardPage)

	router.Get("/dashboard_pages/user/settings", s.userSettingsDashboardPage)
	// the above but with a full rendered form at a shorter path
	router.Get("/dashboard_pages/account/settings", s.accountSettingsDashboardPage)
	// the above but with a full rendered form at a shorter path
}
