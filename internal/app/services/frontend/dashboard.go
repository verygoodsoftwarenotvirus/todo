package frontend

import (
	"html/template"
	// import embed for the side effects.
	_ "embed"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
)

type dashboardPageData struct {
	ContentData                 interface{}
	Title                       string
	PageDescription             string
	PageTitle                   string
	PageImagePreview            string
	PageImagePreviewDescription string
	LoggedIn                    bool
}

//go:embed templates/dashboard.gotpl
var dashboardTemplateSrc string

func (s *Service) homepage(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		// that's okay, it's the homepage.
		_ = err
	}

	tmpl := s.renderTemplateIntoDashboard("", nil)
	x := &dashboardPageData{
		LoggedIn:    sessionCtxData != nil,
		Title:       "Home",
		ContentData: "",
	}

	if err = s.renderTemplateToResponse(tmpl, x, res); err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering item viewer into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Service) renderTemplateIntoDashboard(templateSrc string, funcMap template.FuncMap) *template.Template {
	return parseListOfTemplates(mergeFuncMaps(s.templateFuncMap, funcMap), "dashboard", dashboardTemplateSrc, wrapTemplateInContentDefinition(templateSrc))
}
