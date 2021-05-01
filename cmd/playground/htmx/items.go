package main

import (
	"bytes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"html/template"
	"net/http"
)

func itemsDashboardPage(res http.ResponseWriter, req *http.Request) {
	renderStringToResponse(buildDashboardSubpageString("Items", template.HTML(buildItemsTable())))(res, req)
}

const itemViewerTemplate = `<div id="content" class="">
    <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom"><h1 class="h2">Item #1627715355</h1></div>
    <div class="col-md-8 order-md-1">
        <form class="needs-validation" novalidate="">
            <div class="mb3">
                <label for="name">Name</label>
                <div class="input-group">
                    <input class="form-control" id="name" placeholder="Name" required="" value="{{ .Name }}" />
                    <div class="invalid-feedback" style="width: 100%;">Name is required.</div>
                </div>
            </div>
            <div class="mb-3">
				<label for="details">Details</label>
				<input class="form-control" id="details" placeholder="Details" value="{{ .Details }}" />
			</div>
            <hr class="mb-4" />
            <button class="btn btn-primary btn-lg btn-block" type="submit">Save</button>
        </form>
    </div>
</div>`

var itemEditorTemplate = template.Must(template.New("").Parse(itemViewerTemplate))

func buildItemViewer(x *types.Item) string {
	var b bytes.Buffer
	if err := itemEditorTemplate.Execute(&b, x); err != nil {
		panic(err)
	}
	return b.String()
}

const itemsTableTemplateFormat = `<table class="table table-striped">
    <thead>
        <tr>
            <th>ID</th>
            <th>Name</th>
            <th>Details</th>
            <th>Belongs To Account</th>
            <th>Last Updated On</th>
            <th>Created On</th>
        </tr>
    </thead>
    <tbody>{{ range $i, $x := .Items }}
        <tr>
            <td><a href="" hx-get="/dashboard_pages/items/123" hx-target="#content">{{ $x.ID }}</a></td>
            <td>{{ $x.Name }}</td>
            <td>{{ $x.Details }}</td>
            <td>{{ $x.BelongsToAccount }}</td>
            <td>{{ relativeTimeFromPtr $x.LastUpdatedOn }}</td>
            <td>{{ relativeTime $x.CreatedOn }}</td>
        </tr>
    {{ end }}</tbody>
</table>
`

var defaultFuncMap = map[string]interface{}{
	"relativeTime":        relativeTime,
	"relativeTimeFromPtr": relativeTimeFromPtr,
}

var itemsTableTemplate = template.Must(template.New("").Funcs(defaultFuncMap).Parse(itemsTableTemplateFormat))

func buildItemsTable() string {
	items := fakes.BuildFakeItemList()
	return renderTemplateToString(itemsTableTemplate, items)
}
