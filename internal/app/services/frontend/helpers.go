package frontend

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"net/url"
)

const (
	redirectToQueryKey = "redirectTo"
)

func buildRedirectURL(basePath, redirectTo string) string {
	u := &url.URL{
		Path:     basePath,
		RawQuery: url.Values{redirectToQueryKey: {redirectTo}}.Encode(),
	}

	return u.String()
}

func pluckRedirectURL(req *http.Request) string {
	return req.URL.Query().Get(redirectToQueryKey)
}

func parseListOfTemplates(funcMap template.FuncMap, name string, templates ...string) *template.Template {
	tmpl := template.New(name).Funcs(funcMap)

	for _, t := range templates {
		tmpl = template.Must(tmpl.Parse(t))
	}

	return tmpl
}

func renderStringToResponse(thing string, res http.ResponseWriter) {
	renderBytesToResponse([]byte(thing), res)
}

func renderBytesToResponse(thing []byte, res http.ResponseWriter) {
	if _, err := res.Write(thing); err != nil {
		log.Fatalln(err)
	}
}

func renderTemplateToBytes(tmpl *template.Template, x interface{}, funcMap template.FuncMap) ([]byte, error) {
	var b bytes.Buffer

	if err := tmpl.Funcs(funcMap).Execute(&b, x); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (s *Service) renderTemplateToResponse(tmpl *template.Template, x interface{}, res http.ResponseWriter) error {
	content, err := renderTemplateToBytes(tmpl, x, s.templateFuncMap)
	if err != nil {
		return err
	}

	_, err = res.Write(content)
	return err
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

// parseBool differs from strconv.ParseBool in that it returns false by default.
func parseBool(str string) bool {
	switch str {
	case "1", "t", "T", "true", "TRUE", "True":
		return true
	default:
		return false
	}
}

func isAdminRequest(req *http.Request) bool {
	return parseBool(req.URL.Query().Get("admin"))
}

func (s *Service) parseTemplate(name, source string, funcMap template.FuncMap) *template.Template {
	return template.Must(template.New(name).Funcs(mergeFuncMaps(s.templateFuncMap, funcMap)).Parse(source))
}
