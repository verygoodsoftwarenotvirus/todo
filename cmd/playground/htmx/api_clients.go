package main

import (
	"bytes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"html/template"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

func buildAPIClientViewer(x *types.APIClient) string {
	var b bytes.Buffer
	if err := template.Must(template.New("").Funcs(defaultFuncMap).Parse(buildBasicEditorTemplate(&basicEditorTemplateConfig{
		Name: "API Client",
		ID:   12345,
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
	}),
	)).Execute(&b, x); err != nil {
		panic(err)
	}
	return b.String()
}

func buildAPIClientsTable() string {
	apiClients := fakes.BuildFakeAPIClientList()
	return renderTemplateToString(template.Must(template.New("").Funcs(defaultFuncMap).Parse(buildBasicTableTemplate(&basicTableTemplateConfig{
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
	}))), apiClients)
}

func apiClientsDashboardPage(res http.ResponseWriter, req *http.Request) {
	renderStringToResponse(buildDashboardSubpageString("API Clients", template.HTML(buildAPIClientsTable())))(res, req)
}
