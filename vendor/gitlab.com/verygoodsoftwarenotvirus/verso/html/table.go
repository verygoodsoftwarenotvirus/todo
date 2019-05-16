// +build wasm

package html

import (
	"syscall/js"
)

// TODO: the variable scope of these fields is whack

/////////////////////////////////////////////////////////////////////////// TABLE ROW

// Table represents a table element
type Table struct {
	Element

	head   *Element
	body   *Element
	footer *Element

	Headers        []*TableHeader
	Rows           []*TableRow
	FooterElements []js.Value
}

// NewTable builds a HTMLTable
func NewTable() *Table {
	t := &Table{
		Element: *(NewElement("table")),
		Headers: []*TableHeader{},
		Rows:    []*TableRow{},

		head:   NewElement("thead"),
		body:   NewElement("tbody"),
		footer: NewElement("tfoot"),
	}
	t.Element.AppendChild(t.head)
	t.Element.AppendChild(t.body)
	t.Element.AppendChild(t.footer)
	return t
}

// AddHeader adds a header to the table
func (t *Table) AddHeader(name string) *TableHeader {
	th := NewTableHeader(name)

	t.Headers = append(t.Headers, th)
	t.head.AppendChild(th)

	return th
}

// AddRawTableHeader adds a TableHeader to the table
func (t *Table) AddRawTableHeader(th *TableHeader) {
	t.Headers = append(t.Headers, th)
}

// AddRow adds a row to the table
func (t *Table) AddRow(cells map[string]string) {
	headers := []string{}
	for k := range cells {
		headers = append(headers, k)
	}
	tr := NewTableRow(headers, cells)
	t.Rows = append(t.Rows, tr)
	t.body.AppendChild(tr)
}

// AddRawTableRow adds a TableRow to the table
func (t *Table) AddRawTableRow(tr *TableRow) {
	t.Rows = append(t.Rows, tr)
}

// Render returns the table as a js.Value, with all its child html attached
func (t *Table) Render() {
	for _, header := range t.Headers {
		t.head.AppendChild(header)
	}
	for _, row := range t.Rows {
		t.body.AppendChild(row)
	}

	for _, v := range t.FooterElements {
		t.footer.JSValue().Call("appendChild", v)
	}
}

// AppendToFooter allows you to add things to the otherwise inaccessible footer
func (t *Table) AppendToFooter(value js.Value) {
	t.FooterElements = append(t.FooterElements, value)
}

/////////////////////////////////////////////////////////////////////////// TABLE ROW

// TableHeader represents a <th> tag
type TableHeader struct {
	Element
}

// NewTableHeader builds a TableHeader
func NewTableHeader(name string) *TableHeader {
	th := &TableHeader{Element: *(NewElement("th"))}
	th.SetTextContent(name)
	return th
}

/////////////////////////////////////////////////////////////////////////// TABLE ROW

// TableRow represents a <tr> tag
type TableRow struct {
	Element

	headers []string
	data    map[string]*TableCell
}

// NewTableRow builds a TableRow
func NewTableRow(headers []string, data map[string]string) *TableRow {
	tr := &TableRow{
		Element: *(NewElement("tr")),
		headers: []string{},
		data:    map[string]*TableCell{},
	}
	if headers != nil {
		tr.headers = headers
	}

	for _, col := range tr.headers {
		if cell, ok := data[col]; ok {
			tr.addCell(col, cell)
		} else {
			tr.addCell(col, "")
		}
	}
	return tr
}

// AddCell adds a cell to our table row
func (tr *TableRow) addCell(col, cell string) {
	c := NewTableCell(cell)
	tr.data[col] = c
	tr.AppendChild(c)
}

// GetCell gets a cell
func (tr *TableRow) GetCell(name string) *TableCell {
	if tc, ok := tr.data[name]; ok {
		return tc
	}
	return &TableCell{Content: ""}
}

/////////////////////////////////////////////////////////////////////////// TABLE ROW

// TableCell represents a <td> element
type TableCell struct {
	Element

	Content string
}

// NewTableCell builds a TableCell
func NewTableCell(content string) *TableCell {
	td := &TableCell{
		Element: *(NewElement("td")),
		Content: content,
	}
	td.SetTextContent(td.Content)
	return td
}
