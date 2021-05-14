package frontend

import (
	// import embed for the side effect.
	_ "embed"
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
)

type pageData struct {
	ContentData                 interface{}
	Title                       string
	PageDescription             string
	PageTitle                   string
	PageImagePreview            string
	PageImagePreviewDescription string
	InheritedQuery              string
	IsLoggedIn                  bool
	IsServiceAdmin              bool
}

//go:embed templates/base_template.gotpl
var baseTemplateSrc string

func (s *Service) homepage(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	tracing.AttachRequestToSpan(span, req)

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

	s.renderTemplateToResponse(ctx, tmpl, x, res)
}

func wrapTemplateInContentDefinition(tmpl string) string {
	return fmt.Sprintf(`{{ define "content" }}
	%s
{{ end }}
`, tmpl)
}
