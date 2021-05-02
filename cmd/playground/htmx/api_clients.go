package main

import (
	"bytes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"html/template"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

var apiClientEditorTemplateSrc = buildGenericEditorTemplate(&genericEditorTemplateConfig{
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
})

var apiClientEditorTemplate = template.Must(template.New("").Funcs(defaultFuncMap).Parse(apiClientEditorTemplateSrc))

func buildAPIClientViewer(x *types.APIClient) string {
	var b bytes.Buffer
	if err := apiClientEditorTemplate.Execute(&b, x); err != nil {
		panic(err)
	}
	return b.String()
}

var apiClientsTableTemplateSrc = buildGenericTableTemplate(&genericTableTemplateConfig{
	ExternalURL: "/api_clients/123",
	GetURL:      "/dashboard_pages/api_clients/123",
	Columns: []string{
		"ID",
		"Name",
		"External ID",
		"Client ID",
		"Belongs To User",
		"Created On",
	},
	CellFields: []string{
		"Name",
		"ExternalID",
		"ClientID",
		"BelongsToUser",
	},
	RowDataFieldName:     "Clients",
	IncludeLastUpdatedOn: false,
	IncludeCreatedOn:     true,
})

var apiClientsTableTemplate = template.Must(template.New("").Funcs(defaultFuncMap).Parse(apiClientsTableTemplateSrc))

func buildAPIClientsTable() string {
	apiClients := fakes.BuildFakeAPIClientList()
	return renderTemplateToString(apiClientsTableTemplate, apiClients)
}

func apiClientsDashboardPage(res http.ResponseWriter, req *http.Request) {
	renderStringToResponse(buildDashboardSubpageString("APIClients", template.HTML(buildAPIClientsTable())))(res, req)
}
