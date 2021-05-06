package frontend

import (
	"bytes"
	"html/template"

	// import embed for the side effects.
	_ "embed"
	"net/http"
)

type dashboardPageData struct {
	Title                       string
	ContentData                 interface{}
	PageDescription             string
	PageTitle                   string
	PageImagePreview            string
	PageImagePreviewDescription string
}

//go:embed templates/dashboard.gotpl
var dashboardTemplateSrc string

func (s *Service) homepage(res http.ResponseWriter, _ *http.Request) {
	dash, err := renderTemplateIntoDashboardAsString("Home", "", "", nil)
	if err != nil {
		panic(err)
	}

	renderStringToResponse(dash, res)
}

func renderTemplateIntoDashboard(templateSrc string, funcMap template.FuncMap) *template.Template {
	return parseListOfTemplates(funcMap, "dashboard", dashboardTemplateSrc, wrapTemplateInContentDefinition(templateSrc))
}

func renderTemplateIntoDashboardAsString(title, templateSrc string, contentData interface{}, funcMap template.FuncMap) (string, error) {
	x := &dashboardPageData{
		Title:       title,
		ContentData: contentData,
	}

	tmpl := parseListOfTemplates(funcMap, "dashboard", dashboardTemplateSrc, templateSrc)

	var b bytes.Buffer
	if err := tmpl.Execute(&b, x); err != nil {
		return "", err
	}

	return b.String(), nil
}
