package frontend

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
)

const (
	redirectToQueryKey = "redirectTo"

	htmxRedirectionHeader = "HX-Redirect"
)

func (s *Service) extractFormFromRequest(ctx context.Context, req *http.Request) (url.Values, error) {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithRequest(req)

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "reading form from request")
	}

	form, err := url.ParseQuery(string(bodyBytes))
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "parsing request form")
	}

	return form, nil
}

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

func htmxRedirectTo(res http.ResponseWriter, path string) {
	res.Header().Set(htmxRedirectionHeader, path)
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

func (s *Service) parseTemplate(name, source string, funcMap template.FuncMap) *template.Template {
	return template.Must(template.New(name).Funcs(mergeFuncMaps(s.templateFuncMap, funcMap)).Parse(source))
}
