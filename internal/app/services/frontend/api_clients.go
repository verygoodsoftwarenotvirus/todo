package frontend

import (
	"fmt"
	"log"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

var apiClientsTableConfig = &basicTableTemplateConfig{
	Title:       "API Clients",
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
}

var apiClientEditorConfig = &basicEditorTemplateConfig{
	Fields: []basicEditorField{
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
	FuncMap: map[string]interface{}{
		"componentTitle": func(x *types.APIClient) string {
			return fmt.Sprintf("Client #%d", x.ID)
		},
	},
}

func (s *Service) apiClientsEditorView(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var apiClient *types.APIClient
	if s.useFakes {
		logger.Debug("using fakes")
		apiClient = fakes.BuildFakeAPIClient()
	}

	tmpl := parseTemplate("", buildBasicEditorTemplate(apiClientEditorConfig), apiClientEditorConfig.FuncMap)

	if err := renderTemplateToResponse(tmpl, apiClient, res); err != nil {
		log.Panic(err)
	}
}

func (s *Service) apiClientsTableView(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var apiClients *types.APIClientList
	if s.useFakes {
		logger.Debug("using fakes")
		apiClients = fakes.BuildFakeAPIClientList()
	}

	tmpl := parseTemplate("dashboard", buildBasicTableTemplate(apiClientsTableConfig), nil)

	if err := renderTemplateToResponse(tmpl, apiClients, res); err != nil {
		log.Panic(err)
	}
}

func (s *Service) apiClientDashboardView(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var apiClient *types.APIClient
	if s.useFakes {
		logger.Debug("using fakes")
		apiClient = fakes.BuildFakeAPIClient()
	}

	view := renderTemplateIntoDashboard(buildBasicEditorTemplate(apiClientEditorConfig), apiClientEditorConfig.FuncMap)

	page := &dashboardPageData{
		Title:       fmt.Sprintf("APIClient #%d", apiClient.ID),
		ContentData: apiClient,
	}

	if err := renderTemplateToResponse(view, page, res); err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering API clients dashboard view")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Service) apiClientsDashboardView(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var apiClients *types.APIClientList
	if s.useFakes {
		logger.Debug("using fakes")
		apiClients = fakes.BuildFakeAPIClientList()
	}

	view := renderTemplateIntoDashboard(buildBasicTableTemplate(apiClientsTableConfig), nil)

	page := &dashboardPageData{
		Title:       "APIClients",
		ContentData: apiClients,
	}

	if err := renderTemplateToResponse(view, page, res); err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering API clients dashboard view")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}
