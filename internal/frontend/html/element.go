// +build wasm

package html

import (
	"log"
	"sync/atomic"
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

// SetStyle sets the `.style` field
func (e *Element) SetStyle(css string) {
	e.this.Set("style", css)
}

// ToHTML returns the `.outerHTML` field
func (e *Element) ToHTML() string {
	return e.this.Get("outerHTML").String()
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

// AppendChildren appends multiple children
func (e *Element) AppendChildren(children ...Valuer) {
	for _, child := range children {
		e.this.Call("appendChild", child.JSValue())
	}
}

// RemoveChild removes a child
func (e *Element) RemoveChild(child Valuer) {
	e.this.Call("removeChild", child.JSValue())
}

// OrphanChildren removes all child elements from the element
func (e *Element) OrphanChildren() {
	nodes := e.this.Get("childNodes")
	nodeCount := nodes.Length()

	log.Printf(`

	iterating over %d child nodes

	`, nodeCount)

	removedCount := 0
	for i := 0; i < nodeCount; i++ {
		node := nodes.Index(i)
		e.RemoveChild(node)
		removedCount++
	}

	log.Printf(`

	removed %d child nodes

	`, removedCount)
}

// OnClick registers a function to run upon click
func (e *Element) OnClick(f func()) {
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			f()
		}()
		return nil
	})
	e.this.Set("onclick", cb)
}

// OnClickN registers a function to run upon clicking, but only N times
func (e *Element) OnClickN(f func(), n uint64) {
	var cb js.Func
	var callCount uint64
	cb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			atomic.AddUint64(&callCount, 1)
			f()
			if callCount >= n {
				cb.Release()
			}
		}()
		return nil
	})
	e.this.Set("onclick", cb)
}
