package main

import (
	"bytes"
	_ "embed"
	"html/template"
)

const baseTemplate = `{{ define "base" }}
<html>	
	<head>
		<title>{{ .Title }}</title>
	</head>
	<body>
		{{ template "content" .ContentData }}
	</body>
</html>
{{ end }}`

type thing struct {
	ContentData interface{}
	Title       string
}

const body1Template = `{{ define "content" }}
		<div>
			{{ .Thing }} <img src="https://balls.com/dong.bmp">
		</div>
{{ end }}`

const body2Template = `{{ define "content" }}
		<div>
			{{ .Stuff }} <input type=password>
		</div>
{{ end }}`

var defaultFuncMap = map[string]interface{}{
	"translate": func(s string) string {
		return s
	},
}

func parseListOfTemplates(name string, templates ...string) *template.Template {
	tmpl := template.New(name).Funcs(defaultFuncMap)

	for _, t := range templates {
		tmpl = template.Must(tmpl.Parse(t))
	}

	return tmpl
}

//go:embed templates/dashboard.gotpl
var dashboardTemplateSrc string

func main() {
	t := thing{
		Title: "fart",
		ContentData: struct {
			Thing string
			Stuff string
		}{
			Thing: "thing",
			Stuff: "stuff",
		},
	}

	tmpl := parseListOfTemplates("dashboard", dashboardTemplateSrc, body2Template)

	var b bytes.Buffer
	if err := tmpl.Execute(&b, &t); err != nil {
		panic(err)
	}

	output := b.String()

	println(output)
}
