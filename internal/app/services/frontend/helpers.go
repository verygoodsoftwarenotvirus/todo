package frontend

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
)

func parseListOfTemplates(name string, templates ...string) *template.Template {
	tmpl := template.New(name).Funcs(defaultFuncMap)

	for _, t := range templates {
		tmpl = template.Must(tmpl.Parse(t))
	}

	return tmpl
}

func renderStringToResponse(thing string, res http.ResponseWriter) {
	if _, err := res.Write([]byte(thing)); err != nil {
		log.Fatalln(err)
	}
}

func renderTemplateToString(tmpl *template.Template, x interface{}) (string, error) {
	var b bytes.Buffer

	if err := tmpl.Funcs(defaultFuncMap).Execute(&b, x); err != nil {
		return "", err
	}

	return b.String(), nil
}

func parseTemplate(name, source string) *template.Template {
	return template.Must(template.New(name).Funcs(defaultFuncMap).Parse(source))
}
