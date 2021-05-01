package main

import (
	"html/template"
	"log"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

func main() {
	mux := http.NewServeMux()

	// dashboard pages
	mux.HandleFunc("/", renderRawStringIntoDashboard(""))
	mux.HandleFunc("/dashboard", renderRawStringIntoDashboard(""))
	mux.HandleFunc("/login", renderRawStringIntoDashboard(loginPrompt))
	mux.HandleFunc("/accounts", renderRawStringIntoDashboard(buildDashboardSubpageString("Accounts", template.HTML(exampleAccountsTable()))))
	mux.HandleFunc("/webhooks", renderRawStringIntoDashboard(buildDashboardSubpageString("Webhooks", template.HTML(exampleWebhooksTable()))))
	mux.HandleFunc("/items", renderRawStringIntoDashboard(buildDashboardSubpageString("Items", template.HTML(buildItemsTable()))))
	mux.HandleFunc("/items/123", renderRawStringIntoDashboard(buildItemViewer(fakes.BuildFakeItem())))

	// individual component pages
	mux.HandleFunc("/components/login_prompt", loginComponent)
	mux.HandleFunc("/dashboard_pages/items", itemsDashboardPage)
	mux.HandleFunc("/dashboard_pages/items/123", renderStringToResponse(buildItemViewer(fakes.BuildFakeItem())))
	mux.HandleFunc("/dashboard_pages/api_clients", apiClientsDashboardPage)
	mux.HandleFunc("/dashboard_pages/accounts", accountsDashboardPage)
	mux.HandleFunc("/dashboard_pages/account/webhooks", webhooksDashboardPage)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalln(err)
	}
}
