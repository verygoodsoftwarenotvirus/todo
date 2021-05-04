package elements

import (
	"html/template"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

func buildAccountViewer(x *types.Account) template.HTML {
	editorConfig := &basicEditorTemplateConfig{
		Name: "Account",
		ID:   x.ID,
		Fields: []genericEditorField{
			{
				Name:      "Name",
				InputType: "text",
				Required:  true,
			},
		},
	}
	tmpl := template.Must(template.New("").Funcs(defaultFuncMap).Parse(buildBasicEditorTemplate(editorConfig)))

	return renderTemplateToHTML(tmpl, x)
}

func buildAccountsTable() template.HTML {
	accounts := fakes.BuildFakeAccountList()
	tableConfig := &basicTableTemplateConfig{
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
	}

	tmpl := template.Must(template.New("").Funcs(defaultFuncMap).Parse(buildBasicTableTemplate(tableConfig)))

	return renderTemplateToHTML(tmpl, accounts)
}

func accountsDashboardPage(res http.ResponseWriter, req *http.Request) {
	renderHTMLTemplateToResponse(buildDashboardSubpageString("Accounts", buildAccountsTable()))(res, req)
}
