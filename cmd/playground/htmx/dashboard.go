package main

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
)

type dashboardPageData struct {
	Title                       string
	Page                        template.HTML
	PageDescription             string
	PageTitle                   string
	PageImagePreview            string
	PageImagePreviewDescription string
}

//go:embed templates/dashboard.gotpl
var dashboardTemplateSrc string

func renderRawStringIntoDashboard(thing string) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, _ *http.Request) {
		x := &dashboardPageData{
			Title: "Dashboard",
			Page:  template.HTML(thing),
		}

		if err := template.Must(template.New("").Funcs(defaultFuncMap).Parse(dashboardTemplateSrc)).Execute(res, x); err != nil {
			log.Fatalln(err)
		}
	}
}

const dashboardPageTemplateFormat = `<div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
	<h1 class="h2">{{ .Title }}</h1>
</div>
{{ .Page }}
`

func buildDashboardSubpageString(title string, content template.HTML) string {
	x := &dashboardPageData{
		Page:  content,
		Title: title,
	}
	return renderTemplateToString(template.Must(template.New("").Parse(dashboardPageTemplateFormat)), x)
}
