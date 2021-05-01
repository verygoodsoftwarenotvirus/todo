package main

import (
	"html/template"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

const accountsTableTemplateFormat = `<table class="table table-striped">
    <thead>
        <tr>
            <th>ID</th>
            <th>Name</th>
            <th>External ID</th>
            <th>Belongs To User</th>
            <th>Last Updated On</th>
            <th>Created On</th>
        </tr>
    </thead>
    <tbody>{{ range $i, $x := .Accounts }}
        <tr>
            <td><a href="" hx-get="/dashboard_pages/accounts/123" hx-target="#content">{{ $x.ID }}</a></td>
            <td>{{ $x.Name }}</td>
            <td>{{ $x.ExternalID }}</td>
            <td>{{ $x.BelongsToUser }}</td>
            <td>{{ relativeTimeFromPtr $x.LastUpdatedOn }}</td>
            <td>{{ relativeTime $x.CreatedOn }}</td>
        </tr>
    {{ end }}</tbody>
</table>
`

var accountsTableTemplate = template.Must(template.New("").Funcs(defaultFuncMap).Parse(accountsTableTemplateFormat))

func exampleAccountsTable() string {
	accounts := fakes.BuildFakeAccountList()
	return renderTemplateToString(accountsTableTemplate, accounts)
}

func accountsDashboardPage(res http.ResponseWriter, req *http.Request) {
	renderStringToResponse(buildDashboardSubpageString("Accounts", template.HTML(exampleAccountsTable())))(res, req)
}
