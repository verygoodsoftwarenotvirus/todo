package frontend

import (
	"fmt"
	"log"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

var itemsTableConfig = &basicTableTemplateConfig{
	Title:       "Items",
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

var itemEditorConfig = &basicEditorTemplateConfig{
	Fields: []basicEditorField{
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
	FuncMap: map[string]interface{}{
		"componentTitle": func(x *types.Item) string {
			return fmt.Sprintf("Item #%d", x.ID)
		},
	},
}

func (s *Service) itemsEditorView(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var item *types.Item
	if s.useFakes {
		logger.Debug("using fakes")
		item = fakes.BuildFakeItem()
	}

	tmpl := parseTemplate("", buildBasicEditorTemplate(itemEditorConfig), itemEditorConfig.FuncMap)

	if err := renderTemplateToResponse(tmpl, item, res); err != nil {
		log.Panic(err)
	}
}

func (s *Service) itemsTableView(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var items *types.ItemList
	if s.useFakes {
		logger.Debug("using fakes")
		items = fakes.BuildFakeItemList()
	}

	tmpl := parseTemplate("dashboard", buildBasicTableTemplate(itemsTableConfig), nil)

	if err := renderTemplateToResponse(tmpl, items, res); err != nil {
		log.Panic(err)
	}
}

func (s *Service) itemDashboardView(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var item *types.Item
	if s.useFakes {
		logger.Debug("using fakes")
		item = fakes.BuildFakeItem()
	}

	view := renderTemplateIntoDashboard(buildBasicEditorTemplate(itemEditorConfig), itemEditorConfig.FuncMap)

	page := &dashboardPageData{
		Title:       fmt.Sprintf("Item #%d", item.ID),
		ContentData: item,
	}

	if err := renderTemplateToResponse(view, page, res); err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering items dashboard view")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Service) itemsDashboardView(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var items *types.ItemList
	if s.useFakes {
		logger.Debug("using fakes")
		items = fakes.BuildFakeItemList()
	}

	view := renderTemplateIntoDashboard(buildBasicTableTemplate(itemsTableConfig), nil)

	page := &dashboardPageData{
		Title:       "Items",
		ContentData: items,
	}

	if err := renderTemplateToResponse(view, page, res); err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering items dashboard view")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}
