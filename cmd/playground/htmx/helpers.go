package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
)

const useFakesQueryKey = "fake"

func useFakes(req *http.Request) bool {
	return parseBool(req.URL.Query().Get(useFakesQueryKey))
}

// parseBool differs from strconv.ParseBool in that it returns false by default.
func parseBool(str string) bool {
	switch str {
	case "1", "t", "T", "true", "TRUE", "True":
		return true
	default:
		return false
	}
}

func renderStringToResponse(thing string) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, _ *http.Request) {
		if _, err := res.Write([]byte(thing)); err != nil {
			log.Fatalln(err)
		}
	}
}

func renderTemplateToString(tmpl *template.Template, x interface{}) string {
	var b bytes.Buffer
	if err := tmpl.Funcs(defaultFuncMap).Execute(&b, x); err != nil {
		panic(err)
	}
	return b.String()
}
