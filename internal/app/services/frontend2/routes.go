package frontend2

import (
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
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
	getLocalizer()

	router.Get("/", renderRawStringIntoDashboard(""))
	router.Get("/favicon.ico", s.favicon)

	router.Get("/dashboard", renderRawStringIntoDashboard(""))
	router.Get("/login", renderRawStringIntoDashboard(loginPrompt))

	// components
	router.Get("/components/login_prompt", loginComponent)

	router.Get("/items", s.ItemsPage)
	router.Get("/items/123", renderRawStringIntoDashboard(buildItemViewer(fakes.BuildFakeItem())))
	router.Get("/dashboard_pages/items", itemsDashboardPage)
	router.Get("/dashboard_pages/items/123", renderHTMLTemplateToResponse(buildItemViewer(fakes.BuildFakeItem())))

	router.Get("/api_clients", renderRawStringIntoDashboard(buildDashboardSubpageString("API Clients", buildAPIClientsTable())))
	router.Get("/api_clients/123", renderRawStringIntoDashboard(buildAPIClientViewer(fakes.BuildFakeAPIClient())))
	router.Get("/dashboard_pages/api_clients", apiClientsDashboardPage)
	router.Get("/dashboard_pages/api_clients/123", renderHTMLTemplateToResponse(buildAPIClientViewer(fakes.BuildFakeAPIClient())))

	router.Get("/accounts", renderRawStringIntoDashboard(buildDashboardSubpageString("Accounts", buildAccountsTable())))
	router.Get("/accounts/123", renderRawStringIntoDashboard(buildAccountViewer(fakes.BuildFakeAccount())))
	router.Get("/dashboard_pages/accounts", accountsDashboardPage)
	router.Get("/dashboard_pages/accounts/123", renderHTMLTemplateToResponse(buildAccountViewer(fakes.BuildFakeAccount())))

	router.Get("/account/webhooks", renderRawStringIntoDashboard(buildDashboardSubpageString("Webhooks", buildWebhooksTable())))
	router.Get("/account/webhooks/123", renderRawStringIntoDashboard(buildWebhookViewer(fakes.BuildFakeWebhook())))
	router.Get("/dashboard_pages/account/webhooks", webhooksDashboardPage)
	router.Get("/dashboard_pages/account/webhooks/123", renderHTMLTemplateToResponse(buildWebhookViewer(fakes.BuildFakeWebhook())))

	router.Get("/dashboard_pages/user/settings", userSettingsDashboardPage)
	router.Get("/user/settings", renderRawStringIntoDashboard(buildUserSettingsDashboardPage()))
	router.Get("/dashboard_pages/account/settings", accountSettingsDashboardPage)
	router.Get("/account/settings", renderRawStringIntoDashboard(buildAccountSettingsDashboardPage()))
}
