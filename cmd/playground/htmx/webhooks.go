package main

import (
	"bytes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"html/template"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

var webhookEditorTemplateSrc = buildBasicEditorTemplate(&basicEditorTemplateConfig{
	Name: "Webhook",
	ID:   12345,
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
})

var webhookEditorTemplate = template.Must(template.New("").Funcs(defaultFuncMap).Parse(webhookEditorTemplateSrc))

func buildWebhookViewer(x *types.Webhook) string {
	var b bytes.Buffer
	if err := webhookEditorTemplate.Execute(&b, x); err != nil {
		panic(err)
	}
	return b.String()
}

var webhooksTableTemplateSrc = buildBasicTableTemplate(&basicTableTemplateConfig{
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
})

var webhooksTableTemplate = template.Must(template.New("").Funcs(defaultFuncMap).Parse(webhooksTableTemplateSrc))

func buildWebhooksTable() string {
	webhooks := fakes.BuildFakeWebhookList()
	return renderTemplateToString(webhooksTableTemplate, webhooks)
}

func webhooksDashboardPage(res http.ResponseWriter, req *http.Request) {
	renderStringToResponse(buildDashboardSubpageString("Webhooks", template.HTML(buildWebhooksTable())))(res, req)
}
