package frontend

import (
	"html/template"
	// import embed for the side effect.
	_ "embed"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
)

type pageData struct {
	ContentData                 interface{}
	Title                       string
	PageDescription             string
	PageTitle                   string
	PageImagePreview            string
	PageImagePreviewDescription string
	IsLoggedIn                  bool
	IsServiceAdmin              bool
}

//go:embed templates/base_template.gotpl
var baseTemplateSrc string

func (s *Service) homepage(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		// that's okay, it's the homepage.
		_ = err
	}

	tmpl := s.renderTemplateIntoBaseTemplate("", nil)
	x := &pageData{
		IsLoggedIn:  sessionCtxData != nil,
		Title:       "Home",
		ContentData: "",
	}
	if sessionCtxData != nil {
		x.IsServiceAdmin = sessionCtxData.Requester.ServiceAdminPermission.IsServiceAdmin()
	}

	if err = s.renderTemplateToResponse(tmpl, x, res); err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering item viewer into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Service) renderTemplateIntoBaseTemplate(templateSrc string, funcMap template.FuncMap) *template.Template {
	return parseListOfTemplates(mergeFuncMaps(s.templateFuncMap, funcMap), "dashboard", baseTemplateSrc, wrapTemplateInContentDefinition(templateSrc))
}
