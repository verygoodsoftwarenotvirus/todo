package main

import (
	"bytes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"html/template"
	"net/http"
)

func buildItemViewer(x *types.Item) string {
	var b bytes.Buffer
	if err := template.Must(template.New("").Funcs(defaultFuncMap).Parse(buildBasicEditorTemplate(&basicEditorTemplateConfig{
		Name: "Item",
		ID:   12345,
		Fields: []genericEditorField{
			{
				Name:      "Name",
				InputType: "text",
				Required:  true,
			},
			{
				Name:      "Details",
				InputType: "text",
				Required:  false,
			},
		},
	}))).Execute(&b, x); err != nil {
		panic(err)
	}
	return b.String()
}

func buildItemsTable() string {
	items := fakes.BuildFakeItemList()
	return renderTemplateToString(template.Must(template.New("").Funcs(defaultFuncMap).Parse(buildBasicTableTemplate(&basicTableTemplateConfig{
		ExternalURL: "/items/123",
		GetURL:      "/dashboard_pages/items/123",
		Columns:     fetchTableColumns("columns.items"),
		CellFields: []string{
			"Name",
			"Details",
		},
		RowDataFieldName:     "Items",
		IncludeLastUpdatedOn: true,
		IncludeCreatedOn:     true,
	}))), items)
}

func itemsDashboardPage(res http.ResponseWriter, req *http.Request) {
	renderStringToResponse(buildDashboardSubpageString("Items", template.HTML(buildItemsTable())))(res, req)
}
