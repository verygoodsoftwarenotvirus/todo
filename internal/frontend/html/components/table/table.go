// +build wasm

package table

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"sort"
	"syscall/js"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/frontend/html"
)

const (
	//DefaultTagName is the tag name we use to see what a given field wants to be called.
	// For instance, if I have the following struct:
	// type Person struct {
	//  	Name string `column:"banana"`
	// }
	// the column will show up as "banana" even though it is called ".Name"
	DefaultTagName = "column"
)

// PageProvider is a mechanism for fetching the next page of items from a data source
type PageProvider interface {
	LastPage() interface{}
	NextPage() interface{}
}

type column struct {
	DisplayName   string
	ReferenceName string
}

// Table represents a dynamic table that can self-sort any simple structs
type Table struct {
	id string

	HTMLTable    *html.Table
	pageProvider PageProvider

	columns   []column
	columnMap map[string]column

	rowData []map[string]string
	rows    []*html.TableRow

	sortBy  string
	sortAsc bool
}

// NewTableFromMap provides a new HTMLTable with an input of an array of maps
func NewTableFromMap(id string, input []map[string]string) *Table {
	t := html.NewTable()
	t.SetID(id)

	tbl := &Table{
		id:        id,
		HTMLTable: t,
		columns:   []column{},
		columnMap: make(map[string]column),
		rows:      []*html.TableRow{},
	}

	for _, m := range input {
		for k := range m {
			if _, ok := tbl.columnMap[k]; !ok {
				h := column{DisplayName: k, ReferenceName: k}
				tbl.columnMap[k] = h
				tbl.columns = append(tbl.columns, h)
			}
		}
	}

	for _, m := range input {
		tbl.AddRow(m)
	}
	return tbl
}

// NewTableFromStructs provides a new table given an array of simple structs
func NewTableFromStructs(id string, input interface{}) (*Table, error) {
	t := html.NewTable()
	t.SetID(id)

	tbl := &Table{
		id:        id,
		HTMLTable: t,
		columns:   []column{},
		columnMap: map[string]column{},
		rows:      []*html.TableRow{},
	}

	if err := tbl.loadData(input); err != nil {
		return nil, err
	}

	return tbl, nil
}

func (tbl *Table) loadData(input interface{}) error {
	tbl.columns = []column{}
	tbl.columnMap = map[string]column{}
	tbl.rows = []*html.TableRow{}

	i := reflect.TypeOf(input)

	if i.Kind() == reflect.Slice {
		tbl.determineColumns(i.Elem())
		val := reflect.ValueOf(input)

		for i := 0; i < val.Len(); i++ {
			newRow := map[string]string{}

			for _, column := range tbl.columns {
				f := val.Index(i).FieldByName(column.ReferenceName)
				x := f.Interface()
				v := reflect.ValueOf(x)

				switch v.Kind() {
				case reflect.Ptr:
					if !v.IsNil() {
						newRow[column.DisplayName] = fmt.Sprintf("%v", reflect.Indirect(f).Interface())
					} else {
						newRow[column.DisplayName] = ""
					}
				case reflect.Struct:
					continue
				default:
					newRow[column.DisplayName] = fmt.Sprintf("%v", x)
				}
			}

			tbl.AddRow(newRow)
		}
	} else {
		return errors.New("type provided must be slice")
	}
	return nil
}

func (t *Table) determineColumns(typ reflect.Type) {
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)

		var fn = f.Name
		if v, ok := f.Tag.Lookup(DefaultTagName); ok {
			fn = v
		}

		col := column{
			DisplayName:   fn,
			ReferenceName: f.Name,
		}

		t.AddHeader(col)
		if t.sortBy == "" {
			t.sortBy = f.Name
		}
	}
}

func (t Table) Len() int {
	return len(t.rows)
}

func (t Table) Swap(i, j int) {
	t.rows[i], t.rows[j] = t.rows[j], t.rows[i]
	t.rowData[i], t.rowData[j] = t.rowData[j], t.rowData[i]
}

func (t Table) Less(i, j int) bool {
	sb := t.sortBy
	if sb == "" {
		for _, column := range t.columns {
			sb = column.DisplayName
			break
		}
	}

	if !t.sortAsc {
		return t.rows[i].GetCell(sb).Content > t.rows[j].GetCell(sb).Content
	}

	return t.rows[i].GetCell(sb).Content < t.rows[j].GetCell(sb).Content
}

// SetPageProvider sets the page provider
func (t *Table) SetPageProvider(input PageProvider) {
	t.pageProvider = input
}

// AddHeader adds a column element (column) to the table
func (t *Table) AddHeader(h column) {
	if _, exists := t.columnMap[h.DisplayName]; !exists {
		t.columns = append(t.columns, h)
		t.columnMap[h.DisplayName] = h

		column := t.HTMLTable.AddHeader(h.DisplayName)
		column.SetTextContent(h.DisplayName)
		column.OnClick(t.columnSort(h))
	}
}

func (t *Table) columnSort(h column) func() {
	return func() {
		if t.sortBy != h.DisplayName {
			t.sortBy = h.DisplayName
		} else {
			t.sortAsc = !t.sortAsc
		}
		sort.Sort(t)

		t.Redraw()
	}
}

// AddRow adds a row of data to the table
func (t *Table) AddRow(data map[string]string) {
	t.rowData = append(t.rowData, data)
	t.rows = append(t.rows, t.buildRow(data))

	var cols []string
	for _, c := range t.columns {
		cols = append(cols, c.DisplayName)
	}

	t.HTMLTable.AddRow(cols, data)
}

func (t *Table) buildRow(data map[string]string) *html.TableRow {
	var (
		cols     []string
		cellData []string
	)

	for _, col := range t.columns {
		cellData = append(cellData, data[col.DisplayName])
		cols = append(cols, col.DisplayName)
	}

	return html.NewTableRow(cols, data)
}

// Redraw redraws the table with the current contents of the table
func (t *Table) Redraw() {
	body := html.Body()

	oldTable := html.GetDocument().GetElementByID(t.id)
	oldTable.ParentElement().RemoveChild(oldTable)

	t.Render()
	body.AppendChild(t.HTMLTable)
	t.HTMLTable.SetID(t.id)
}

// Render creates our outer table shell element so we can inject our column and row html in afterwards
func (t *Table) Render() {
	t2 := html.NewTable()

	var cols []string
	for _, column := range t.columns {
		cols = append(cols, column.DisplayName)
		th := t2.AddHeader(column.DisplayName)
		th.OnClick(t.columnSort(column))
		t2.AddRawTableHeader(th)
	}

	for _, tr := range t.rowData {
		t2.AddRow(cols, tr)
	}

	if t.pageProvider != nil {
		prevButton := html.NewButton("<")
		prevButton.OnClick(func() {
			data := t.pageProvider.LastPage()
			if err := t.loadData(data); err != nil {
				log.Println("error loading data")
			}
			t.Redraw()
		})

		nextButton := html.NewButton(">")
		nextButton.OnClick(func() {
			data := t.pageProvider.NextPage()
			if err := t.loadData(data); err != nil {
				log.Println("error loading data")
			}
			t.Redraw()
		})

		t2.AppendToFooter(prevButton.JSValue())
		t2.AppendToFooter(nextButton.JSValue())
	}

	t2.Render()
	oldClassList := t.HTMLTable.ClassList.String()
	if t.HTMLTable.ClassList != nil && oldClassList != "" {
		t2.ClassList.Add(oldClassList)
	}
	t.HTMLTable = t2
}

// JSValue retuns the inner value of the table
func (t *Table) JSValue() js.Value {
	return t.HTMLTable.JSValue()
}
