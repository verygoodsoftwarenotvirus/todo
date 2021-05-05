package frontend

import (
	"bytes"
	"fmt"

	// import embed for the side effects.
	_ "embed"
)

type genericEditorField struct {
	Name      string
	InputType string
	Required  bool
}

type basicEditorTemplateConfig struct {
	Name   string
	Fields []genericEditorField
	ID     uint64
}

//go:embed templates/basic_editor.gotpl
var basicEditorTemplateSrc string

func buildBasicEditorTemplate(cfg *basicEditorTemplateConfig) string {
	var b bytes.Buffer

	if err := parseTemplate("", basicEditorTemplateSrc).Execute(&b, cfg); err != nil {
		panic(err)
	}

	return b.String()
}

type basicTableTemplateConfig struct {
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

func buildBasicTableTemplate(cfg *basicTableTemplateConfig) string {
	var b bytes.Buffer

	if err := parseTemplate("", basicTableTemplateSrc).Execute(&b, cfg); err != nil {
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
