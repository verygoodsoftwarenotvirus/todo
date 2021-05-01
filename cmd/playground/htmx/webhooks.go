package main

import (
	"html/template"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

const webhooksTableTemplateFormat = `<table class="table table-striped">
    <thead>
        <tr>
            <th>ID</th>
            <th>Name</th>
            <th>External ID</th>
            <th>URL</th>
            <th>Content Type</th>
            <th>Belongs To Account</th>
            <th>Created On</th>
        </tr>
    </thead>
    <tbody>{{ range $i, $x := .Webhooks }}
        <tr>
            <td><a href="" hx-get="/dashboard_pages/webhooks/123" hx-target="#content">{{ $x.ID }}</a></td>
            <td>{{ $x.Name }}</td>
			<td>{{ $x.ExternalID }}</td>
			<td>{{ $x.URL }}</td>
			<td>{{ $x.ContentType }}</td>
            <td>{{ $x.BelongsToAccount }}</td>
            <td>{{ relativeTimeFromPtr $x.LastUpdatedOn }}</td>
            <td>{{ relativeTime $x.CreatedOn }}</td>
        </tr>
    {{ end }}</tbody>
</table>
`

var webhooksTableTemplate = template.Must(template.New("").Funcs(defaultFuncMap).Parse(webhooksTableTemplateFormat))

func exampleWebhooksTable() string {
	webhooks := fakes.BuildFakeWebhookList()
	return renderTemplateToString(webhooksTableTemplate, webhooks)
}

func webhooksDashboardPage(res http.ResponseWriter, req *http.Request) {
	renderStringToResponse(buildDashboardSubpageString("Webhooks", template.HTML(exampleWebhooksTable())))(res, req)
}
