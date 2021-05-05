package frontend

import (
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

func buildViewerForItem(x *types.Item) (string, error) {
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
	tmpl := parseTemplate("", buildBasicEditorTemplate(tmplConfig))

	return renderTemplateToString(tmpl, x)
}

func (s *Service) itemDashboardPage(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var item *types.Item
	if s.useFakes {
		logger.Debug("using fakes")
		item = fakes.BuildFakeItem()
	}

	page, err := buildViewerForItem(item)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering item table template into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderStringToResponse(page, res)
}

func buildItemsTableDashboardPage(items *types.ItemList) (string, error) {
	tmpl := parseTemplate("dashboard", buildBasicTableTemplate(itemsTableConfig))

	return renderTemplateToString(tmpl, items)
}

func (s *Service) itemsDashboardPage(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var items *types.ItemList
	if s.useFakes {
		logger.Debug("using fakes")
		items = fakes.BuildFakeItemList()
	}

	page, err := buildItemsTableDashboardPage(items)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering item table template into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderStringToResponse(page, res)
}

func buildItemDashboardView(x *types.Item) (string, error) {
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

	return renderTemplateIntoDashboard("Items", wrapTemplateInContentDefinition(buildBasicEditorTemplate(tmplConfig)), x)
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

	dashboard, err := buildItemDashboardView(item)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering item viewer into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderStringToResponse(dashboard, res)
}

func buildItemsTableDashboardView(items *types.ItemList) (string, error) {
	tmpl := wrapTemplateInContentDefinition(buildBasicTableTemplate(itemsTableConfig))

	return renderTemplateIntoDashboard("Items", tmpl, items)
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

	dashboard, err := buildItemsTableDashboardView(items)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering item table template into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderStringToResponse(dashboard, res)
}
