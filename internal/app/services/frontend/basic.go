package frontend

import (
	"bytes"
	"fmt"
	"html/template"

	// import embed for the side effects.
	_ "embed"
)

type basicEditorField struct {
	Name      string
	InputType string
	Required  bool
}

type basicEditorTemplateConfig struct {
	FuncMap template.FuncMap
	Fields  []basicEditorField
}

//go:embed templates/basic_editor.gotpl
var basicEditorTemplateSrc string

func (s *Service) buildBasicEditorTemplate(cfg *basicEditorTemplateConfig) string {
	var b bytes.Buffer

	if err := s.parseTemplate("", basicEditorTemplateSrc, cfg.FuncMap).Execute(&b, cfg); err != nil {
		panic(err)
	}

	return b.String()
}

type basicTableTemplateConfig struct {
	FuncMap              template.FuncMap
	GetURL               string
	ExternalURL          string
	RowDataFieldName     string
	Title                string
	Columns              []string
	CellFields           []string
	ExcludeIDRow         bool
	IncludeLastUpdatedOn bool
	IncludeCreatedOn     bool
}

//go:embed templates/basic_table.gotpl
var basicTableTemplateSrc string

func (s *Service) buildBasicTableTemplate(cfg *basicTableTemplateConfig) string {
	var b bytes.Buffer

	if err := s.parseTemplate("", basicTableTemplateSrc, cfg.FuncMap).Execute(&b, cfg); err != nil {
		panic(err)
	}

	return b.String()
}

func wrapTemplateInContentDefinition(tmpl string) string {
	return fmt.Sprintf(`{{ define "content" }}
	%s
{{ end }}
	`, tmpl)
}
