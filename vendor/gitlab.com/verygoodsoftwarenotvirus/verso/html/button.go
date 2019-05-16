// +build wasm

package html

import (
// "syscall/js"
)

// Button represents a Button
type Button struct {
	Element
}

// NewButton builds a Button
func NewButton(name string) *Button {
	b := &Button{
		Element: *(NewElement("button")),
	}
	b.SetTextContent(name)
	return b
}

// SetFormAction sets the formaction value of a button
func (b *Button) SetFormAction(action string) {
	b.Element.JSValue().Set("formaction", action)
}
