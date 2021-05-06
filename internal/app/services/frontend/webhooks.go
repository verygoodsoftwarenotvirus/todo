package frontend

import (
	"fmt"
	"log"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

var webhooksTableConfig = &basicTableTemplateConfig{
	Title:       "Webhooks",
	ExternalURL: "/account/webhooks/123",
	GetURL:      "/dashboard_pages/account/webhooks/123",
	Columns:     fetchTableColumns("columns.webhooks"),
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

var webhookEditorConfig = &basicEditorTemplateConfig{
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

func (s *Service) webhooksEditorView(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var webhook *types.Webhook
	if s.useFakes {
		logger.Debug("using fakes")
		webhook = fakes.BuildFakeWebhook()
	}

	tmpl := parseTemplate("", buildBasicEditorTemplate(webhookEditorConfig), webhookEditorConfig.FuncMap)

	if err := renderTemplateToResponse(tmpl, webhook, res); err != nil {
		log.Panic(err)
	}
}

func (s *Service) webhooksTableView(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var webhooks *types.WebhookList
	if s.useFakes {
		logger.Debug("using fakes")
		webhooks = fakes.BuildFakeWebhookList()
	}

	tmpl := parseTemplate("dashboard", buildBasicTableTemplate(webhooksTableConfig), nil)

	if err := renderTemplateToResponse(tmpl, webhooks, res); err != nil {
		log.Panic(err)
	}
}

func (s *Service) webhookDashboardView(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var webhook *types.Webhook
	if s.useFakes {
		logger.Debug("using fakes")
		webhook = fakes.BuildFakeWebhook()
	}

	view := renderTemplateIntoDashboard(buildBasicEditorTemplate(webhookEditorConfig), webhookEditorConfig.FuncMap)

	page := &dashboardPageData{
		Title:       fmt.Sprintf("Webhook #%d", webhook.ID),
		ContentData: webhook,
	}

	if err := renderTemplateToResponse(view, page, res); err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering webhooks dashboard view")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Service) webhooksDashboardView(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var webhooks *types.WebhookList
	if s.useFakes {
		logger.Debug("using fakes")
		webhooks = fakes.BuildFakeWebhookList()
	}

	view := renderTemplateIntoDashboard(buildBasicTableTemplate(webhooksTableConfig), nil)

	page := &dashboardPageData{
		Title:       "Webhooks",
		ContentData: webhooks,
	}

	if err := renderTemplateToResponse(view, page, res); err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering webhooks dashboard view")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}
