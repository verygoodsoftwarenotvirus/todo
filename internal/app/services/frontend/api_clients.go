package frontend

import (
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

func buildViewerForAPIClient(x *types.APIClient) (string, error) {
	tmplConfig := &basicEditorTemplateConfig{
		Name: "API Client",
		ID:   x.ID,
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
	}

	tmpl := parseTemplate("", buildBasicEditorTemplate(tmplConfig))

	return renderTemplateToString(tmpl, x)
}

func (s *Service) apiClientDashboardPage(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var apiClient *types.APIClient
	if s.useFakes {
		logger.Debug("using fakes")
		apiClient = fakes.BuildFakeAPIClient()
	}

	page, err := buildViewerForAPIClient(apiClient)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering apiClient table template into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderStringToResponse(page, res)
}

func buildAPIClientsTableDashboardPage(apiClients *types.APIClientList) (string, error) {
	tmpl := parseTemplate("dashboard", buildBasicTableTemplate(apiClientsTableConfig))

	return renderTemplateToString(tmpl, apiClients)
}

func (s *Service) apiClientsDashboardPage(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var apiClients *types.APIClientList
	if s.useFakes {
		logger.Debug("using fakes")
		apiClients = fakes.BuildFakeAPIClientList()
	}

	page, err := buildAPIClientsTableDashboardPage(apiClients)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering apiClient table template into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderStringToResponse(page, res)
}

func buildAPIClientDashboardView(x *types.APIClient) (string, error) {
	tmplConfig := &basicEditorTemplateConfig{
		Name: "API Client",
		ID:   x.ID,
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
	}

	return renderTemplateIntoDashboard("API Clients", wrapTemplateInContentDefinition(buildBasicEditorTemplate(tmplConfig)), x)
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

	dashboard, err := buildAPIClientDashboardView(apiClient)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering apiClient viewer into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderStringToResponse(dashboard, res)
}

func buildAPIClientsTableDashboardView(apiClients *types.APIClientList) (string, error) {
	tmpl := wrapTemplateInContentDefinition(buildBasicTableTemplate(apiClientsTableConfig))

	return renderTemplateIntoDashboard("APIClients", tmpl, apiClients)
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

	dashboard, err := buildAPIClientsTableDashboardView(apiClients)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering apiClient table template into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderStringToResponse(dashboard, res)
}
