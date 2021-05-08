package frontend

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

const (
	webhookIDURLParamKey = "webhook"
)

func (s *Service) buildWebhooksTableConfig() *basicTableTemplateConfig {
	return &basicTableTemplateConfig{
		Title:          "Webhooks",
		ExternalURL:    "/account/webhooks/123",
		CreatorPageURL: "/accounts/webhooks/new",
		GetURL:         "/dashboard_pages/account/webhooks/123",
		Columns:        s.fetchTableColumns("columns.webhooks"),
		CellFields: []string{
			"Name",
			"Method",
			"URL",
			"ContentType",
			"BelongsToAccount",
		},
		RowDataFieldName:     "Webhooks",
		IncludeLastUpdatedOn: true,
		IncludeCreatedOn:     true,
	}
}

func (s *Service) buildWebhookEditorConfig() *basicEditorTemplateConfig {
	return &basicEditorTemplateConfig{
		Fields: []basicEditorField{
			{
				Name:      "Name",
				InputType: "text",
				Required:  true,
			},
			{
				Name:      "Method",
				InputType: "text",
				Required:  true,
			},
			{
				Name:      "ContentType",
				InputType: "text",
				Required:  true,
			},
			{
				Name:      "URL",
				InputType: "text",
				Required:  true,
			},
		},
		FuncMap: map[string]interface{}{
			"componentTitle": func(x *types.Webhook) string {
				return fmt.Sprintf("Webhook #%d", x.ID)
			},
		},
	}
}

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
			observability.AcknowledgeError(err, logger, span, "error fetching webhook from datastore")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		webhookEditorConfig := s.buildWebhookEditorConfig()
		if includeBaseTemplate {
			view := s.renderTemplateIntoBaseTemplate(s.buildBasicEditorTemplate(webhookEditorConfig), webhookEditorConfig.FuncMap)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       fmt.Sprintf("Webhook #%d", webhook.ID),
				ContentData: webhook,
			}
			if sessionCtxData != nil {
				page.IsServiceAdmin = sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin()
			}

			if err = s.renderTemplateToResponse(view, page, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering webhooks dashboard view")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			tmpl := s.parseTemplate("", s.buildBasicEditorTemplate(webhookEditorConfig), webhookEditorConfig.FuncMap)

			if err := s.renderTemplateToResponse(tmpl, webhook, res); err != nil {
				log.Panic(err)
			}
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
			observability.AcknowledgeError(err, logger, span, "error fetching webhooks from datastore")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		webhooksTableConfig := s.buildWebhooksTableConfig()
		if includeBaseTemplate {
			view := s.renderTemplateIntoBaseTemplate(s.buildBasicTableTemplate(webhooksTableConfig), nil)

			page := &pageData{
				IsLoggedIn:  sessionCtxData != nil,
				Title:       "Webhooks",
				ContentData: webhooks,
			}
			if sessionCtxData != nil {
				page.IsServiceAdmin = sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin()
			}

			if err = s.renderTemplateToResponse(view, page, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering webhooks dashboard view")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			tmpl := s.parseTemplate("dashboard", s.buildBasicTableTemplate(webhooksTableConfig), nil)

			if err := s.renderTemplateToResponse(tmpl, webhooks, res); err != nil {
				log.Panic(err)
			}
		}
	}
}
