package frontend

import (
	"bytes"
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

func (s *Service) homepage(res http.ResponseWriter, req *http.Request) {
	dash, err := renderTemplateIntoDashboard("Home", "", "")
	if err != nil {
		panic(err)
	}

	renderStringToResponse(dash, res)
}

func renderTemplateIntoDashboard(title, templateSrc string, contentData interface{}) (string, error) {
	x := &dashboardPageData{
		Title:       title,
		ContentData: contentData,
	}

	tmpl := parseListOfTemplates("dashboard", dashboardTemplateSrc, templateSrc)

	var b bytes.Buffer
	if err := tmpl.Funcs(defaultFuncMap).Execute(&b, x); err != nil {
		return "", err
	}

	return b.String(), nil
}
