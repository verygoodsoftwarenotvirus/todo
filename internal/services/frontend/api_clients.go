package frontend

import (
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
)

const (
	apiClientIDURLParamKey = "api_client"
)

func (s *service) fetchAPIClient(ctx context.Context, sessionCtxData *types.SessionContextData, req *http.Request) (apiClient *types.APIClient, err error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger
	tracing.AttachRequestToSpan(span, req)

	if s.useFakeData {
		apiClient = fakes.BuildFakeAPIClient()
	} else {
		apiClientID := s.routeParamManager.BuildRouteParamIDFetcher(logger, apiClientIDURLParamKey, "API client")(req)
		apiClient, err = s.dataStore.GetAPIClientByDatabaseID(ctx, apiClientID, sessionCtxData.Requester.UserID)
		if err != nil {
			return nil, observability.PrepareError(err, logger, span, "fetching API client data")
		}
	}

	return apiClient, nil
}

//go:embed templates/partials/generated/editors/api_client_editor.gotpl
var apiClientEditorTemplate string

func (s *service) buildAPIClientEditorView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)
		tracing.AttachRequestToSpan(span, req)

		sessionCtxData, err := s.sessionContextDataFetcher(req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
			http.Redirect(res, req, "/login", unauthorizedRedirectResponseCode)
			return
		}

		apiClient, err := s.fetchAPIClient(ctx, sessionCtxData, req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "fetching API client from datastore")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		tmplFuncMap := map[string]interface{}{
			"componentTitle": func(x *types.APIClient) string {
				return fmt.Sprintf("Client #%d", x.ID)
			},
		}

		if includeBaseTemplate {
			tmpl := s.parseTemplate(ctx, "", apiClientEditorTemplate, tmplFuncMap)

			s.renderTemplateToResponse(ctx, tmpl, apiClient, res)
		} else {
			view := s.renderTemplateIntoBaseTemplate(apiClientEditorTemplate, tmplFuncMap)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       fmt.Sprintf("APIClient #%d", apiClient.ID),
				ContentData: apiClient,
			}
			if sessionCtxData != nil {
				page.IsServiceAdmin = sessionCtxData.Requester.ServicePermissions.IsServiceAdmin()
			}

			s.renderTemplateToResponse(ctx, view, page, res)
		}
	}
}

func (s *service) fetchAPIClients(ctx context.Context, sessionCtxData *types.SessionContextData, req *http.Request) (apiClients *types.APIClientList, err error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger
	tracing.AttachRequestToSpan(span, req)

	if s.useFakeData {
		apiClients = fakes.BuildFakeAPIClientList()
	} else {
		filter := types.ExtractQueryFilter(req)
		apiClients, err = s.dataStore.GetAPIClients(ctx, sessionCtxData.Requester.UserID, filter)
		if err != nil {
			return nil, observability.PrepareError(err, logger, span, "fetching API client data")
		}
	}

	return apiClients, nil
}

//go:embed templates/partials/generated/tables/api_clients_table.gotpl
var apiClientsTableTemplate string

func (s *service) buildAPIClientsTableView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)
		tracing.AttachRequestToSpan(span, req)

		sessionCtxData, err := s.sessionContextDataFetcher(req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
			http.Redirect(res, req, "/login", unauthorizedRedirectResponseCode)
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
				page.IsServiceAdmin = sessionCtxData.Requester.ServicePermissions.IsServiceAdmin()
			}

			s.renderTemplateToResponse(ctx, view, page, res)
		} else {
			tmpl := s.parseTemplate(ctx, "dashboard", apiClientsTableTemplate, tmplFuncMap)

			s.renderTemplateToResponse(ctx, tmpl, apiClients, res)
		}
	}
}
