// +build wasm

package html

import (
	"syscall/js"
)

// TODO: global attributes: https://developer.mozilla.org/en-US/docs/Web/HTML/Global_attributes

// Element represents a generic tag
type Element struct {
	this      js.Value
	ClassList *ClassList

	tagType  string
	ID       string
	Hidden   bool
	Disabled bool
}

// AsElement creates a new element from a js.Value
func AsElement(this js.Value) *Element {
	e := &Element{
		this: this,
	}
	return e
}

// NewElement builds a new Element
func NewElement(tag string) *Element {
	this := fetchDocument().Call("createElement", tag)
	e := &Element{
		tagType:   tag,
		this:      this,
		ClassList: NewClassList(this.Get("classList")),
	}
	return e
}

// Type returns an element's tag type (i.e. "div")
func (e *Element) Type() string {
	return e.tagType
}

// JSValue returns an element's inner js.Value
func (e *Element) JSValue() js.Value {
	return e.this
}

// SetTextContent sets the `.textContent` field
func (e *Element) SetTextContent(text string) {
	e.this.Set("textContent", text)
}

// SetInnerHTML sets the `.innerHTML` field
func (e *Element) SetInnerHTML(html string) {
	e.this.Set("innerHTML", html)
}

// SetID sets the `id` field
func (e *Element) SetID(id string) {
	e.this.Set("id", id)
}

// GetID gets the `id` field
func (e *Element) GetID() string {
	return e.this.Get("id").String()
}

// AppendChild appends a child
func (e *Element) AppendChild(child Valuer) {
	e.this.Call("appendChild", child.JSValue())
}

// RemoveChild removes a child
func (e *Element) RemoveChild(child Valuer) {
	e.this.Call("removeChild", child.JSValue())
}

// // SwapChildren can reverse the order two children appear in the list of child nodes, which can sometimes affect appearance.
// func (e *Element) SwapChildren(child1, child2 Valuer) {
// 	// https://stackoverflow.com/a/9732839
// 	child := child1.JSValue()
// 	child.Get("parentNode").Call("insertBefore", child, child2.JSValue())
// }

// OnClick registers a function to run upon click
func (e *Element) OnClick(f func(), once bool) {
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			f()
			if once {
				cb.Release()
			}
		}()
		return nil
	})
	e.this.Set("onclick", cb)
}
