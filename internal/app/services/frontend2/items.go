package frontend2

import (
	"html/template"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

func buildItemViewer(x *types.Item) template.HTML {
	tmplConfig := &basicEditorTemplateConfig{
		Name: "Item",
		ID:   x.ID,
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
	}
	tmpl := template.Must(template.New("").Funcs(defaultFuncMap).Parse(buildBasicEditorTemplate(tmplConfig)))

	return renderTemplateToHTML(tmpl, x)
}

func buildItemsTableDashboardPage(items *types.ItemList) template.HTML {
	itemsTableConfig := &basicTableTemplateConfig{
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
	}
	tmpl := template.Must(template.New("").Funcs(defaultFuncMap).Parse(buildBasicTableTemplate(itemsTableConfig)))

	return buildDashboardSubpageString(
		"Items",
		renderTemplateToHTML(tmpl, items),
	)
}

// ItemsPage is the dashboard page with items table included.
func (s *Service) ItemsPage(res http.ResponseWriter, req *http.Request) {
	var items *types.ItemList
	if useFakes(req) {
		items = fakes.BuildFakeItemList()
	}

	renderRawStringIntoDashboard(buildItemsTableDashboardPage(items))(res, req)
}

func itemsDashboardPage(res http.ResponseWriter, req *http.Request) {
	var items *types.ItemList
	if useFakes(req) {
		items = fakes.BuildFakeItemList()
	}

	renderHTMLTemplateToResponse(buildDashboardSubpageString("Items", buildItemsTableDashboardPage(items)))(res, req)
}
