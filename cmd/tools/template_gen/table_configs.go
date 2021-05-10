package main

import (
	"bytes"
	_ "embed"
	"text/template"
)

type basicTableTemplateConfig struct {
	FuncMap              template.FuncMap
	GetURL               string
	CreatorPageURL       string
	RowDataFieldName     string
	Title                string
	Columns              []string
	CellFields           []string
	ExcludeIDRow         bool
	IncludeLastUpdatedOn bool
	IncludeCreatedOn     bool
	IncludeDeleteRow     bool
}

//go:embed templates/table.gotpl
var basicTableTemplateSrc string

func buildBasicTableTemplate(cfg *basicTableTemplateConfig) string {
	var b bytes.Buffer

	if err := parseTemplate("", basicTableTemplateSrc, cfg.FuncMap).Execute(&b, cfg); err != nil {
		panic(err)
	}

	return b.String()
}

var tableConfigs = map[string]*basicTableTemplateConfig{
	"internal/app/services/frontend/templates/partials/tables/api_clients_table.gotpl": {
		Title:          "API Clients",
		CreatorPageURL: "/api_clients/new",
		GetURL:         "/dashboard_pages/api_clients/123",
		Columns: []string{
			"ID",
			"Name",
			"External ID",
			"Client ID",
			"Belongs To User",
			"Created On",
		},
		CellFields: []string{
			"ID",
			"Name",
			"ExternalID",
			"ClientID",
			"BelongsToUser",
			"CreatedOn",
		},
		RowDataFieldName:     "Clients",
		IncludeLastUpdatedOn: false,
		IncludeCreatedOn:     true,
	},
	"internal/app/services/frontend/templates/partials/tables/accounts_table.gotpl": {
		Title:          "Accounts",
		CreatorPageURL: "/accounts/new",
		GetURL:         "/dashboard_pages/accounts/123",
		Columns: []string{
			"ID",
			"Name",
			"External ID",
			"Belongs To User",
			"Last Updated On",
			"Created On",
		},
		CellFields: []string{
			"Name",
			"ExternalID",
			"BelongsToUser",
		},
		RowDataFieldName:     "Accounts",
		IncludeLastUpdatedOn: true,
		IncludeCreatedOn:     true,
	},
	"internal/app/services/frontend/templates/partials/tables/webhooks_table.gotpl": {
		Title:          "Webhooks",
		CreatorPageURL: "/accounts/webhooks/new",
		GetURL:         "/dashboard_pages/account/webhooks/123",
		Columns: []string{
			"ID",
			"Name",
			"Method",
			"URL",
			"Content Type",
			"Belongs To Account",
			"Last Updated On",
			"Created On",
		},
		CellFields: []string{
			"Name",
			"Method",
			"URL",
			"ContentType",
			"BelongsToAccount",
		},
		RowDataFieldName:     "Webhooks",
		IncludeLastUpdatedOn: true,
		IncludeCreatedOn:     true,
	},
	"internal/app/services/frontend/templates/partials/tables/items_table.gotpl": {
		Title:          "Items",
		CreatorPageURL: "/items/new",
		Columns: []string{
			"ID",
			"Name",
			"Details",
			"Last Updated On",
			"Created On",
			"",
		},
		CellFields: []string{
			"Name",
			"Details",
		},
		RowDataFieldName:     "Items",
		IncludeLastUpdatedOn: true,
		IncludeCreatedOn:     true,
		IncludeDeleteRow:     true,
	},
}
