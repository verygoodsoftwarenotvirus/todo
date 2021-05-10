package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

//go:embed templates/editor.gotpl
var basicEditorTemplateSrc string

func buildBasicEditorTemplate(cfg *basicEditorTemplateConfig) string {
	var b bytes.Buffer

	if err := parseTemplate("", basicEditorTemplateSrc, cfg.FuncMap).Execute(&b, cfg); err != nil {
		panic(err)
	}

	return b.String()
}

type basicEditorTemplateConfig struct {
	FuncMap       template.FuncMap
	Fields        []formField
	SubmissionURL string
}

var editorConfigs = map[string]*basicEditorTemplateConfig{
	"internal/app/services/frontend/templates/partials/editors/account_editor.gotpl": {
		Fields: []formField{
			{
				LabelName:       "name",
				FormName:        "name",
				StructFieldName: "Name",
				InputType:       "text",
				Required:        true,
			},
		},
		FuncMap: map[string]interface{}{
			"componentTitle": func(x *types.Account) string {
				return fmt.Sprintf("Account #%d", x.ID)
			},
		},
	},
	"internal/app/services/frontend/templates/partials/editors/api_client_editor.gotpl": {
		Fields: []formField{
			{
				LabelName:       "name",
				FormName:        "name",
				StructFieldName: "Name",
				InputType:       "text",
				Required:        true,
			},
			{
				LabelName:       "client_id",
				FormName:        "client_id",
				StructFieldName: "ClientID",
				InputType:       "text",
				Required:        true,
			},
			{
				LabelName:       "external ID",
				FormName:        "external_id",
				StructFieldName: "ExternalID",
				InputType:       "text",
				Required:        true,
			},
		},
	},
	"internal/app/services/frontend/templates/partials/editors/webhook_editor.gotpl": {
		Fields: []formField{
			{
				LabelName:       "name",
				StructFieldName: "Name",
				InputType:       "text",
				Required:        true,
			},
			{
				LabelName:       "Method",
				StructFieldName: "Method",
				InputType:       "text",
				Required:        true,
			},
			{
				LabelName:       "ContentType",
				StructFieldName: "ContentType",
				InputType:       "text",
				Required:        true,
			},
			{
				LabelName:       "URL",
				StructFieldName: "URL",
				InputType:       "text",
				Required:        true,
			},
		},
	},
	"internal/app/services/frontend/templates/partials/editors/item_editor.gotpl": {
		Fields: []formField{
			{
				LabelName:       "name",
				FormName:        "name",
				StructFieldName: "Name",
				InputType:       "text",
				Required:        true,
			},
			{
				LabelName:       "details",
				FormName:        "details",
				StructFieldName: "Details",
				InputType:       "text",
				Required:        false,
			},
		},
		FuncMap: map[string]interface{}{
			"componentTitle": func(x *types.Item) string {
				return fmt.Sprintf("Item #%d", x.ID)
			},
		},
	},
}
