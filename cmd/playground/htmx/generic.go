package main

import (
	"bytes"
	"html/template"
)

type genericEditorField struct {
	Name      string
	InputType string
	Required  bool
}

type genericEditorTemplateConfig struct {
	ID     uint64
	Name   string
	Fields []genericEditorField
}

const genericEditorTemplateSrc = `<div id="content" class="">
    <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
		<h1 class="h2">{{ .Name }} #{{ "{{ .ID }}" }}</h1>
	</div>
    <div class="col-md-8 order-md-1">
        <form class="needs-validation" novalidate="">{{ range $i, $field := .Fields }}
            <div class="mb3">
                <label for="{{ $field.Name }}">{{ $field.Name }}</label>
                <div class="input-group">
                    <input class="form-control" {{- if ne $field.InputType "" }} type="{{ $field.InputType }}"{{ end}} id="{{ $field.Name }}" placeholder="{{ $field.Name }}" {{- if $field.Required }}required=""{{ end}} value="{{ print "{{ ." $field.Name " }}" }}" />
                    {{ if $field.Required }}<div class="invalid-feedback" style="width: 100%;">{{ $field.Name }} is required.</div>{{ end }}
                </div>
            </div>
			{{ end }}
            <hr class="mb-4" />
            <button class="btn btn-primary btn-lg btn-block" type="submit">Save</button>
        </form>
    </div>
</div>`

func buildGenericEditorTemplate(cfg *genericEditorTemplateConfig) string {
	var b bytes.Buffer

	if err := template.Must(template.New("").Parse(genericEditorTemplateSrc)).Execute(&b, cfg); err != nil {
		panic(err)
	}

	return b.String()
}

type genericTableTemplateConfig struct {
	GetURL               string
	ExternalURL          string
	RowDataFieldName     string
	Columns              []string
	CellFields           []string
	ExcludeIDRow         bool
	IncludeLastUpdatedOn bool
	IncludeCreatedOn     bool
}

const genericTableTemplateSrc = `<table class="table table-striped">
    <thead>
        <tr>
            {{ range $i, $col := .Columns }}<th>{{ $col }}</th>
			{{ end }}
        </tr>
    </thead>
    <tbody>{{ print "{{ range $i, $x := ." .RowDataFieldName " }}" }}
        <tr>
            {{ if not .ExcludeIDRow }}<td><a href="" hx-push-url="{{ .ExternalURL }}" hx-get="{{ .GetURL }}" hx-target="#content">{{ "{{ $x.ID }}" }}</a></td>{{ end }}
			{{ range $i, $x := .CellFields }}<td>{{ print "{{ $x." $x " }}" }}</td>
			{{ end }}{{ if .IncludeLastUpdatedOn }}<td>{{ "{{ relativeTimeFromPtr $x.LastUpdatedOn }}" }}</td>{{ end }}
            {{ if .IncludeCreatedOn }}<td>{{ "{{ relativeTime $x.CreatedOn }}" }}</td>{{ end }}
        </tr>
    {{ "{{ end }}" }}</tbody>
</table>
`

func buildGenericTableTemplate(cfg *genericTableTemplateConfig) string {
	var b bytes.Buffer

	if err := template.Must(template.New("").Parse(genericTableTemplateSrc)).Execute(&b, cfg); err != nil {
		panic(err)
	}

	return b.String()
}
