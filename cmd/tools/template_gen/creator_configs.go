package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

//go:embed templates/creator.gotpl
var basicCreatorTemplateSrc string

func buildBasicCreatorTemplate(cfg *basicCreatorTemplateConfig) string {
	var b bytes.Buffer

	if err := parseTemplate("", basicCreatorTemplateSrc, cfg.FuncMap).Execute(&b, cfg); err != nil {
		panic(err)
	}

	return b.String()
}

type basicCreatorTemplateConfig struct {
	Title         string
	FuncMap       template.FuncMap
	Fields        []formField
	SubmissionURL string
}

var creatorConfigs = map[string]*basicCreatorTemplateConfig{
	"internal/app/services/frontend/templates/partials/creators/account_creator.gotpl": {
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
		FuncMap: map[string]interface{}{
			"componentTitle": func(x *types.Account) string {
				return fmt.Sprintf("Account #%d", x.ID)
			},
		},
	},
	"internal/app/services/frontend/templates/partials/creators/api_client_creator.gotpl": {
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
	"internal/app/services/frontend/templates/partials/creators/webhook_creator.gotpl": {
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
	"internal/app/services/frontend/templates/partials/creators/item_creator.gotpl": {
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
