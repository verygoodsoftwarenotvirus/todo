package elements

import (
	"html/template"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

func buildAPIClientViewer(x *types.APIClient) template.HTML {
	editorConfig := &basicEditorTemplateConfig{
		Name: "API Client",
		ID:   x.ID,
		Fields: []genericEditorField{
			{
				Name:      "Name",
				InputType: "text",
				Required:  true,
			},
			{
				Name:      "ExternalID",
				InputType: "text",
				Required:  true,
			},
			{
				Name:      "ClientID",
				InputType: "text",
				Required:  true,
			},
		},
	}
	tmpl := template.Must(template.New("").Funcs(defaultFuncMap).Parse(buildBasicEditorTemplate(editorConfig)))

	return renderTemplateToHTML(tmpl, x)
}

func buildAPIClientsTable() template.HTML {
	apiClients := fakes.BuildFakeAPIClientList()
	tableConfig := &basicTableTemplateConfig{
		ExternalURL: "/api_clients/123",
		GetURL:      "/dashboard_pages/api_clients/123",
		Columns:     fetchTableColumns("columns.apiClients"),
		CellFields: []string{
			"Name",
			"ExternalID",
			"ClientID",
			"BelongsToUser",
		},
		RowDataFieldName:     "Clients",
		IncludeLastUpdatedOn: false,
		IncludeCreatedOn:     true,
	}
	tmpl := template.Must(template.New("").Funcs(defaultFuncMap).Parse(buildBasicTableTemplate(tableConfig)))

	return renderTemplateToHTML(tmpl, apiClients)
}

func apiClientsDashboardPage(res http.ResponseWriter, req *http.Request) {
	renderHTMLTemplateToResponse(buildDashboardSubpageString("API Clients", buildAPIClientsTable()))(res, req)
}
