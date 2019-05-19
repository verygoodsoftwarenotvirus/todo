// +build wasm

package html

import (
	"syscall/js"
)

var (
	cachedJSDocument js.Value
	globalDocument   Document
)

// Document is a stand-in for a browser's `document` object
type Document interface {
	CreateElement(tagName string) *Element
	GetElementByID(id string) *Element
	QuerySelector(selectors string) *Element
	QuerySelectorAll(selectors string) []Element
	Cookie() string
}

type document struct {
	jsDocument js.Value
}

func (d *document) CreateElement(tagName string) *Element {
	this := d.jsDocument.Call("createElement", tagName)
	e := &Element{
		tagType:   tagName,
		this:      this,
		ClassList: NewClassList(this.Get("classList")),
	}
	return e
}

func (d *document) GetElementByID(id string) *Element {
	return AsElement(d.jsDocument.Call("getElementById", id))
}

func (d *document) QuerySelector(selectors string) *Element {
	return AsElement(d.jsDocument.Call("querySelector", selectors))
}

func (d *document) QuerySelectorAll(selectors string) []Element {
	val := d.jsDocument.Call("querySelectorAll", selectors)

	var out []Element
	for i := 0; i < val.Length(); i++ {
		x := val.Index(i)
		out = append(out, *AsElement(x))
	}

	return out
}

func (d *document) Cookie() string {
	return d.jsDocument.Get("cookie").String()
}

func fetchDocument() js.Value {
	if cachedJSDocument != js.Undefined() && cachedJSDocument != js.Null() {
		return cachedJSDocument
	}

	cachedJSDocument = js.Global().Get("document")
	return cachedJSDocument
}

// GetDocument returns the wrapper struct for the `document` object in a browser
func GetDocument() Document {
	if globalDocument != nil {
		return globalDocument
	}

	globalDocument = &document{
		jsDocument: fetchDocument(),
	}

	return globalDocument
}

// Body returns the js.Value for the `body` object in a browser
func Body() *Element {
	return AsElement(fetchDocument().Call("getElementsByTagName", "body").Index(0))
}

// Head returns the js.Value for the `body` object in a browser
func Head() *Element {
	return AsElement(fetchDocument().Call("getElementsByTagName", "head").Index(0))
}
