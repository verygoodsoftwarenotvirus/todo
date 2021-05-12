package frontend

import (
	"context"
	"html/template"

	// Import embed for the side effect.
	_ "embed"
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

const (
	webhookIDURLParamKey = "webhook"
)

func (s *Service) fetchWebhook(ctx context.Context, sessionCtxData *types.SessionContextData, req *http.Request) (webhook *types.Webhook, err error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger
	webhookID := s.routeParamManager.BuildRouteParamIDFetcher(logger, webhookIDURLParamKey, "webhook")(req)

	if s.useFakeData {
		webhook = fakes.BuildFakeWebhook()
	} else {
		webhook, err = s.dataStore.GetWebhook(ctx, webhookID, sessionCtxData.Requester.ID)
		if err != nil {
			return nil, observability.PrepareError(err, logger, span, "fetching webhook data")
		}
	}

	return webhook, nil
}

//go:embed templates/partials/generated/editors/webhook_editor.gotpl
var webhookEditorTemplate string

func (s *Service) buildWebhookEditorView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
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

		webhook, err := s.fetchWebhook(ctx, sessionCtxData, req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "fetching webhook from datastore")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		if includeBaseTemplate {
			view := s.renderTemplateIntoBaseTemplate(webhookEditorTemplate, nil)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       fmt.Sprintf("Webhook #%d", webhook.ID),
				ContentData: webhook,
			}
			if sessionCtxData != nil {
				page.IsServiceAdmin = sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin()
			}

			s.renderTemplateToResponse(ctx, view, page, res)
		} else {
			tmpl := s.parseTemplate(ctx, "", webhookEditorTemplate, nil)

			s.renderTemplateToResponse(ctx, tmpl, webhook, res)
		}
	}
}

func (s *Service) fetchWebhooks(ctx context.Context, sessionCtxData *types.SessionContextData, req *http.Request) (webhooks *types.WebhookList, err error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger

	if s.useFakeData {
		webhooks = fakes.BuildFakeWebhookList()
	} else {
		filter := types.ExtractQueryFilter(req)
		webhooks, err = s.dataStore.GetWebhooks(ctx, sessionCtxData.Requester.ID, filter)
		if err != nil {
			return nil, observability.PrepareError(err, logger, span, "fetching webhook data")
		}
	}

	return webhooks, nil
}

//go:embed templates/partials/generated/tables/webhooks_table.gotpl
var webhooksTableTemplate string

func (s *Service) buildWebhooksTableView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
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

		webhooks, err := s.fetchWebhooks(ctx, sessionCtxData, req)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "fetching webhooks from datastore")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		tmplFuncMap := map[string]interface{}{
			"individualURL": func(x *types.Webhook) template.URL {
				/* #nosec G203 */
				return template.URL(fmt.Sprintf("/webhooks/%d", x.ID))
			},
			"pushURL": func(x *types.Webhook) template.URL {
				/* #nosec G203 */
				return template.URL(fmt.Sprintf("/webhooks/%d", x.ID))
			},
		}

		if includeBaseTemplate {
			view := s.renderTemplateIntoBaseTemplate(webhooksTableTemplate, tmplFuncMap)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       "Webhooks",
				ContentData: webhooks,
			}
			if sessionCtxData != nil {
				page.IsServiceAdmin = sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin()
			}

			s.renderTemplateToResponse(ctx, view, page, res)
		} else {
			tmpl := s.parseTemplate(ctx, "dashboard", webhooksTableTemplate, tmplFuncMap)

			s.renderTemplateToResponse(ctx, tmpl, webhooks, res)
		}
	}
}
