package elements

import (
	"errors"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/html"
)

type Table struct {
	element *Element

	rows    []html.HTML
	headers []string
}

func (t *Table) HTML() html.HTML {
	headers := []html.HTML{}
	for _, name := range t.headers {
		headers = append(headers, html.New("th", name))
	}
	tableHeader := html.New("thead", html.New("tr", headers))

	return html.New("table", tableHeader, t.rows)
}

var ErrInadequateRowDataCount = errors.New("not enough values in the row to satisfy all headers")

func (t *Table) AddRow(values ...interface{}) error {
	if len(values) != len(t.headers) {
		return ErrInadequateRowDataCount
	}

	var cells []html.HTML
	for _, cell := range values {
		cells = append(cells, html.New("td", cell))
	}
	t.rows = append(t.rows, html.New("tr", cells))

	return nil
}

func NewTable(headerNames ...string) *Table {
	t := &Table{
		headers: headerNames,
		element: &Element{
			name:     "table",
			attr:     html.Attr{},
			children: []interface{}{},
		},
	}

	return t
}
