package frontend

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

const (
	apiClientIDURLParamKey = "api_client"
)

func (s *Service) buildAPIClientsTableConfig() *basicTableTemplateConfig {
	return &basicTableTemplateConfig{
		Title:       "API Clients",
		ExternalURL: "/api_clients/123",
		GetURL:      "/dashboard_pages/api_clients/123",
		Columns:     s.fetchTableColumns("columns.apiClients"),
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
}

func (s *Service) buildAPIClientEditorConfig() *basicEditorTemplateConfig {
	return &basicEditorTemplateConfig{
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
}

func (s *Service) fetchAPIClient(ctx context.Context, sessionCtxData *types.SessionContextData, req *http.Request) (apiClient *types.APIClient, err error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger

	apiClientID := s.routeParamManager.BuildRouteParamIDFetcher(logger, apiClientIDURLParamKey, "API client")(req)

	if s.useFakeData {
		apiClient = fakes.BuildFakeAPIClient()
	} else {
		apiClient, err = s.dataStore.GetAPIClientByDatabaseID(ctx, apiClientID, sessionCtxData.Requester.ID)
		if err != nil {
			return nil, observability.PrepareError(err, logger, span, "fetching API client data")
		}
	}

	return apiClient, nil
}

func (s *Service) buildAPIClientEditorView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		sessionCtxData, err := s.sessionContextDataFetcher(req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
			http.Redirect(res, req, "/login", http.StatusSeeOther)
			return
		}

		apiClient, err := s.fetchAPIClient(ctx, sessionCtxData, req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "error fetching item from datastore")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		apiClientEditorConfig := s.buildAPIClientEditorConfig()
		if includeBaseTemplate {
			tmpl := s.parseTemplate("", s.buildBasicEditorTemplate(apiClientEditorConfig), apiClientEditorConfig.FuncMap)

			if err = s.renderTemplateToResponse(tmpl, apiClient, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering API client editor view")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			view := s.renderTemplateIntoDashboard(s.buildBasicEditorTemplate(apiClientEditorConfig), apiClientEditorConfig.FuncMap)

			page := &dashboardPageData{
				LoggedIn:    sessionCtxData != nil,
				Title:       fmt.Sprintf("APIClient #%d", apiClient.ID),
				ContentData: apiClient,
			}

			if err = s.renderTemplateToResponse(view, page, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering API clients dashboard view")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}

func (s *Service) fetchAPIClients(ctx context.Context, sessionCtxData *types.SessionContextData, req *http.Request) (apiClients *types.APIClientList, err error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger

	if s.useFakeData {
		apiClients = fakes.BuildFakeAPIClientList()
	} else {
		filter := types.ExtractQueryFilter(req)
		apiClients, err = s.dataStore.GetAPIClients(ctx, sessionCtxData.Requester.ID, filter)
		if err != nil {
			return nil, observability.PrepareError(err, logger, span, "fetching API client data")
		}
	}

	return apiClients, nil
}

func (s *Service) buildAPIClientsTableView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		sessionCtxData, err := s.sessionContextDataFetcher(req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
			http.Redirect(res, req, "/login", http.StatusSeeOther)
			return
		}

		apiClients, err := s.fetchAPIClients(ctx, sessionCtxData, req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "error fetching API client from datastore")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		apiClientsTableConfig := s.buildAPIClientsTableConfig()
		if includeBaseTemplate {
			view := s.renderTemplateIntoDashboard(s.buildBasicTableTemplate(apiClientsTableConfig), nil)

			page := &dashboardPageData{
				LoggedIn:    sessionCtxData != nil,
				Title:       "APIClients",
				ContentData: apiClients,
			}

			if err = s.renderTemplateToResponse(view, page, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering API clients dashboard view")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			tmpl := s.parseTemplate("dashboard", s.buildBasicTableTemplate(apiClientsTableConfig), nil)

			if err = s.renderTemplateToResponse(tmpl, apiClients, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering API clients table view")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}
