package main

import (
	"log"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/html"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/html/elements"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/panicking"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/nleeper/goment"
)

var gomentPanicker = panicking.NewProductionPanicker()

func mustGoment(ts uint64) *goment.Goment {
	g, err := goment.Unix(int64(ts))
	if err != nil {
		// literally impossible
		gomentPanicker.Panic(err)
	}

	return g
}

func relativeTime(ts uint64) string {
	g := mustGoment(ts)

	return g.FromNow()
}

func relativeTimeFromPtr(ts *uint64) string {
	if ts == nil {
		return "never"
	}

	return relativeTime(*ts)
}

// parseBool differs from strconv.ParseBool in that it returns false by default.
func parseBool(str string) bool {
	switch str {
	case "1", "t", "T", "true", "TRUE", "True":
		return true
	default:
		return false
	}
}

const useFakesQueryKey = "fake"

func useFakes(req *http.Request) bool {
	return parseBool(req.URL.Query().Get(useFakesQueryKey))
}

func exampleItemsTable() html.HTML {
	var items []*types.Item
	//if useFakes(req) {
	items = fakes.BuildFakeItemList().Items
	//}

	t := elements.NewTable("ID", "Name", "Details", "Belongs To Account", "Last Updated On", "Created On")

	for _, i := range items {
		err := t.AddRow(
			i.ID,
			i.Name,
			i.Details,
			i.BelongsToAccount,
			relativeTimeFromPtr(i.LastUpdatedOn),
			relativeTime(i.CreatedOn),
		)

		if err != nil {
			log.Fatal(err)
		}
	}

	return t.HTML()
}

func exampleAccountsTable() html.HTML {
	var accounts []*types.Account
	//if useFakes(req) {
	accounts = fakes.BuildFakeAccountList().Accounts
	//}

	t := elements.NewTable("ID", "Name", "External ID", "Belongs To User", "Last Updated On", "Created On")

	for _, x := range accounts {
		err := t.AddRow(
			x.ID,
			x.Name,
			x.ExternalID,
			x.BelongsToUser,
			relativeTimeFromPtr(x.LastUpdatedOn),
			relativeTime(x.CreatedOn),
		)

		if err != nil {
			log.Fatal(err)
		}
	}

	return t.HTML()
}

func exampleAPIClientsTable() html.HTML {
	var clients []*types.APIClient
	//if useFakes(req) {
	clients = fakes.BuildFakeAPIClientList().Clients
	//}

	t := elements.NewTable("ID", "Name", "Client ID", "Belongs To User", "Created On")

	for _, x := range clients {
		err := t.AddRow(
			x.ID,
			x.Name,
			x.ClientID,
			x.BelongsToUser,
			mustGoment(x.CreatedOn).FromNow(),
		)

		if err != nil {
			log.Fatal(err)
		}
	}

	return t.HTML()
}

func exampleWebhooksTable() html.HTML {
	var webhooks []*types.Webhook
	//if useFakes(req) {
	webhooks = fakes.BuildFakeWebhookList().Webhooks
	//}

	t := elements.NewTable("ID", "Name", "External ID", "URL", "Content Type", "Belongs To Account", "Last Updated On", "Created On")

	for _, x := range webhooks {
		err := t.AddRow(
			x.ID,
			x.Name,
			x.ExternalID,
			x.URL,
			x.ContentType,
			x.BelongsToAccount,
			relativeTimeFromPtr(x.LastUpdatedOn),
			relativeTime(x.CreatedOn),
		)

		if err != nil {
			log.Fatal(err)
		}
	}

	return t.HTML()
}

func buildExampleButton() html.HTML {
	return html.New(
		"button",
		html.Attr{
			"hx-get":    "/items",
			"hx-target": "#content",
		},
		"Press me!",
	)
}

func divertRequestTo(path string) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		http.Redirect(res, req, path, http.StatusFound)
	}
}

func htmlBytes(h html.HTML) []byte {
	return []byte(h())
}

func htmlString(h html.HTML) string {
	return string(htmlBytes(h))
}

func buildDashboardSubpage(title string, h html.HTML) string {
	return html.RenderMultiple(
		html.New(
			"div",
			html.Attribute(
				html.WithClasses(
					"d-flex",
					"justify-content-between",
					"flex-wrap",
					"flex-md-nowrap",
					"align-items-center",
					"pt-3",
					"pb-2",
					"mb-3",
					"border-bottom",
				),
			),
			html.New("h1", html.Attribute(html.WithClass("h2")), title),
		),
		h,
	)
}

func main() {
	mux := http.NewServeMux()

	// dashboard pages
	mux.HandleFunc("/", renderRawStringIntoDashboard(htmlString(buildExampleButton())))
	mux.HandleFunc("/dashboard", renderRawStringIntoDashboard(htmlString(buildExampleButton())))
	mux.HandleFunc("/login", renderRawStringIntoDashboard(loginPrompt))
	mux.HandleFunc("/accounts", renderRawStringIntoDashboard(buildDashboardSubpage("Accounts", exampleAccountsTable())))
	mux.HandleFunc("/webhooks", renderRawStringIntoDashboard(buildDashboardSubpage("Webhooks", exampleWebhooksTable())))
	mux.HandleFunc("/items", renderRawStringIntoDashboard(buildDashboardSubpage("Items", exampleItemsTable())))

	// individual component pages
	mux.HandleFunc("/components/login_prompt", renderStringToResponse(loginPrompt))
	mux.HandleFunc("/dashboard_pages/items", renderStringToResponse(buildDashboardSubpage("Items", exampleItemsTable())))
	mux.HandleFunc("/dashboard_pages/api_clients", renderStringToResponse(buildDashboardSubpage("API Clients", exampleAPIClientsTable())))
	mux.HandleFunc("/dashboard_pages/accounts", renderStringToResponse(buildDashboardSubpage("Accounts", exampleAccountsTable())))
	mux.HandleFunc("/dashboard_pages/account/webhooks", renderStringToResponse(buildDashboardSubpage("Webhooks", exampleWebhooksTable())))

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalln(err)
	}
}

func renderHTML(thing html.HTML) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, _ *http.Request) {
		if _, err := res.Write(htmlBytes(thing)); err != nil {
			log.Fatalln(err)
		}
	}
}

func renderStringToResponse(thing string) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, _ *http.Request) {
		if _, err := res.Write([]byte(thing)); err != nil {
			log.Fatalln(err)
		}
	}
}

const loginPrompt = `<form hx-post="/login" hx-ext="json-enc, ajax-header, event-header">
   <h1 class="h3 mb-3 fw-normal">Sign in</h1>
   <div class="form-floating"><input placeholder="username" type="text" class="form-control" id="usernameInput" name="username"><label for="usernameInput"></label></div>
   <div class="form-floating"><input type="password" class="form-control" id="passwordInput" name="password" placeholder="password"><label for="passwordInput"></label></div>
   <div class="form-floating"><input id="totpTokenInput" name="totpToken" placeholder="123456" type="numeric" class="form-control"><label for="totpTokenInput"></label></div>
   <button class="w-100 btn btn-lg btn-primary" type="submit">Sign in</button>
</form>`

func buildLoginPrompt() string {
	return loginPrompt
}
