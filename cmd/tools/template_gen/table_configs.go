package main

import (
	"bytes"

	_ "embed"
)

type basicTableTemplateConfig struct {
	SearchURL            string
	CreatorPageURL       string
	RowDataFieldName     string
	Title                string
	CreatorPagePushURL   string
	CellFields           []string
	Columns              []string
	EnableSearch         bool
	ExcludeIDRow         bool
	ExcludeLink          bool
	IncludeLastUpdatedOn bool
	IncludeCreatedOn     bool
	IncludeDeleteRow     bool
}

//go:embed templates/table.gotpl
var basicTableTemplateSrc string

func buildBasicTableTemplate(cfg *basicTableTemplateConfig) string {
	var b bytes.Buffer

	if err := parseTemplate("", basicTableTemplateSrc, nil).Execute(&b, cfg); err != nil {
		panic(err)
	}

	return b.String()
}

var tableConfigs = map[string]*basicTableTemplateConfig{
	"internal/services/frontend/templates/partials/generated/tables/api_clients_table.gotpl": {
		Title:              "API Clients",
		CreatorPagePushURL: "/api_clients/new",
		CreatorPageURL:     "/dashboard_pages/api_clients/new",
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
	"internal/services/frontend/templates/partials/generated/tables/accounts_table.gotpl": {
		Title:              "Accounts",
		CreatorPagePushURL: "/accounts/new",
		CreatorPageURL:     "/dashboard_pages/accounts/new",
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
	"internal/services/frontend/templates/partials/generated/tables/users_table.gotpl": {
		Title: "Users",
		Columns: []string{
			"ID",
			"Username",
			"Last Updated On",
			"Created On",
		},
		CellFields: []string{
			"Username",
		},
		EnableSearch:         true,
		RowDataFieldName:     "Users",
		IncludeLastUpdatedOn: true,
		IncludeCreatedOn:     true,
		IncludeDeleteRow:     false,
		ExcludeLink:          true,
	},
	"internal/services/frontend/templates/partials/generated/tables/webhooks_table.gotpl": {
		Title:              "Webhooks",
		CreatorPagePushURL: "/accounts/webhooks/new",
		CreatorPageURL:     "/dashboard_pages/accounts/webhooks/new",
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
	"internal/services/frontend/templates/partials/generated/tables/items_table.gotpl": {
		Title:              "Items",
		CreatorPagePushURL: "/items/new",
		CreatorPageURL:     "/dashboard_pages/items/new",
		Columns: []string{
			"ID",
			"Name",
			"Details",
			"Last Updated On",
			"Created On",
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
