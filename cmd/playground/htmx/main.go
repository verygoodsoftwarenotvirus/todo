package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed translations/*.toml
var translationsDir embed.FS

func prepareColumns(s string) []string {
	out := []string{}

	for _, x := range strings.Split(s, ",") {
		out = append(out, strings.TrimSpace(x))
	}

	return out
}

var defaultFuncMap = map[string]interface{}{
	"relativeTime":        relativeTime,
	"relativeTimeFromPtr": relativeTimeFromPtr,
}

var localizer *i18n.Localizer

func initializeLocalizer() *i18n.Localizer {
	if localizer == nil {
		bundle := i18n.NewBundle(language.English)
		bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

		translationFolderContents, err := fs.ReadDir(translationsDir, "translations")
		if err != nil {
			panic(err)
		}

		for _, f := range translationFolderContents {
			translationFileBytes, err := fs.ReadFile(translationsDir, filepath.Join("translations", f.Name()))
			if err != nil {
				panic(err)
			}

			bundle.MustParseMessageFileBytes(translationFileBytes, f.Name())
		}

		localizer = i18n.NewLocalizer(bundle, "en")
	}

	return localizer
}

func init() {
	initializeLocalizer()
}

func main() {
	initializeLocalizer()

	mux := http.NewServeMux()

	// dashboard pages
	mux.HandleFunc("/", renderRawStringIntoDashboard(""))
	mux.HandleFunc("/dashboard", renderRawStringIntoDashboard(""))
	mux.HandleFunc("/login", renderRawStringIntoDashboard(loginPrompt))

	// components
	mux.HandleFunc("/components/login_prompt", loginComponent)

	mux.HandleFunc("/items", renderRawStringIntoDashboard(buildDashboardSubpageString("Items", template.HTML(buildItemsTable()))))
	mux.HandleFunc("/items/123", renderRawStringIntoDashboard(buildItemViewer(fakes.BuildFakeItem())))
	mux.HandleFunc("/dashboard_pages/items", itemsDashboardPage)
	mux.HandleFunc("/dashboard_pages/items/123", renderStringToResponse(buildItemViewer(fakes.BuildFakeItem())))

	mux.HandleFunc("/api_clients", renderRawStringIntoDashboard(buildDashboardSubpageString("APIClients", template.HTML(buildAPIClientsTable()))))
	mux.HandleFunc("/api_clients/123", renderRawStringIntoDashboard(buildAPIClientViewer(fakes.BuildFakeAPIClient())))
	mux.HandleFunc("/dashboard_pages/api_clients", apiClientsDashboardPage)
	mux.HandleFunc("/dashboard_pages/api_clients/123", renderStringToResponse(buildAPIClientViewer(fakes.BuildFakeAPIClient())))

	mux.HandleFunc("/accounts", renderRawStringIntoDashboard(buildDashboardSubpageString("Accounts", template.HTML(buildAccountsTable()))))
	mux.HandleFunc("/accounts/123", renderRawStringIntoDashboard(buildAccountViewer(fakes.BuildFakeAccount())))
	mux.HandleFunc("/dashboard_pages/accounts", accountsDashboardPage)
	mux.HandleFunc("/dashboard_pages/accounts/123", renderStringToResponse(buildAccountViewer(fakes.BuildFakeAccount())))

	mux.HandleFunc("/account/webhooks", renderRawStringIntoDashboard(buildDashboardSubpageString("Webhooks", template.HTML(buildWebhooksTable()))))
	mux.HandleFunc("/account/webhooks/123", renderRawStringIntoDashboard(buildWebhookViewer(fakes.BuildFakeWebhook())))
	mux.HandleFunc("/dashboard_pages/account/webhooks", webhooksDashboardPage)
	mux.HandleFunc("/dashboard_pages/account/webhooks/123", renderStringToResponse(buildWebhookViewer(fakes.BuildFakeWebhook())))

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalln(err)
	}
}
