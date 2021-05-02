package main

import (
	"bytes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"html/template"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

var accountEditorTemplateSrc = buildGenericEditorTemplate(&genericEditorTemplateConfig{
	Name: "Account",
	ID:   12345,
	Fields: []genericEditorField{
		{
			Name:      "Name",
			InputType: "text",
			Required:  true,
		},
	},
})

var accountEditorTemplate = template.Must(template.New("").Funcs(defaultFuncMap).Parse(accountEditorTemplateSrc))

func buildAccountViewer(x *types.Account) string {
	var b bytes.Buffer
	if err := accountEditorTemplate.Execute(&b, x); err != nil {
		panic(err)
	}
	return b.String()
}

var accountsTableTemplateSrc = buildGenericTableTemplate(&genericTableTemplateConfig{
	ExternalURL: "/accounts/123",
	GetURL:      "/dashboard_pages/accounts/123",
	Columns: []string{
		"ID",
		"Name",
		"External ID",
		"Belongs To User",
		"Last Updated On",
		"Created On",
	},
	CellFields: []string{
		"Name",
		"ExternalID",
		"BelongsToUser",
	},
	RowDataFieldName:     "Accounts",
	IncludeLastUpdatedOn: true,
	IncludeCreatedOn:     true,
})

var accountsTableTemplate = template.Must(template.New("").Funcs(defaultFuncMap).Parse(accountsTableTemplateSrc))

func buildAccountsTable() string {
	accounts := fakes.BuildFakeAccountList()
	return renderTemplateToString(accountsTableTemplate, accounts)
}

func accountsDashboardPage(res http.ResponseWriter, req *http.Request) {
	renderStringToResponse(buildDashboardSubpageString("Accounts", template.HTML(buildAccountsTable())))(res, req)
}
