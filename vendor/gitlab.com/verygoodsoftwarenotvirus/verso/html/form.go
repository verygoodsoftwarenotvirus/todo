// +build wasm

package html

import (
// "syscall/js"
)

// Form represents a form tag
type Form struct {
	Element

	action string
}

// NewForm builds a Button
func NewForm(action string) *Form {
	f := &Form{
		Element: *(NewElement("form")),
		action: action,
	}

	return f
}
