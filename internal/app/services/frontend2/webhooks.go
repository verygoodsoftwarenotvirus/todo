package frontend2

import (
	"html/template"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

func buildWebhookViewer(x *types.Webhook) template.HTML {
	tmplConfig := &basicEditorTemplateConfig{
		Name: "Webhook",
		ID:   x.ID,
		Fields: []genericEditorField{
			{
				Name:      "Name",
				InputType: "text",
				Required:  true,
			},
			{
				Name:      "Method",
				InputType: "text",
				Required:  true,
			},
			{
				Name:      "ContentType",
				InputType: "text",
				Required:  true,
			},
			{
				Name:      "URL",
				InputType: "text",
				Required:  true,
			},
		},
	}
	tmpl := template.Must(template.New("").Funcs(defaultFuncMap).Parse(buildBasicEditorTemplate(tmplConfig)))

	return renderTemplateToHTML(tmpl, x)
}

func buildWebhooksTable() template.HTML {
	webhooks := fakes.BuildFakeWebhookList()

	tableConfig := &basicTableTemplateConfig{
		ExternalURL: "/account/webhooks/123",
		GetURL:      "/dashboard_pages/account/webhooks/123",
		Columns:     fetchTableColumns("columns.webhooks"),
		CellFields: []string{
			"Name",
			"Method",
			"URL",
			"ContentType",
			"BelongsToAccount",
		},
		RowDataFieldName:     "Webhooks",
		IncludeLastUpdatedOn: true,
		IncludeCreatedOn:     true,
	}
	tmpl := template.Must(template.New("").Funcs(defaultFuncMap).Parse(buildBasicTableTemplate(tableConfig)))

	return renderTemplateToHTML(tmpl, webhooks)
}

func webhooksDashboardPage(res http.ResponseWriter, req *http.Request) {
	renderHTMLTemplateToResponse(buildDashboardSubpageString("Webhooks", buildWebhooksTable()))(res, req)
}
