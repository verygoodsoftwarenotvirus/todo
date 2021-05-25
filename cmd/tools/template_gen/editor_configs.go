package main

import (
	"bytes"

	_ "embed"
)

//go:embed templates/editor.gotpl
var basicEditorTemplateSrc string

func buildBasicEditorTemplate(cfg *basicEditorTemplateConfig) string {
	var b bytes.Buffer

	if err := parseTemplate("", basicEditorTemplateSrc, nil).Execute(&b, cfg); err != nil {
		panic(err)
	}

	return b.String()
}

type basicEditorTemplateConfig struct {
	SubmissionURL string
	Fields        []formField
}

var editorConfigs = map[string]*basicEditorTemplateConfig{
	"internal/services/frontend/templates/partials/generated/editors/account_editor.gotpl": {
		Fields: []formField{
			{
				LabelName:       "name",
				FormName:        "name",
				StructFieldName: "Name",
				InputType:       "text",
				Required:        true,
			},
		},
	},
	"internal/services/frontend/templates/partials/generated/editors/account_subscription_plan_editor.gotpl": {
		Fields: []formField{
			{
				LabelName:       "name",
				FormName:        "name",
				StructFieldName: "Name",
				InputType:       "text",
				Required:        true,
			},
			{
				LabelName:       "price",
				FormName:        "price",
				StructFieldName: "Price",
				InputType:       "numeric",
				Required:        true,
			},
		},
	},
	"internal/services/frontend/templates/partials/generated/editors/api_client_editor.gotpl": {
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
	"internal/services/frontend/templates/partials/generated/editors/webhook_editor.gotpl": {
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
	"internal/services/frontend/templates/partials/generated/editors/item_editor.gotpl": {
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
				Required:        true,
			},
		},
	},
}
