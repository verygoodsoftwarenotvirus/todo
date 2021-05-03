package main

import (
	"bytes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"html/template"
	"net/http"
)

func init() {
	getLocalizer()
}

var itemEditorTemplateSrc = buildBasicEditorTemplate(&basicEditorTemplateConfig{
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
})

var itemEditorTemplate = template.Must(template.New("").Funcs(defaultFuncMap).Parse(itemEditorTemplateSrc))

func buildItemViewer(x *types.Item) string {
	var b bytes.Buffer
	if err := itemEditorTemplate.Execute(&b, x); err != nil {
		panic(err)
	}
	return b.String()
}

var itemsTableTemplateSrc = buildBasicTableTemplate(&basicTableTemplateConfig{
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
})

var itemsTableTemplate = template.Must(template.New("").Funcs(defaultFuncMap).Parse(itemsTableTemplateSrc))

func buildItemsTable() string {
	items := fakes.BuildFakeItemList()
	return renderTemplateToString(itemsTableTemplate, items)
}

func itemsDashboardPage(res http.ResponseWriter, req *http.Request) {
	renderStringToResponse(buildDashboardSubpageString("Items", template.HTML(buildItemsTable())))(res, req)
}
