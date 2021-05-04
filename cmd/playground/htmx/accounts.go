package main

import (
	"bytes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"html/template"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

func buildAccountViewer(x *types.Account) string {
	var b bytes.Buffer
	if err := template.Must(template.New("").Funcs(defaultFuncMap).Parse(buildBasicEditorTemplate(&basicEditorTemplateConfig{
		Name: "Account",
		ID:   12345,
		Fields: []genericEditorField{
			{
				Name:      "Name",
				InputType: "text",
				Required:  true,
			},
		},
	}))).Execute(&b, x); err != nil {
		panic(err)
	}
	return b.String()
}

func buildAccountsTable() string {
	accounts := fakes.BuildFakeAccountList()
	return renderTemplateToString(template.Must(template.New("").Funcs(defaultFuncMap).Parse(buildBasicTableTemplate(&basicTableTemplateConfig{
		ExternalURL: "/accounts/123",
		GetURL:      "/dashboard_pages/accounts/123",
		Columns:     fetchTableColumns("columns.accounts"),
		CellFields: []string{
			"Name",
			"ExternalID",
			"BelongsToUser",
		},
		RowDataFieldName:     "Accounts",
		IncludeLastUpdatedOn: true,
		IncludeCreatedOn:     true,
	}))), accounts)
}

func accountsDashboardPage(res http.ResponseWriter, req *http.Request) {
	renderStringToResponse(buildDashboardSubpageString("Accounts", template.HTML(buildAccountsTable())))(res, req)
}
