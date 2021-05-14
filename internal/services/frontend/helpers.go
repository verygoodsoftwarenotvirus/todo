package frontend

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"net/http"
	"net/url"

	observability "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
)

const (
	redirectToQueryKey = "redirectTo"

	htmxRedirectionHeader = "HX-Redirect"
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

func (s *Service) renderStringToResponse(thing string, res http.ResponseWriter) {
	s.renderBytesToResponse([]byte(thing), res)
}

func (s *Service) renderBytesToResponse(thing []byte, res http.ResponseWriter) {
	if _, err := res.Write(thing); err != nil {
		s.logger.Error(err, "writing response")
	}
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

func (s *Service) renderTemplateIntoBaseTemplate(templateSrc string, funcMap template.FuncMap) *template.Template {
	return parseListOfTemplates(mergeFuncMaps(s.templateFuncMap, funcMap), "dashboard", baseTemplateSrc, wrapTemplateInContentDefinition(templateSrc))
}

func (s *Service) renderTemplateToResponse(ctx context.Context, tmpl *template.Template, x interface{}, res http.ResponseWriter) {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	var b bytes.Buffer
	if err := tmpl.Funcs(s.templateFuncMap).Execute(&b, x); err != nil {
		observability.AcknowledgeError(err, s.logger, span, "rendering accounts dashboard view")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.renderBytesToResponse(b.Bytes(), res)
}

func (s *Service) parseTemplate(ctx context.Context, name, source string, funcMap template.FuncMap) *template.Template {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	return template.Must(template.New(name).Funcs(mergeFuncMaps(s.templateFuncMap, funcMap)).Parse(source))
}
