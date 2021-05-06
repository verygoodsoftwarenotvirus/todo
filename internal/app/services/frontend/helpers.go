package frontend

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
)

func parseListOfTemplates(funcMap template.FuncMap, name string, templates ...string) *template.Template {
	tmpl := template.New(name).Funcs(appendToDefaultFuncMap(funcMap))

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

func renderTemplateToBytes(tmpl *template.Template, x interface{}) ([]byte, error) {
	var b bytes.Buffer

	if err := tmpl.Funcs(defaultFuncMap).Execute(&b, x); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func renderTemplateToString(tmpl *template.Template, x interface{}) (string, error) {
	out, err := renderTemplateToBytes(tmpl, x)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func renderTemplateToResponse(tmpl *template.Template, x interface{}, res http.ResponseWriter) error {
	content, err := renderTemplateToBytes(tmpl, x)
	if err != nil {
		return err
	}

	_, err = res.Write(content)
	return err
}

func appendToDefaultFuncMap(input template.FuncMap) template.FuncMap {
	out := map[string]interface{}{}

	for k, v := range defaultFuncMap {
		out[k] = v
	}

	for k, v := range input {
		out[k] = v
	}

	return out
}

func parseTemplate(name, source string, funcMap template.FuncMap) *template.Template {
	return template.Must(template.New(name).Funcs(appendToDefaultFuncMap(funcMap)).Parse(source))
}
