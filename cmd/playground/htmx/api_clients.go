package main

import (
	"html/template"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

const apiClientsTableTemplateFormat = `<table class="table table-striped">
    <thead>
        <tr>
            <th>ID</th>
            <th>Name</th>
            <th>Client ID</th>
            <th>Belongs To User</th>
            <th>Created On</th>
        </tr>
    </thead>
    <tbody>{{ range $i, $x := .Clients }}
        <tr>
            <td><a href="" hx-get="/dashboard_pages/api_clients/123" hx-target="#content">{{ $x.ID }}</a></td>
            <td>{{ $x.Name }}</td>
            <td>{{ $x.ClientID }}</td>
            <td>{{ $x.BelongsToUser }}</td>
            <td>{{ relativeTime $x.CreatedOn }}</td>
        </tr>
    {{ end }}</tbody>
</table>
`

var apiClientsTableTemplate = template.Must(template.New("").Funcs(defaultFuncMap).Parse(apiClientsTableTemplateFormat))

func exampleAPIClientsTable() string {
	clients := fakes.BuildFakeAPIClientList()
	return renderTemplateToString(apiClientsTableTemplate, clients)
}

func apiClientsDashboardPage(res http.ResponseWriter, req *http.Request) {
	renderStringToResponse(buildDashboardSubpageString("API Clients", template.HTML(exampleAPIClientsTable())))(res, req)
}
