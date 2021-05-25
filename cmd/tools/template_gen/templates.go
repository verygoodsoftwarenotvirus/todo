package main

import (
	"text/template"

	_ "embed"
)

func parseTemplate(name, source string, funcMap template.FuncMap) *template.Template {
	return template.Must(template.New(name).Funcs(mergeFuncMaps(defaultTemplateFuncMap, funcMap)).Parse(source))
}

func mergeFuncMaps(a, b template.FuncMap) template.FuncMap {
	out := map[string]interface{}{}

	for k, v := range a {
		out[k] = v
	}

	for k, v := range b {
		out[k] = v
	}

	return out
}

type formField struct {
	LabelName        string
	StructFieldName  string
	FormName         string
	TagID            string
	InputType        string
	InputPlaceholder string
	Required         bool
}
