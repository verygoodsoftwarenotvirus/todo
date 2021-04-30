package elements

import (
	"errors"
	"log"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/html"
)

type Table struct {
	element *Element

	rows    []html.HTML
	headers []string
}

/*
<div class="table-responsive">
   <table class="table table-striped table-sm">
      <thead>
         <tr>
            <th>#</th>
            <th>Header</th>
            <th>Header</th>
            <th>Header</th>
            <th>Header</th>
         </tr>
      </thead>
      <tbody>
         <tr>
            <td>1,001</td>
            <td>random</td>
            <td>data</td>
            <td>placeholder</td>
            <td>text</td>
         </tr>
         <tr>
            <td>1,002</td>
            <td>placeholder</td>
            <td>irrelevant</td>
            <td>visual</td>
            <td>layout</td>
         </tr>
         <tr>
            <td>1,003</td>
            <td>data</td>
            <td>rich</td>
            <td>dashboard</td>
            <td>tabular</td>
         </tr>
      </tbody>
   </table>
</div>
*/

func (t *Table) HTML() html.HTML {
	headers := []html.HTML{}
	for _, name := range t.headers {
		headers = append(headers, html.New("th", name))
	}
	tableHeader := html.New("thead", html.New("tr", headers))

	log.Printf("returning an HTML table with %d rows", len(t.rows))

	// return html.New("table", tableHeader, t.rows)
	return html.New("table", html.Attribute(html.WithClasses("table", "table-striped")), tableHeader, t.rows)
}

var ErrInadequateRowDataCount = errors.New("not enough values in the row to satisfy all headers")

func (t *Table) AddRow(values ...interface{}) error {
	if len(values) != len(t.headers) {
		return ErrInadequateRowDataCount
	}

	log.Println("adding item to rows")

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
