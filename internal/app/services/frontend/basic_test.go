package frontend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_buildBasicEditorTemplate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleInput := &basicEditorTemplateConfig{
			Name: "Item",
			ID:   12345,
			Fields: []genericEditorField{
				{
					Name:      "Name",
					InputType: "text",
					Required:  true,
				},
				{
					Name:      "Details",
					InputType: "text",
					Required:  false,
				},
			},
		}

		expected := `<div id="content" class="">
    <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
        <h1 class="h2">Item #{{ .ID }}</h1>
    </div>
    <div class="col-md-8 order-md-1">
        <form class="needs-validation" novalidate="">
            <div class="mb3">
                <label for="Name">Name</label>
                <div class="input-group">
                    <input class="form-control" type="text" id="Name" placeholder="Name"required="" value="{{ .Name }}" />
                    <div class="invalid-feedback" style="width: 100%;">Name is required.</div>
                </div>
            </div>
            <div class="mb3">
                <label for="Details">Details</label>
                <div class="input-group">
                    <input class="form-control" type="text" id="Details" placeholder="Details" value="{{ .Details }}" />
                    
                </div>
            </div>
            <hr class="mb-4" />
            <button class="btn btn-primary btn-lg btn-block" type="submit">Save</button>
        </form>
    </div>
</div>`

		actual := buildBasicEditorTemplate(exampleInput)

		assert.Equal(t, expected, actual)
	})
}

func Test_buildBasicTableTemplate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleInput := &basicTableTemplateConfig{
			GetURL: "/dashboard_pages/items/123",
			Columns: []string{
				"ID",
				"Name",
				"Details",
				"Belongs To Account",
				"Last Updated On",
				"Created On",
			},
			CellFields: []string{
				"Name",
				"Details",
				"BelongsToAccount",
			},
			RowDataFieldName:     "Items",
			IncludeLastUpdatedOn: true,
			IncludeCreatedOn:     true,
		}

		expected := `<div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
    <h1 class="h2"></h1>
</div>
<table class="table table-striped">
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
        <td><a href="" hx-push-url="" hx-get="/dashboard_pages/items/123" hx-target="#content">{{ $x.ID }}</a></td>
        <td>{{ $x.Name }}</td>
        <td>{{ $x.Details }}</td>
        <td>{{ $x.BelongsToAccount }}</td>
        <td>{{ relativeTimeFromPtr $x.LastUpdatedOn }}</td>
        <td>{{ relativeTime $x.CreatedOn }}</td>
    </tr>
    {{ end }}</tbody>
</table>`

		actual := buildBasicTableTemplate(exampleInput)

		assert.Equal(t, expected, actual)
	})
}
