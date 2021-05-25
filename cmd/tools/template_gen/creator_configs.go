package main

import (
	"bytes"

	_ "embed"
)

//go:embed templates/creator.gotpl
var basicCreatorTemplateSrc string

func buildBasicCreatorTemplate(cfg *basicCreatorTemplateConfig) string {
	var b bytes.Buffer

	if err := parseTemplate("", basicCreatorTemplateSrc, nil).Execute(&b, cfg); err != nil {
		panic(err)
	}

	return b.String()
}

type basicCreatorTemplateConfig struct {
	Title         string
	SubmissionURL string
	Fields        []formField
}

var creatorConfigs = map[string]*basicCreatorTemplateConfig{
	"internal/services/frontend/templates/partials/generated/creators/account_creator.gotpl": {
		Title:         "New Account",
		SubmissionURL: "/accounts/new/submit",
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
	"internal/services/frontend/templates/partials/generated/creators/api_client_creator.gotpl": {
		Title:         "New API Client",
		SubmissionURL: "/api_clients/new/submit",
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
	"internal/services/frontend/templates/partials/generated/creators/webhook_creator.gotpl": {
		Title:         "New Webhook",
		SubmissionURL: "/webhooks/new/submit",
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
	"internal/services/frontend/templates/partials/generated/creators/item_creator.gotpl": {
		Title:         "New Item",
		SubmissionURL: "/items/new/submit",
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
