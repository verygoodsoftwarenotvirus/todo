package frontend

import (
	"context"
	// import embed for the side effect.
	_ "embed"
	"fmt"
	"html/template"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

const (
	apiClientIDURLParamKey = "api_client"
)

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

//go:embed templates/partials/generated/editors/api_client_editor.gotpl
var apiClientEditorTemplate string

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
			observability.AcknowledgeError(err, logger, span, "fetching item from datastore")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		tmplFuncMap := map[string]interface{}{
			"componentTitle": func(x *types.APIClient) string {
				return fmt.Sprintf("Client #%d", x.ID)
			},
		}

		if includeBaseTemplate {
			tmpl := s.parseTemplate("", apiClientEditorTemplate, tmplFuncMap)

			if err = s.renderTemplateToResponse(tmpl, apiClient, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering API client editor view")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			view := s.renderTemplateIntoBaseTemplate(apiClientEditorTemplate, tmplFuncMap)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       fmt.Sprintf("APIClient #%d", apiClient.ID),
				ContentData: apiClient,
			}
			if sessionCtxData != nil {
				page.IsServiceAdmin = sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin()
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

//go:embed templates/partials/generated/tables/api_clients_table.gotpl
var apiClientsTableTemplate string

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
			observability.AcknowledgeError(err, logger, span, "fetching API client from datastore")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		tmplFuncMap := map[string]interface{}{
			"individualURL": func(x *types.APIClient) template.URL {
				/* #nosec G203 */
				return template.URL(fmt.Sprintf("/api_clients/%d", x.ID))
			},
			"pushURL": func(x *types.APIClient) template.URL {
				/* #nosec G203 */
				return template.URL(fmt.Sprintf("/api_clients/%d", x.ID))
			},
		}

		if includeBaseTemplate {
			view := s.renderTemplateIntoBaseTemplate(apiClientsTableTemplate, tmplFuncMap)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       "APIClients",
				ContentData: apiClients,
			}
			if sessionCtxData != nil {
				page.IsServiceAdmin = sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin()
			}

			if err = s.renderTemplateToResponse(view, page, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering API clients dashboard view")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			tmpl := s.parseTemplate("dashboard", apiClientsTableTemplate, tmplFuncMap)

			if err = s.renderTemplateToResponse(tmpl, apiClients, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering API clients table view")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}
