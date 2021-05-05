package frontend

import (
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

func buildViewerForWebhook(x *types.Webhook) (string, error) {
	tmplConfig := &basicEditorTemplateConfig{
		Name: "Webhook",
		ID:   x.ID,
		Fields: []genericEditorField{
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
	}

	tmpl := parseTemplate("", buildBasicEditorTemplate(tmplConfig))

	return renderTemplateToString(tmpl, x)
}

func (s *Service) webhookDashboardPage(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var webhook *types.Webhook
	if s.useFakes {
		logger.Debug("using fakes")
		webhook = fakes.BuildFakeWebhook()
	}

	page, err := buildViewerForWebhook(webhook)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering webhook table template into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderStringToResponse(page, res)
}

func buildWebhooksTableDashboardPage(webhooks *types.WebhookList) (string, error) {
	tmpl := parseTemplate("dashboard", buildBasicTableTemplate(webhooksTableConfig))

	return renderTemplateToString(tmpl, webhooks)
}

func (s *Service) webhooksDashboardPage(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var webhooks *types.WebhookList
	if s.useFakes {
		logger.Debug("using fakes")
		webhooks = fakes.BuildFakeWebhookList()
	}

	page, err := buildWebhooksTableDashboardPage(webhooks)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering webhook table template into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderStringToResponse(page, res)
}

func buildWebhookDashboardView(x *types.Webhook) (string, error) {
	tmplConfig := &basicEditorTemplateConfig{
		Name: "Webhook",
		ID:   x.ID,
		Fields: []genericEditorField{
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
	}

	return renderTemplateIntoDashboard("Webhooks", wrapTemplateInContentDefinition(buildBasicEditorTemplate(tmplConfig)), x)
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

	dashboard, err := buildWebhookDashboardView(webhook)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering webhook viewer into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderStringToResponse(dashboard, res)
}

func buildWebhooksTableDashboardView(webhooks *types.WebhookList) (string, error) {
	tmpl := wrapTemplateInContentDefinition(buildBasicTableTemplate(webhooksTableConfig))

	return renderTemplateIntoDashboard("Webhooks", tmpl, webhooks)
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

	dashboard, err := buildWebhooksTableDashboardView(webhooks)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering webhook table template into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderStringToResponse(dashboard, res)
}
