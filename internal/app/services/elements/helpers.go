package elements

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
)

func useFakes(_ *http.Request) bool {
	return true

	// you should switch strings.TrimSpace(strings.ToLower(req.URL.Query().Get("fake"))) here
}

func renderHTMLTemplateToResponse(thing template.HTML) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, _ *http.Request) {
		if _, err := res.Write([]byte(thing)); err != nil {
			log.Fatalln(err)
		}
	}
}

func renderTemplateToHTML(tmpl *template.Template, x interface{}) template.HTML {
	var b bytes.Buffer

	if err := tmpl.Funcs(defaultFuncMap).Execute(&b, x); err != nil {
		panic(err)
	}

	return template.HTML(b.String())
}
